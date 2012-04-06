//  Copyright © 2009 bjarneh
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package compiler

import (
    "cmplr/dag"
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "regexp"
    "strings"
    "utilz/global"
    "utilz/handy"
    "utilz/say"
    "utilz/stringset"
    "utilz/walker"
)

var includes []string
var srcroot string
var libroot string
var pathLinker string
var pathCompiler string
var suffix string

func Init(srcdir string, include []string) {

    srcroot = srcdir
    includes = include

    if global.GetString("-lib") != "" {
        libroot = global.GetString("-lib")
    } else {
        libroot = srcroot
    }

    InitBackend()
}

func InitBackend(){

    switch global.GetString("-backend") {
    case "gcc", "gccgo":
        gcc()
    case "gc":
        gc()
    case "express":
        express()
    default:
        log.Fatalf("[ERROR] '%s' unknown backend\n",
            global.GetString("-backend"))
    }
}

func express() {

    var err error

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

//TODO fix this mess
func gc() {

    var (
        A   string // A:architecture
        S   string // S:suffix
        C   string // C:compiler
        L   string // L:linker
        R   string // R:goroot
        O   string // O:goOS
    )

    var err error

    A = handy.GOARCH()
    R = handy.GOROOT()
    O = handy.GOOS()

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

    path_C := filepath.Join(R, "pkg", "tool", (O + "_" + A), C)

    pathCompiler, err = exec.LookPath(path_C)

    if err != nil {
        log.Fatalf("[ERROR] could not find compiler: %s\n", C)
    }

    path_L := filepath.Join(R, "pkg", "tool", (O + "_" + A), L)

    pathLinker, err = exec.LookPath(path_L)

    if err != nil {
        log.Fatalf("[ERROR] could not find linker: %s\n", L)
    }

    suffix = S

}

func gcc() {

    var err error

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

        golibs := handy.GoPathImports(global.GetString("-backend"))
        for j := 0; j < len(golibs); j++ {
            argv = append(argv, "-I")
            argv = append(argv, golibs[j])
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

func Dryrun(pkgs []*dag.Package) {
    binary := filepath.Base(pathCompiler)
    for y := 0; y < len(pkgs); y++ {
        args := strings.Join(pkgs[y].Argv[1:], " ")
        fmt.Printf("%s %s || exit 1\n", binary, args)
    }
}

// this is faster than ParallelCompile i.e. (the old version).
// after release.r60.1 this is used for all compile jobs
func Compile(pkgs []*dag.Package) bool {
    // set indegree, i.e. how many jobs to wait for
    for y := 0; y < len(pkgs); y++ {
        pkgs[y].ResetIndegree()
    }
    // init waitgroup to number of jobs calculated above
    for y := 0; y < len(pkgs); y++ {
        pkgs[y].InitWaitGroup()
    }
    // start up one go-routine for each package
    ch := make(chan int)
    for y := 0; y < len(pkgs); y++ {
        go pkgs[y].Compile(ch)
    }
    // make sure all jobs finished, i.e. drain channel
    for y := 0; y < len(pkgs); y++ {
        _ = <-ch
    }
    close(ch)
    return !dag.OldPkgYet()
}

// for removal of temoprary packages created for testing and so on..
func DeletePackages(pkgs []*dag.Package) bool {

    var ok = true

    for i := 0; i < len(pkgs); i++ {

        for y := 0; y < len(pkgs[i].Files); y++ {
            handy.Delete(pkgs[i].Files[y], false)
        }
        if !global.GetBool("-dryrun") {
            pcompile := filepath.Join(libroot, pkgs[i].Name) + suffix
            ok = handy.Delete(pcompile, false)
        }
    }

    return ok
}

// Recompile packages that contain test files (*_test.go), 
// i.e. test-code should not be part of the packages after compilation.
func ReCompile(pkgs []*dag.Package) bool {

    var doRecompile bool

    for i := 0; i < len(pkgs); i++ {
        if pkgs[i].HasTestAndInit() {
            doRecompile = true
        }
    }

    return doRecompile

}

//TODO rewrite the whole link stuff, make i run in parallel
func ForkLinkAll(pkgs []*dag.Package, up2date bool) {

    mainPkgs := make([]*dag.Package, 0)

    for i := 0; i < len(pkgs); i++ {
        if pkgs[i].ShortName == "main" {
            mainPkgs = append(mainPkgs, pkgs[i])
        }
    }

    if len(mainPkgs) == 0 {
        log.Fatal("[ERROR] (linking) no main package found\n")
    }

    handy.DirOrMkdir("bin")

    for i := 0; i < len(mainPkgs); i++ {
        toks := strings.Split(mainPkgs[i].Name, "/")
        // do this for main packages which are placed in directories
        // if toks < 2 this cannot be true, i.e. then the main package
        // lives under the src-root and cannot be filtered
        if len(toks) >= 2 {
            nameOfBinary := toks[len(toks)-2]
            pathToBinary := filepath.Join("bin", nameOfBinary)
            global.SetString("-main", nameOfBinary)
            ForkLink(pathToBinary, pkgs, nil, up2date)
        }
    }
}

func ForkLink(output string, pkgs []*dag.Package, extra []*dag.Package, up2date bool) {

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

    if up2date && !global.GetBool("-dryrun") && handy.IsFile(output) {
        if handy.ModifyTimestamp(compiled) < handy.ModifyTimestamp(output) {
            say.Printf("up 2 date: %s\n", output)
            return
        }
    }

    argv := make([]string, 0)
    argv = append(argv, pathLinker)

    switch global.GetString("-backend") {
    case "gc", "express":
        argv = append(argv, "-L")
        argv = append(argv, libroot)
        if global.GetString("-backend") == "gc" {
            golibs := handy.GoPathImports("gc")
            for j := 0; j < len(golibs); j++ {
                argv = append(argv, "-L")
                argv = append(argv, golibs[j])
            }
            if global.GetBool("-strip") {
                argv = append(argv, "-s")
            }
        }
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
        linker := filepath.Base(pathLinker)
        fmt.Printf("%s %s || exit 1\n", linker, strings.Join(argv[1:], " "))
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
            log.Fatalf("[ERROR] %s\n", e)
        }
        argv = append(argv, vmrun)
    }

    argv = append(argv, filepath.Join(pwd, global.GetString("-test-bin")))

    if global.GetString("-bench") != "" {
        argv = append(argv, "-test.bench")
        argv = append(argv, global.GetString("-bench"))
    } else if global.GetString("-test.bench") != "" {
        argv = append(argv, "-test.bench")
        argv = append(argv, global.GetString("-test.bench"))
    }

    if global.GetString("-match") != "" {
        argv = append(argv, "-test.run")
        argv = append(argv, global.GetString("-match"))
    } else if global.GetString("-test.run") != "" {
        argv = append(argv, "-test.run")
        argv = append(argv, global.GetString("-test.run"))
    }

    if global.GetString("-test.timeout") != "" {
        argv = append(argv, "-test.timeout")
        argv = append(argv, global.GetString("-test.timeout"))
    }

    if global.GetString("-test.benchtime") != "" {
        argv = append(argv, "-test.benchtime")
        argv = append(argv, global.GetString("-test.benchtime"))
    }

    if global.GetString("-test.parallel") != "" {
        argv = append(argv, "-test.parallel")
        argv = append(argv, global.GetString("-test.parallel"))
    }

    if global.GetString("-test.cpu") != "" {
        argv = append(argv, "-test.cpu")
        argv = append(argv, global.GetString("-test.cpu"))
    }

    if global.GetString("-test.cpuprofile") != "" {
        argv = append(argv, "-test.cpuprofile")
        argv = append(argv, global.GetString("-test.cpuprofile"))
    }

    if global.GetString("-test.memprofile") != "" {
        argv = append(argv, "-test.memprofile")
        argv = append(argv, global.GetString("-test.memprofile"))
    }

    if global.GetString("-test.memprofilerate") != "" {
        argv = append(argv, "-test.memprofilerate")
        argv = append(argv, global.GetString("-test.memprofilerate"))
    }

    if global.GetBool("-verbose") || global.GetBool("-test.v") {
        argv = append(argv, "-test.v")
    }

    if global.GetBool("-test.short") {
        argv = append(argv, "-test.short")
    }

    return argv
}

func FormatFiles(files []string) {

    var i int
    var argv []string
    var tabWidth string = "-tabwidth=4"
    var useTabs string = "-tabs=false"
    var rewRule string = global.GetString("-rew-rule")
    var fmtexec string
    var err error

    fmtexec, err = exec.LookPath("gofmt")

    if err != nil {
        log.Fatal("[ERROR] could not find 'gofmt' in $PATH")
    }

    if global.GetString("-tabwidth") != "" {
        tabWidth = "-tabwidth=" + global.GetString("-tabwidth")
    }
    if global.GetBool("-tab") {
        useTabs = "-tabs=true"
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
        if global.GetBool("-dryrun") {
            argv[0] = filepath.Base(argv[0])
            fmt.Printf(" %s\n", strings.Join(argv, " "))
        } else {
            say.Printf("gofmt: %s\n", files[y])
            _ = handy.StdExecve(argv, true)
        }
    }
}

func DeleteObjects(dir string, pkgs []*dag.Package) {

    var stub, tmp string

    suffixes := []string{".8", ".6", ".5", ".o", ".vmo"}

    libdir := global.GetString("-lib")

    if libdir != "" {
        dir = libdir
    }

    for i := 0; i < len(pkgs); i++ {
        stub = filepath.Join(dir, pkgs[i].Name)
        for j := 0; j < len(suffixes); j++ {
            tmp = stub + suffixes[j]
            if handy.IsFile(tmp) {
                if global.GetBool("-dryrun") {
                    say.Printf("[dryrun] rm: %s\n", tmp)
                } else {
                    say.Printf("rm: %s\n", tmp)
                    handy.Delete(tmp, false)
                }
            }
        }
    }

    // remove entire dir if empty after objects are deleted.
    // only do this if -lib is present, there is no reason to
    // do this (extra treewalk) if objects are in src directory
    if libdir != "" && handy.IsDir(dir) {
        walker.IncludeFile = func(s string) bool { return true }
        walker.IncludeDir = func(s string) bool { return true }
        if len(walker.PathWalk(dir)) == 0 {
            if global.GetBool("-dryrun") {
                fmt.Printf("[dryrun] rm: %s\n", dir)
            } else {
                say.Printf("rm: %s\n", dir)
                handy.RmRf(dir, true) // die on error
            }
        }
    }
}
