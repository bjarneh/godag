// © Knug Industries 2009 all rights reserved
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package compiler

import (
    "os"
    "container/vector"
    "fmt"
    "path"
    "log"
    "exec"
    "strings"
    "utilz/walker"
    "utilz/stringset"
    "utilz/handy"
    "utilz/global"
    "cmplr/dag"
)


type Compiler struct {
    root, arch, suffix string
    executable, linker string
    dryrun             bool
    includes           []string
}

func New(root string, include []string) *Compiler {
    c := new(Compiler)
    c.root = root
    c.dryrun = global.GetBool("-dryrun")
    c.includes = include
    c.archDependantInfo(global.GetString("-arch"))
    return c
}

func (c *Compiler) archDependantInfo(arch string) {

    var A string // a:architecture

    var pathCompiler, pathLinker string
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
        log.Exitf("[ERROR] unknown architecture: %s\n", A)
    }

    pathCompiler, err = exec.LookPath(C)

    if err != nil {
        log.Exitf("[ERROR] could not find compiler: %s\n", C)
    }

    pathLinker, err = exec.LookPath(L)

    if err != nil {
        log.Exitf("[ERROR] could not find linker: %s\n", L)
    }

    c.arch = A
    c.executable = pathCompiler
    c.linker = pathLinker
    c.suffix = S

}


func (c *Compiler) CreateArgv(pkgs *vector.Vector) {

    var argv []string

    includeLen := c.extraPkgIncludes()

    for y := 0; y < pkgs.Len(); y++ {
        pkg, _ := pkgs.At(y).(*dag.Package) //safe cast, only Packages there

        argv = make([]string, 5+pkg.Files.Len()+(includeLen*2))
        i := 0
        argv[i] = c.executable
        i++
        argv[i] = "-I"
        i++
        argv[i] = c.root
        i++
        if includeLen > 0 {
            for y := 0; y < includeLen; y++ {
                argv[i] = "-I"
                i++
                argv[i] = c.includes[y]
                i++
            }
        }
        argv[i] = "-o"
        i++
        argv[i] = path.Join(c.root, pkg.Name) + c.suffix
        i++

        for z := 0; z < pkg.Files.Len(); z++ {
            argv[i] = pkg.Files.At(z)
            i++
        }

        pkg.Argv = argv
    }
}

func (c *Compiler) SerialCompile(pkgs *vector.Vector) {

    var oldPkgFound bool = false

    for y := 0; y < pkgs.Len(); y++ {
        pkg, _ := pkgs.At(y).(*dag.Package) //safe cast, only Packages there

        if c.dryrun {
            dryRun(pkg.Argv)
        } else {
            if oldPkgFound || !pkg.UpToDate() {
                fmt.Println("compiling:", pkg.Name)
                handy.StdExecve(pkg.Argv, true)
                oldPkgFound = true
            } else {
                fmt.Println("up 2 date:", pkg.Name)
            }
        }
    }
}

func (c *Compiler) ParallelCompile(pkgs *vector.Vector) {

    var localDeps *stringset.StringSet
    var compiledDeps *stringset.StringSet
    var pkg, cpkg *dag.Package
    var y, z int
    var parallel *vector.Vector
    var oldPkgFound bool = false

    localDeps = stringset.New()
    compiledDeps = stringset.New()

    for y = 0; y < pkgs.Len(); y++ {
        pkg, _ = pkgs.At(y).(*dag.Package)
        localDeps.Add(pkg.Name)
    }

    parallel = new(vector.Vector)

    for y = 0; y < pkgs.Len(); {

        pkg, _ = pkgs.At(y).(*dag.Package)

        if !pkg.Ready(localDeps, compiledDeps) {

            oldPkgFound = c.compileMultipe(parallel, oldPkgFound)

            for z = 0; z < parallel.Len(); z++ {
                cpkg, _ = parallel.At(z).(*dag.Package)
                compiledDeps.Add(cpkg.Name)
            }

            parallel = new(vector.Vector)

        } else {
            parallel.Push(pkg)
            y++
        }
    }

    if parallel.Len() > 0 {
        oldPkgFound = c.compileMultipe(parallel, oldPkgFound)
    }

}

func (c *Compiler) compileMultipe(pkgs *vector.Vector, oldPkgFound bool) bool {

    var ok bool
    var max int = pkgs.Len()
    var pkg *dag.Package
    var trouble bool = false

    if max == 0 {
        log.Exit("[ERROR] trying to compile 0 packages in parallel\n")
    }

    if max == 1 {
        pkg, _ = pkgs.At(0).(*dag.Package)
        if oldPkgFound || !pkg.UpToDate() {
            fmt.Println("compiling:", pkg.Name)
            handy.StdExecve(pkg.Argv, true)
            oldPkgFound = true
        } else {
            fmt.Println("up 2 date:", pkg.Name)
        }
    } else {

        ch := make(chan bool, pkgs.Len())

        for y := 0; y < max; y++ {
            pkg, _ := pkgs.At(y).(*dag.Package)
            if oldPkgFound || !pkg.UpToDate() {
                fmt.Println("compiling:", pkg.Name)
                oldPkgFound = true
                go gCompile(pkg.Argv, ch)
            } else {
                fmt.Println("up 2 date:", pkg.Name)
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
        log.Exit("[ERROR] failed batch compile job\n")
    }

    return oldPkgFound
}

func gCompile(argv []string, c chan bool) {
    ok := handy.StdExecve(argv, false) // don't exit on error
    c <- ok
}

// for removal of temoprary packages created for testing and so on..
func (c *Compiler) DeletePackages(pkgs *vector.Vector) bool {

    var ok = true
    var e os.Error

    for i := 0; i < pkgs.Len(); i++ {
        pkg, _ := pkgs.At(i).(*dag.Package) //safe cast, only Packages there

        for y := 0; y < pkg.Files.Len(); y++ {
            e = os.Remove(pkg.Files.At(y))
            if e != nil {
                ok = false
                log.Printf("[ERROR] %s\n", e)
            }
        }
        if !c.dryrun {
            pcompile := path.Join(c.root, pkg.Name) + c.suffix
            e = os.Remove(pcompile)
            if e != nil {
                ok = false
                log.Printf("[ERROR] %s\n", e)
            }
        }
    }

    return ok
}

func (c *Compiler) ForkLink(pkgs *vector.Vector, output string) {

    var mainPKG *dag.Package

    gotMain := new(vector.Vector)

    for i := 0; i < pkgs.Len(); i++ {
        pk, _ := pkgs.At(i).(*dag.Package)
        if pk.ShortName == "main" {
            gotMain.Push(pk)
        }
    }

    if gotMain.Len() == 0 {
        log.Exit("[ERROR] (linking) no main package found\n")
    }

    if gotMain.Len() > 1 {
        choice := mainChoice(gotMain)
        mainPKG, _ = gotMain.At(choice).(*dag.Package)
    } else {
        mainPKG, _ = gotMain.Pop().(*dag.Package)
    }

    includeLen := c.extraPkgIncludes()
    staticXtra := 0
    if global.GetBool("-static") {
        staticXtra++
    }

    compiled := path.Join(c.root, mainPKG.Name) + c.suffix

    argv := make([]string, 6+(includeLen*2)+staticXtra)
    i := 0
    argv[i] = c.linker
    i++
    argv[i] = "-o"
    i++
    argv[i] = output
    i++
    argv[i] = "-L"
    i++
    argv[i] = c.root
    i++
    if global.GetBool("-static") {
        argv[i] = "-d"
        i++
    }
    if includeLen > 0 {
        for y := 0; y < includeLen; y++ {
            argv[i] = "-L"
            i++
            argv[i] = c.includes[y]
            i++
        }
    }
    argv[i] = compiled
    i++

    if c.dryrun {
        dryRun(argv)
    } else {
        fmt.Println("linking  :", output)
        handy.StdExecve(argv, true)
    }
}

func mainChoice(pkgs *vector.Vector) int {

    fmt.Println("\n More than one main package found\n")

    for i := 0; i < pkgs.Len(); i++ {
        pk, _ := pkgs.At(i).(*dag.Package)
        fmt.Printf(" type %2d  for: %s\n", i, pk.Name)
    }

    var choice int

    fmt.Printf("\n type your choice: ")

    n, e := fmt.Scanf("%d", &choice)

    if e != nil {
        log.Exitf("%s\n", e)
    }
    if n != 1 {
        log.Exit("failed to read input\n")
    }

    if choice >= pkgs.Len() || choice < 0 {
        log.Exitf(" bad choice: %d\n", choice)
    }

    fmt.Printf(" chosen main-package: %s\n\n", pkgs.At(choice).(*dag.Package).Name)

    return choice
}


func dryRun(argv []string) {
    var cmd string

    for i := 0; i < len(argv); i++ {
        cmd = fmt.Sprintf("%s %s ", cmd, argv[i])
    }

    fmt.Printf("%s || exit 1\n", cmd)
}

func (c *Compiler) extraPkgIncludes() int {
    if c.includes != nil && len(c.includes) > 0 {
        return len(c.includes)
    }
    return 0
}


func CreateTestArgv() []string {

    var numArgs int = 1

    pwd, e := os.Getwd()

    if e != nil {
        log.Exit("[ERROR] could not locate working directory\n")
    }

    arg0 := path.Join(pwd, global.GetString("-test-bin"))

    if global.GetString("-benchmarks") != "" {
        numArgs += 2
    }
    if global.GetString("-match") != "" {
        numArgs += 2
    }
    if global.GetBool("-verbose") {
        numArgs++
    }

    var i = 1
    argv := make([]string, numArgs)
    argv[0] = arg0
    if global.GetString("-benchmarks") != "" {
        argv[i] = "-benchmarks"
        i++
        argv[i] = global.GetString("-benchmarks")
        i++
    }
    if global.GetString("-match") != "" {
        argv[i] = "-match"
        i++
        argv[i] = global.GetString("-match")
        i++
    }
    if global.GetBool("-verbose") {
        argv[i] = "-v"
    }
    return argv
}

func Remove865a(srcdir string) {

    // override IncludeFile to make walker pick up only .[865a] files
    walker.IncludeFile = func(s string) bool {
        return strings.HasSuffix(s, ".8") ||
            strings.HasSuffix(s, ".6") ||
            strings.HasSuffix(s, ".5") ||
            strings.HasSuffix(s, ".a")

    }

    handy.DirOrExit(srcdir)

    compiled := walker.PathWalk(path.Clean(srcdir))

    for i := 0; i < compiled.Len(); i++ {

        if ! global.GetBool("-dryrun") {

            e := os.Remove(compiled.At(i))
            if e != nil {
                log.Printf("[ERROR] could not delete file: %s\n", compiled.At(i))
            } else {
                fmt.Printf("rm: %s\n", compiled.At(i))
            }

        } else {
            fmt.Printf("[dryrun] rm: %s\n", compiled.At(i))
        }
    }
}


func FormatFiles(files *vector.StringVector) {

    var i, argvLen int
    var argv []string
    var tabWidth string = "-tabwidth=4"
    var useTabs string = "-tabindent=false"
    var comments string = "-comments=true"
    var rewRule string = global.GetString("-rew-rule")
    var fmtexec string
    var err os.Error

    fmtexec, err = exec.LookPath("gofmt")

    if err != nil {
        log.Exit("[ERROR] could not find 'gofmt' in $PATH")
    }

    if global.GetString("-tabwidth") != "" {
        tabWidth = "-tabwidth=" + global.GetString("-tabwidth")
    }
    if global.GetBool("-no-comments") {
        comments = "-comments=false"
    }
    if rewRule != "" {
        argvLen++
    }
    if global.GetBool("-tab") {
        useTabs = "-tabindent=true"
    }

    argv = make([]string, 6+argvLen)

    if fmtexec == "" {
        log.Exit("[ERROR] could not find: gofmt\n")
    }

    argv[i] = fmtexec
    i++
    argv[i] = "-w=true"
    i++
    argv[i] = tabWidth
    i++
    argv[i] = useTabs
    i++
    argv[i] = comments
    i++

    if rewRule != "" {
        argv[i] = fmt.Sprintf("-r='%s'", rewRule)
        i++
    }

    for y := 0; y < files.Len(); y++ {
        argv[i] = files.At(y)
        if ! global.GetBool("-dryrun") {
            fmt.Printf("gofmt : %s\n", files.At(y))
            _ = handy.StdExecve(argv, true)
        } else {
            fmt.Printf(" %s\n", strings.Join(argv, " "))
        }
    }
}
