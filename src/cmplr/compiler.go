// © Knug Industries 2009 all rights reserved
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package compiler

import (
    "os"
    "fmt"
    "log"
    "exec"
    "strings"
    "regexp"
    "path/filepath"
    "utilz/walker"
    "utilz/stringset"
    "utilz/handy"
    "utilz/say"
    "utilz/global"
    "cmplr/dag"
)


var includes []string
var srcroot string
var libroot string
var pathLinker string
var pathCompiler string
var suffix string


func Init(srcdir, arch string, include []string) {

    srcroot = srcdir
    includes = include

    if global.GetString("-lib") != "" {
        libroot = global.GetString("-lib")
    } else {
        libroot = srcroot
    }

    switch global.GetString("-backend") {
    case "gcc", "gccgo":
        gcc()
    case "gc":
        gc(arch)
    case "express":
        express()
    default:
        log.Fatalf("[ERROR] '%s' unknown backend\n",
            global.GetString("-backend"))
    }
}

func express() {

    var err os.Error

    pathCompiler, err = exec.LookPath("vmgc")

    if err != nil {
        log.Fatalf("[ERROR] could not find compiler: %s\n", pathCompiler)
    }

    pathLinker, err = exec.LookPath("vmld")

    if err != nil {
        log.Fatalf("[ERROR] could not find linker: %s\n", pathLinker)
    }

    suffix = ".vmo"
}

func gc(arch string) {

    var A string // a:architecture
    var err os.Error

    if arch == "" {
        A = os.Getenv("GOARCH")
    } else {
        A = arch
    }

    var S, C, L string // S:suffix, C:compiler, L:linker

    switch A {
    case "arm":
        S = ".5"
        C = "5g"
        L = "5l"
    case "amd64":
        S = ".6"
        C = "6g"
        L = "6l"
    case "386":
        S = ".8"
        C = "8g"
        L = "8l"
    default:
        log.Fatalf("[ERROR] unknown architecture: %s\n", A)
    }

    pathCompiler, err = exec.LookPath(C)

    if err != nil {
        log.Fatalf("[ERROR] could not find compiler: %s\n", C)
    }

    pathLinker, err = exec.LookPath(L)

    if err != nil {
        log.Fatalf("[ERROR] could not find linker: %s\n", L)
    }

    suffix = S

}

func gcc() {

    var err os.Error

    pathCompiler, err = exec.LookPath("gccgo")

    if err != nil {
        log.Fatalf("[ERROR] could not find compiler: %s\n", err)
    }

    pathLinker = pathCompiler

    suffix = ".o"
}


func CreateArgv(pkgs []*dag.Package) {

    var argv []string

    includeLen := len(includes)

    for y := 0; y < len(pkgs); y++ {

        argv = make([]string, 0)
        argv = append(argv, pathCompiler)
        argv = append(argv, "-I")
        argv = append(argv, libroot)
        for y := 0; y < includeLen; y++ {
            argv = append(argv, "-I")
            argv = append(argv, includes[y])
        }

        switch global.GetString("-backend") {
        case "gcc", "gccgo":
            argv = append(argv, "-c")
        }

        argv = append(argv, "-o")
        argv = append(argv, filepath.Join(libroot, pkgs[y].Name)+suffix)

        for z := 0; z < len(pkgs[y].Files); z++ {
            argv = append(argv, pkgs[y].Files[z])
        }

        pkgs[y].Argv = argv
    }
}

func CreateLibArgv(pkgs []*dag.Package) {

    ss := stringset.New()
    for i := range pkgs {
        if len(pkgs[i].Name) > len(pkgs[i].ShortName) {
            ss.Add(pkgs[i].Name[:(len(pkgs[i].Name) - len(pkgs[i].ShortName))])
        }
    }
    slice := ss.Slice()
    for i := 0; i < len(slice); i++ {
        slice[i] = filepath.Join(libroot, slice[i])
        handy.DirOrMkdir(slice[i])
    }

    CreateArgv(pkgs)

}

func SerialCompile(pkgs []*dag.Package) {

    var oldPkgFound bool = false

    for y := 0; y < len(pkgs); y++ {

        if global.GetBool("-dryrun") {
            fmt.Printf("%s || exit 1\n", strings.Join(pkgs[y].Argv, " "))
        } else {
            if oldPkgFound || !pkgs[y].UpToDate() {
                say.Println("compiling:", pkgs[y].Name)
                handy.StdExecve(pkgs[y].Argv, true)
                oldPkgFound = true
            } else {
                say.Println("up 2 date:", pkgs[y].Name)
            }
        }
    }
}

func ParallelCompile(pkgs []*dag.Package) {

    var localDeps *stringset.StringSet
    var compiledDeps *stringset.StringSet
    var y, z, count int
    var parallel []*dag.Package
    var oldPkgFound bool = false
    var zeroFirst []*dag.Package

    localDeps = stringset.New()
    compiledDeps = stringset.New()

    for y = 0; y < len(pkgs); y++ {
        localDeps.Add(pkgs[y].Name)
        pkgs[y].ResetIndegree()
    }

    zeroFirst = make([]*dag.Package, len(pkgs))

    for y = 0; y < len(pkgs); y++ {
        if pkgs[y].Indegree == 0 {
            zeroFirst[count] = pkgs[y]
            count++
        }
    }

    for y = 0; y < len(pkgs); y++ {
        if pkgs[y].Indegree > 0 {
            zeroFirst[count] = pkgs[y]
            count++
        }
    }

    parallel = make([]*dag.Package, 0)

    for y = 0; y < len(zeroFirst); {

        if !zeroFirst[y].Ready(localDeps, compiledDeps) {

            oldPkgFound = compileMultipe(parallel, oldPkgFound)

            for z = 0; z < len(parallel); z++ {
                compiledDeps.Add(parallel[z].Name)
            }

            parallel = make([]*dag.Package, 0)

        } else {
            parallel = append(parallel, zeroFirst[y])
            y++
        }
    }

    if len(parallel) > 0 {
        _ = compileMultipe(parallel, oldPkgFound)
    }

}

func compileMultipe(pkgs []*dag.Package, oldPkgFound bool) bool {

    var ok bool
    var max int = len(pkgs)
    var trouble bool = false

    if max == 0 {
        log.Fatal("[ERROR] trying to compile 0 packages in parallel\n")
    }

    if max == 1 {
        if oldPkgFound || !pkgs[0].UpToDate() {
            say.Println("compiling:", pkgs[0].Name)
            handy.StdExecve(pkgs[0].Argv, true)
            oldPkgFound = true
        } else {
            say.Println("up 2 date:", pkgs[0].Name)
        }
    } else {

        ch := make(chan bool, max)

        for y := 0; y < max; y++ {
            if oldPkgFound || !pkgs[y].UpToDate() {
                say.Println("compiling:", pkgs[y].Name)
                oldPkgFound = true
                go gCompile(pkgs[y].Argv, ch)
            } else {
                say.Println("up 2 date:", pkgs[y].Name)
                ch <- true
            }
        }

        // drain channel (make sure all jobs are finished)
        for z := 0; z < max; z++ {
            ok = <-ch
            if !ok {
                trouble = true
            }
        }
    }

    if trouble {
        log.Fatal("[ERROR] failed batch compile job\n")
    }

    return oldPkgFound
}

func gCompile(argv []string, c chan bool) {
    ok := handy.StdExecve(argv, false) // don't exit on error
    c <- ok
}

// for removal of temoprary packages created for testing and so on..
func DeletePackages(pkgs []*dag.Package) bool {

    var ok = true
    var e os.Error

    for i := 0; i < len(pkgs); i++ {

        for y := 0; y < len(pkgs[i].Files); y++ {
            e = os.Remove(pkgs[i].Files[y])
            if e != nil {
                ok = false
                log.Printf("[ERROR] %s\n", e)
            }
        }
        if !global.GetBool("-dryrun") {
            pcompile := filepath.Join(libroot, pkgs[i].Name) + suffix
            e = os.Remove(pcompile)
            if e != nil {
                ok = false
                log.Printf("[ERROR] %s\n", e)
            }
        }
    }

    return ok
}


func ForkLink(output string, pkgs []*dag.Package, extra []*dag.Package) {

    var mainPKG *dag.Package

    gotMain := make([]*dag.Package, 0)

    for i := 0; i < len(pkgs); i++ {
        if pkgs[i].ShortName == "main" {
            gotMain = append(gotMain, pkgs[i])
        }
    }

    if len(gotMain) == 0 {
        log.Fatal("[ERROR] (linking) no main package found\n")
    }

    if len(gotMain) > 1 {
        choice := mainChoice(gotMain)
        mainPKG = gotMain[choice]
    } else {
        mainPKG = gotMain[0]
    }

    compiled := filepath.Join(libroot, mainPKG.Name) + suffix

    argv := make([]string, 0)
    argv = append(argv, pathLinker)

    switch global.GetString("-backend") {
    case "gc", "express":
        argv = append(argv, "-L")
        argv = append(argv, libroot)
    }

    argv = append(argv, "-o")
    argv = append(argv, output)

    // gcc get's this no matter what...
    if global.GetString("-backend") == "gcc" ||
        global.GetString("-backend") == "gccgo" {
        argv = append(argv, "-static")
    } else if global.GetBool("-static") {
        argv = append(argv, "-d")
    }

    switch global.GetString("-backend") {
    case "gccgo", "gcc":
        walker.IncludeFile = func(s string) bool {
            return strings.HasSuffix(s, ".o")
        }
        walker.IncludeDir = func(s string) bool { return true }

        for y := 0; y < len(includes); y++ {
            argv = append(argv, walker.PathWalk(includes[y])...)
        }
    case "gc", "express":
        for y := 0; y < len(includes); y++ {
            argv = append(argv, "-L")
            argv = append(argv, includes[y])
        }
    }

    argv = append(argv, compiled)

    if global.GetString("-backend") == "gcc" ||
        global.GetString("-backend") == "gccgo" {

        ss := stringset.New()

        if len(extra) > 0 {
            for j := 0; j < len(extra); j++ {
                // main package untestable using GCC
                if extra[j].ShortName != "main" {
                    ss.Add(filepath.Join(libroot, extra[j].Name) + suffix)
                }
            }
        } else {
            for k := 0; k < len(pkgs); k++ {
                ss.Add(filepath.Join(libroot, pkgs[k].Name) + suffix)
            }
            ss.Remove(compiled)
        }

        if ss.Len() > 0 {
            argv = append(argv, ss.Slice()...)
        }
    }

    if global.GetBool("-dryrun") {
        fmt.Printf("%s || exit 1\n", strings.Join(argv, " "))
    } else {
        say.Println("linking  :", output)
        handy.StdExecve(argv, true)
    }
}

func mainChoice(pkgs []*dag.Package) int {

    var cnt int
    var choice int

    for i := 0; i < len(pkgs); i++ {
        ok, _ := regexp.MatchString(global.GetString("-main"), pkgs[i].Name)
        if ok {
            cnt++
            choice = i
        }
    }

    if cnt == 1 {
        return choice
    }

    fmt.Println("\n More than one main package found\n")

    for i := 0; i < len(pkgs); i++ {
        fmt.Printf(" type %2d  for: %s\n", i, pkgs[i].Name)
    }

    fmt.Printf("\n type your choice: ")

    n, e := fmt.Scanf("%d", &choice)

    if e != nil {
        log.Fatalf("%s\n", e)
    }
    if n != 1 {
        log.Fatal("failed to read input\n")
    }

    if choice >= len(pkgs) || choice < 0 {
        log.Fatalf(" bad choice: %d\n", choice)
    }

    fmt.Printf(" chosen main-package: %s\n\n", pkgs[choice].Name)

    return choice
}


func CreateTestArgv() []string {

    pwd, e := os.Getwd()

    if e != nil {
        log.Fatal("[ERROR] could not locate working directory\n")
    }

    argv := make([]string, 0)

    if global.GetString("-backend") == "express" {
        vmrun, e := exec.LookPath("vmrun")
        if e != nil {
            log.Fatalf("[ERROR] %s\n",e)
        }
        argv = append(argv, vmrun)
    }

    argv = append(argv,filepath.Join(pwd, global.GetString("-test-bin")))

    if global.GetString("-bench") != "" {
        argv = append(argv, "-test.bench")
        argv = append(argv, global.GetString("-bench"))
    }
    if global.GetString("-match") != "" {
        argv = append(argv, "-test.run")
        argv = append(argv, global.GetString("-match"))
    }
    if global.GetBool("-verbose") {
        argv = append(argv, "-test.v")
    }
    return argv
}

func Remove865o(dir string, alsoDir bool) {
    // override IncludeFile to make walker pick up .[865] .o .vmo
    walker.IncludeFile = func(s string) bool {
        return strings.HasSuffix(s, ".8") ||
               strings.HasSuffix(s, ".6") ||
               strings.HasSuffix(s, ".5") ||
               strings.HasSuffix(s, ".o") ||
               strings.HasSuffix(s, ".vmo")
    }

    handy.DirOrExit(dir)

    compiled := walker.PathWalk(filepath.Clean(dir))

    for i := 0; i < len(compiled); i++ {

        shortName := compiled[i]
        pwd, e    := os.Getwd()
        if e == nil {
            if strings.HasPrefix(compiled[i], pwd){
                shortName = shortName[len(pwd)+1:]
            }
        }

        if !global.GetBool("-dryrun") {

            e := os.Remove(compiled[i])
            if e != nil {
                log.Printf("[ERROR] could not delete file: %s\n", compiled[i])
            } else {
                say.Printf("rm: %s\n", shortName)
            }

        } else {
            fmt.Printf("[dryrun] rm: %s\n", shortName)
        }
    }

    if alsoDir {
        // remove entire dir if empty after objects are deleted
        walker.IncludeFile = func(s string) bool { return true }
        walker.IncludeDir = func(s string) bool { return true }
        if len(walker.PathWalk(dir)) == 0 {
            if global.GetBool("-dryrun") {
                fmt.Printf("[dryrun] rm: %s\n", dir)
            } else {
                say.Printf("rm: %s\n", dir)
                e := os.RemoveAll(dir)
                if e != nil {
                    log.Fatalf("[ERROR] %s\n", e)
                }
            }
        }
    }
}


func FormatFiles(files []string) {

    var i int
    var argv []string
    var tabWidth string = "-tabwidth=4"
    var useTabs string = "-tabindent=false"
    var rewRule string = global.GetString("-rew-rule")
    var fmtexec string
    var err os.Error

    fmtexec, err = exec.LookPath("gofmt")

    if err != nil {
        log.Fatal("[ERROR] could not find 'gofmt' in $PATH")
    }

    if global.GetString("-tabwidth") != "" {
        tabWidth = "-tabwidth=" + global.GetString("-tabwidth")
    }
    if global.GetBool("-tab") {
        useTabs = "-tabindent=true"
    }

    argv = make([]string, 0)

    argv = append(argv, fmtexec)
    argv = append(argv, "-w=true")
    argv = append(argv, tabWidth)
    argv = append(argv, useTabs)

    if rewRule != "" {
        argv = append(argv, fmt.Sprintf("-r='%s'", rewRule))
    }

    argv = append(argv, "") // dummy
    i = len(argv) - 1

    for y := 0; y < len(files); y++ {
        argv[i] = files[y]
        if !global.GetBool("-dryrun") {
            say.Printf("gofmt: %s\n", files[y])
            _ = handy.StdExecve(argv, true)
        } else {
            fmt.Printf(" %s\n", strings.Join(argv, " "))
        }
    }
}
