// © Knug Industries 2011 all rights reserved
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package gdmake

import (
    "os"
    "log"
    "bytes"
    "time"
    "go/parser"
    "go/token"
    "go/ast"
    "fmt"
    "strings"
    "cmplr/dag"
    "utilz/handy"
    "utilz/stringbuffer"
    "utilz/stringset"
    "io/ioutil"
)

const (
    Header = iota
    Imports
    Targets
    Playground
    Init
    GoInstall
    Compile
    PackageDef
    PackageStart
    Packages
    Main
)
// makefile structure
var m = map[int]string{
    Header:       HeaderTmpl,
    Imports:      ImportsTmpl,
    Targets:      TargetsTmpl,
    Playground:   PlaygroundTmpl,
    Init:         InitTmpl,
    GoInstall:    GoInstallTmpl,
    Compile:      CompileTmpl,
    PackageDef:   PackageDefTmpl,
    PackageStart: PackageStartTmpl,
    Packages:     "", // this is not static :-)
    Main:         MainTmpl,
}


func Make(fname string, pkgs []*dag.Package, alien []string) {

    if handy.IsFile(fname) {

        modImport, iOk := hasModifiedImports(fname)
        if iOk {
            m[Imports] = modImport
        }

        modPlay, pOk := hasModifiedPlayground(fname)
        if pOk {
            m[Playground] = modPlay
        }

        if pOk || iOk {
            e := os.Rename(fname, fname+".bak")
            if e != nil {
                log.Printf("[WARNING] failed to make backup of: %s\n",fname)
            }
        }
    }

    sb := stringbuffer.New()
    sb.Add(fmt.Sprintf(m[Header], time.UTC()))
    sb.Add(m[Imports])
    sb.Add(m[Targets])
    sb.Add("// PLAYGROUND START\n")
    sb.Add(m[Playground])
    sb.Add("// PLAYGROUND STOP\n")
    sb.Add(m[Init])
    for i := 0; i < len(alien); i++ {
        alien[i] = `"` + alien[i] + `"`
    }
    sb.Add(fmt.Sprintf(m[GoInstall], strings.Join(alien, ",")))
    sb.Add(m[Compile])
    sb.Add(m[PackageDef])
    sb.Add(m[PackageStart])
    for i := 0; i < len(pkgs); i++ {
        sb.Add(pkgs[i].Rep())
    }
    sb.Add("\n}\n")
    sb.Add(m[Main])
    ioutil.WriteFile(fname, sb.Bytes(), 0644)
}

type collector struct {
    deps []string
}

func (c *collector) String() string {
    sb := stringbuffer.New()
    sb.Add("\nimport(\n")
    for i := 0; i < len(c.deps); i++ {
        sb.Add("    "+ c.deps[i] + "\n")
    }
    sb.Add(")\n\n")
    return sb.String()
}

func (c *collector) Visit(node ast.Node) (v ast.Visitor) {
    switch d := node.(type) {
    case *ast.BasicLit:
        c.deps = append(c.deps, d.Value)
    default: // nothing to do if not BasicLit
    }
    return c
}

func hasModifiedImports(fname string) (string, bool) {

    fileset := token.NewFileSet()
    mode := parser.ImportsOnly

    absSynTree, err := parser.ParseFile(fileset, fname, nil, mode)

    if err != nil {
        log.Fatalf("%s\n", err)
    }

    c := &collector{make([]string, 0)}
    ast.Walk(c, absSynTree)

    set := stringset.New()

    set.Add(`"os"`)
    set.Add(`"fmt"`)
    set.Add(`"io/ioutil"`)
    set.Add(`"regexp"`)
    set.Add(`"exec"`)
    set.Add(`"log"`)
    set.Add(`"flag"`)
    set.Add(`"path/filepath"`)

    for i := 0; i < len(c.deps); i++ {
        if ! set.Contains(c.deps[i]) {
            return c.String(), true
        }
    }

    return "", false
}

func hasModifiedPlayground(fname string) (mod string, ok bool) {

    var start, stop, content, playground []byte
    var startOffset, stopOffset int

    start = []byte("// PLAYGROUND START\n")
    stop  = []byte("// PLAYGROUND STOP\n")

    content, err := ioutil.ReadFile(fname)

    if err != nil {
        log.Fatalf("[ERROR] %s\n", err)
    }

    startOffset = bytes.Index(content, start)
    stopOffset  = bytes.Index(content, stop)

    if startOffset == -1 || stopOffset == -1 {
        return "", false
    }

    playground = content[startOffset + len(start):stopOffset]

    ok = ( string(playground) != PlaygroundTmpl )

    return string(playground), ok

}

// Templates for the different parts of the Makefile

var HeaderTmpl = `/* Built : %s */
//-------------------------------------------------------------------
// Auto generated code, but you are encouraged to modify it ☺
// Manual: http://godag.googlecode.com
//-------------------------------------------------------------------

package main
`

var ImportsTmpl = `
import (
    "os"
    "fmt"
    "io/ioutil"
    "regexp"
    "exec"
    "log"
    "flag"
    "path/filepath"
)

`

var TargetsTmpl = `
//-------------------------------------------------------------------
//  User defined targets: flexible pure go way to control build
//-------------------------------------------------------------------

type Target struct {
    desc  string   // description of target
    first func()   // this is called prior to building
    last  func()   // this is called after building
}

/********************************************************************

 Code inside playground is NOT auto generated, each time you run
 this command:

     gd -gdmake=somefilename.go

  this happens:

  if (somefilename.go already exists) {
      transfer playground from somefilename.go to new version
      of somefilename.go, i.e. you never loose the targets you
      create inside the playground..
  }else{
      write somefilename.go with hello world + full target,
      i.e. the default example targets
  }

********************************************************************/

`

var PlaygroundTmpl = `
var targets = map[string]*Target{
    "hello": &Target{
        desc:  "target that says hello to the world",
        first: func() { fmt.Println("hello world"); os.Exit(0) },
        last:  nil,
    },
    "full": &Target{
        desc:  "compile all packges (ignore !modified)",
        first: func() { oldPkgFound = true },
        last:  nil,
    },
}

`

var InitTmpl = `
//-------------------------------------------------------------------
// Execute user defined targets
//-------------------------------------------------------------------

func doFirst() {
    args := flag.Args()
    for i := 0; i < len(args); i++ {
        target, ok := targets[args[i]]
        if ok && target.first != nil {
            target.first()
        }
    }
}

func doLast() {
    args := flag.Args()
    for i := 0; i < len(args); i++ {
        target, ok := targets[args[i]]
        if ok && target.last != nil {
            target.last()
        }
    }
}

//-------------------------------------------------------------------
// Simple way to turn print statements on/off
//-------------------------------------------------------------------

type Say bool

func (s Say) Printf(frmt string, args ...interface{}) {
    if bool(s) {
        fmt.Printf(frmt, args...)
    }
}

func (s Say) Println(args ...interface{}) {
    if bool(s) {
        fmt.Println(args...)
    }
}

//-------------------------------------------------------------------
// Initialize variables, flags and so on
//-------------------------------------------------------------------

var (
    compiler    = ""
    linker      = ""
    suffix      = ""
    backend     = ""
    root        = ""
    output      = ""
    match       = ""
    help        = false
    list        = false
    quiet       = false
    external    = false
    clean       = false
    oldPkgFound = false
)

var includeDirs = []string{"_obj"}

var say = Say(true)


func init() {

    flag.StringVar(&backend, "backend", "gc", "select from [gc,gccgo,express]")
    flag.StringVar(&backend, "B", "gc", "alias for --backend option")
    flag.StringVar(&root, "r", "", "alias for -root (express go-root)")
    flag.StringVar(&root, "root", "", "express go-root, ignores %GOROOT%")
    flag.StringVar(&match, "M", "", "regex to match main package")
    flag.StringVar(&match, "main", "", "regex to match main package")
    flag.StringVar(&output, "o", "", "link main package -> output")
    flag.StringVar(&output, "output", "", "link main package -> output")
    flag.BoolVar(&external, "external", false, "external dependencies")
    flag.BoolVar(&external, "e", false, "external dependencies")
    flag.BoolVar(&quiet, "q", false, "don't print anything but errors")
    flag.BoolVar(&quiet, "quiet", false, "don't print anything but errors")
    flag.BoolVar(&help, "h", false, "help message")
    flag.BoolVar(&help, "help", false, "help message")
    flag.BoolVar(&clean, "clean", false, "delete objects")
    flag.BoolVar(&clean, "c", false, "delete objects")
    flag.BoolVar(&list, "list", false, "list targets for bash autocomplete")

    flag.Usage = func() {
        fmt.Println("\n gdmake.go - makefile in pure go\n")
        fmt.Printf(" usage: %s [OPTIONS] [TARGET]\n\n", os.Args[0])
        fmt.Println(" options:\n")
        fmt.Println("  -h --help         print this menu and exit")
        fmt.Println("  -B --backend      choose backend [gc,gccgo,express]")
        fmt.Println("  -o --output       link main package -> output")
        fmt.Println("  -M --main         regex to match main package")
        fmt.Println("  -c --clean        delete object files")
        fmt.Println("  -q --quiet        quiet unless errors occur")
        fmt.Println("  -r --root         GOROOT for backend express")
        fmt.Println("  -e --external     goinstall external dependencies\n")

        if len(targets) > 0 {
            fmt.Println(" targets:\n")
            for k, v := range targets {
                fmt.Printf("  %-11s  =>   %s\n", k, v.desc)
            }
            fmt.Println("")
        }
    }
}

func initBackend() {
    switch backend {
    case "gc":
        n := archNum()
        compiler, linker, suffix = n+"g", n+"l", "."+n
    case "gccgo", "gcc":
        compiler, linker, suffix = "gccgo", "gccgo", ".o"
        backend = "gccgo"
    case "express":
        compiler, linker, suffix = "vmgc", "vmld", ".vmo"
    default:
        log.Fatalf("[ERROR] unknown backend: %s\n", backend)
    }
}

func archNum() (n string) {
    switch os.Getenv("GOARCH") {
    case "386":
        n = "8"
    case "arm":
        n = "5"
    case "amd64":
        n = "6"
    default:
        log.Fatalf("[ERROR] unknown GOARCH: %s\n", os.Getenv("GOARCH"))
    }
    return
}

`

var GoInstallTmpl = `
//-------------------------------------------------------------------
// External dependencies
//-------------------------------------------------------------------

var alien = []string{%s}

`

var CompileTmpl = `
//-------------------------------------------------------------------
// Functions to build/delete project
//-------------------------------------------------------------------

func osify(pkgs []*Package) {

    for j := 0; j < len(pkgs); j++ {

        if pkgs[j].osified {
            break
        }

        pkgs[j].osified = true
        pkgs[j].output = filepath.FromSlash(pkgs[j].output) + suffix
        for i := 0; i < len(pkgs[j].files); i++ {
            pkgs[j].files[i] = filepath.FromSlash(pkgs[j].files[i])
        }
    }
}

func mkdirs(pkgs []*Package) {
    for i := 0; i < len(pkgs); i++ {
        d, _ := filepath.Split(pkgs[i].output)
        if d != "" && !isDir(d) {
            e := os.MkdirAll(d, 0777)
            if e != nil {
                log.Fatalf("[ERROR] %s\n", e)
            }
        }
    }
}

func compile(pkgs []*Package) {

    osify(pkgs)
    mkdirs(pkgs)

    for i := 0; i < len(pkgs); i++ {
        if oldPkgFound || !pkgs[i].up2date() {
            say.Printf("compiling: %s\n", pkgs[i].loc)
            pkgs[i].compile()
            oldPkgFound = true
        } else {
            say.Printf("up 2 date: %s\n", pkgs[i].loc)
        }
    }
}

func delete(pkgs []*Package) {

    osify(pkgs)

    for j := 0; j < len(pkgs); j++ {
        if isFile(pkgs[j].output) {
            say.Printf("rm: %s\n", pkgs[j].output)
            e := os.Remove(pkgs[j].output)
            if e != nil {
                log.Fatalf("[ERROR] failed to remove: %s\n", pkgs[j].output)
            }
        }
    }

    for i := 0; i < len(includeDirs); i++ {
        if emptyDir(includeDirs[i]){
            say.Printf("rm: %s\n", includeDirs[i])
            e := os.RemoveAll(includeDirs[i])
            if e != nil {
                log.Fatalf("[ERROR] failed to remove: %s\n", includeDirs[i])
            }
        }
    }

}

func link(pkgs []*Package) {

    var mainPackage *Package
    var mainPkgs = make([]*Package, 0)

    for i := 0; i < len(pkgs); i++ {
        if pkgs[i].name == "main" {
            mainPkgs = append(mainPkgs, pkgs[i])
            mainPackage = pkgs[i]
        }
    }

    switch len(mainPkgs) {
    case 0:
        log.Fatalf("[ERROR] no main package found\n")
    case 1: // do nothing... this is good
    default:
        mainPackage = mainChoice(mainPkgs)
    }

    if !oldPkgFound && isFile(output) {
        pkgLastTs, _ := timestamp(pkgs[len(pkgs)-1].output)
        outputTs, ok := timestamp(output)
        if ok && outputTs > pkgLastTs {
            say.Printf("up 2 date: %s\n", output)
            return
        }
    }

    say.Printf("linking  : %s\n", output)

    argv := make([]string, 0)
    argv = append(argv, linker)

    if backend != "gccgo" {
        for i := 0; i < len(includeDirs); i++ {
            argv = append(argv, "-L")
            argv = append(argv, includeDirs[i])
        }
    }

    argv = append(argv, "-o")
    argv = append(argv, output)

    if backend == "gccgo" {
        for i := 0; i < len(pkgs); i++ {
            argv = append(argv, pkgs[i].output)
        }
    }else{
        argv = append(argv, mainPackage.output)
    }

    run(argv)
}

func mainChoice(pkgs []*Package) *Package {

    var cnt, choice int

    for i := 0; i < len(pkgs); i++ {
        ok, _ := regexp.MatchString(match, pkgs[i].loc)
        if ok {
            cnt++
            choice = i
        }
    }

    if cnt == 1 {
        return pkgs[choice]
    }

    fmt.Println("\n More than one main package found\n")

    for i := 0; i < len(pkgs); i++ {
        fmt.Printf(" type %2d  for: %s\n", i, pkgs[i].loc)
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

    fmt.Printf(" chosen main-package: %s\n\n", pkgs[choice].loc)

    return pkgs[choice]
}

func goinstall() {

    argv := make([]string, 4)
    argv[0] = "goinstall"
    argv[1] = "-clean=true"
    argv[2] = "-u=true"

    for i := 0; i < len(alien); i++ {
        say.Printf("goinstall: %s\n", alien[i])
        argv[3] = alien[i]
        run(argv)
    }
}


//-------------------------------------------------------------------
// Utility types and functions
//-------------------------------------------------------------------

type collector struct{
    files []string
}

func (c *collector) VisitDir(pathname string, d *os.FileInfo) bool {
    return true
}

func (c *collector) VisitFile(pathname string, d *os.FileInfo) {
     c.files = append(c.files, pathname)
}

func emptyDir(pathname string) bool {
    if ! isDir(pathname) {
        return false
    }
    errs := make(chan os.Error)
    collect := &collector{make([]string, 0)}
    filepath.Walk(pathname, collect, errs)
    return len(collect.files) == 0
}

func isDir(pathname string) bool {
    fileInfo, err := os.Stat(pathname)
    if err == nil && fileInfo.IsDirectory() {
        return true
    }
    return false
}

func timestamp(s string) (int64, bool) {
    fileInfo, e := os.Stat(s)
    if e == nil && fileInfo.IsRegular() {
        return fileInfo.Mtime_ns, true
    }
    return 0, false
}

func run(argv []string) {

    cmd := exec.Command(argv[0], argv[1:]...)

    // pass-through
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Stdin = os.Stdin

    err := cmd.Start()

    if err != nil {
        log.Fatalf("[ERROR] %s\n", err)
    }

    err = cmd.Wait()

    if err != nil {
        log.Fatalf("[ERROR] %s\n", err)
    }

}

func isFile(pathname string) bool {
    fileInfo, err := os.Stat(pathname)
    if err == nil && fileInfo.IsRegular() {
        return true
    }
    return false
}

func quitter(e os.Error) {
    if e != nil {
        log.Fatalf("[ERROR] %s\n", e)
    }
}


// very basic copy function, dies on error
func cp(from, to string) {
    fromBytes, e := ioutil.ReadFile(from)
    quitter(e)
    e = ioutil.WriteFile(to, fromBytes, 0644)
    quitter(e)
}

func listTargets() {
    if list {
        for k, _ := range targets {
            fmt.Println(k)
        }
        os.Exit(0)
    }
}

`

var PackageDefTmpl = `
//-------------------------------------------------------------------
// Package definition
//-------------------------------------------------------------------

type Package struct {
    name, loc, output string
    osified           bool
    files             []string
}

func (p *Package) up2date() bool {
    mtime, ok := timestamp(p.output)
    if !ok {
        return false
    }
    for i := 0; i < len(p.files); i++ {
        fmtime, ok := timestamp(p.files[i])
        if !ok {
            log.Fatalf("file missing: %s\n", p.files[i])
        }
        if fmtime > mtime {
            return false
        }
    }
    return true
}

func (p *Package) compile() {

    argv := make([]string, 0)
    argv = append(argv, compiler)
    argv = append(argv, "-I")
    for _, inc := range includeDirs {
        argv = append(argv, inc)
    }
    if backend == "gccgo" {
        argv = append(argv, "-c")
    }
    argv = append(argv, "-o")
    argv = append(argv, p.output)
    argv = append(argv, p.files...)

    run(argv)

}

`

var PackageStartTmpl = `
//-------------------------------------------------------------------
// Package info collected by godag
//-------------------------------------------------------------------

var packages = []*Package{
`

var MainTmpl = `
//-------------------------------------------------------------------
// Main - note flags are parsed before we doFirst, i.e. command line
//        options/arguments can be overridden in doFirst targets
//-------------------------------------------------------------------

func main() {

    flag.Parse()
    listTargets() // for bash auto complete
    initBackend() // gc/gcc/express

    doFirst()
    defer doLast()


    if help {
        flag.Usage()
        os.Exit(0)
    }

    if quiet {
        say = Say(false)
    }

    if clean {

        delete(packages)

    } else {

        if external {
            goinstall()
        }

        compile(packages)

        if output != "" {
            link(packages)
        }

    }
}
`
