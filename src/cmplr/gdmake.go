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
    set.Add(`"io"`)
    set.Add(`"fmt"`)
    set.Add(`"strings"`)
    set.Add(`"compress/gzip"`)
    set.Add(`"bytes"`)
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
    "io"
    "strings"
    "compress/gzip"
    "fmt"
    "bytes"
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
      write somefilename.go with example targets
  }

********************************************************************/

`

var PlaygroundTmpl = `
var targets = map[string]*Target{
    "full": &Target{
        desc:  "compile all packges (ignore !modified)",
        first: func() { oldPkgFound = true },
        last:  nil,
    },
    "install": &Target{
        desc:  "install package in a typical Go fashion",
        first: installDoFirst,
        last:  nil,
    },
    "uninstall": &Target{
        desc:  "remove object files from GOPATH||GOROOT",
        first: func() {
            installDoFirst() // same setup
            clean = true
        },
        last:  nil,
    },
}

func installDoFirst(){

    var placement string

    var env = map[string]string{
        "GOOS"   : os.Getenv("GOOS"),
        "GOARCH" : os.Getenv("GOARCH"),
        "GOROOT" : os.Getenv("GOROOT"),
        "GOPATH" : os.Getenv("GOPATH"),
    }

    if env["GOARCH"] == "" || env["GOOS"] == "" || env["GOROOT"] == "" {
        log.Fatalf("GOARCH, GOROOT and GOOS variables must be set\n")
    }

    stub := env["GOOS"] + "_" + env["GOARCH"]

    if env["GOPATH"] != "" {
        placement = filepath.Join(env["GOPATH"], "pkg", stub)
    }else{
        placement = filepath.Join(env["GOROOT"], "pkg", stub)
    }

    includeDir = placement

    for i := 0; i < len(packages); i++ {
        packages[i].output = placement + packages[i].output[4:]
    }
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

var includeDir = "_obj"

var say = Say(true)


func init() {

    flag.StringVar(&backend, "backend", "gc", "select from [gc,gccgo,express]")
    flag.StringVar(&backend, "B", "gc", "alias for --backend option")
    flag.StringVar(&root, "I", "", "import package directory")
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
        fmt.Println("  -e --external     goinstall external dependencies")
        fmt.Println("  -I                import package directory\n")

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
            say.Printf("compiling: %s\n", pkgs[i].full)
            pkgs[i].compile()
            oldPkgFound = true
        } else {
            say.Printf("up 2 date: %s\n", pkgs[i].full)
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

    if emptyDir(includeDir){
        say.Printf("rm: %s\n", includeDir)
        e := os.RemoveAll(includeDir)
        if e != nil {
            log.Fatalf("[ERROR] failed to remove: %s\n", includeDir)
        }
    }

}

func link(pkgs []*Package) {

    if output == "" {
        return
    }

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

        argv = append(argv, "-L")
        argv = append(argv, includeDir)

        if root != "" {
            argv = append(argv, "-L")
            argv = append(argv, root)
        }
    }

    argv = append(argv, "-o")
    argv = append(argv, output)

    if backend == "gccgo" {
        for i := 0; i < len(pkgs); i++ {
            argv = append(argv, pkgs[i].output)
        }
        if root != "" {
            errs := make(chan os.Error)
            collect := &collector{make([]string, 0), nil}
            collect.filter = func(s string)bool{
                return strings.HasSuffix(s, ".o")
            }
            filepath.Walk(root, collect, errs)
            for i := 0; i < len(collect.files); i++ {
                argv = append(argv, collect.files[i])
            }
        }
    }else{
        argv = append(argv, mainPackage.output)
    }

    run(argv)
}

func mainChoice(pkgs []*Package) *Package {

    var cnt, choice int

    for i := 0; i < len(pkgs); i++ {
        ok, _ := regexp.MatchString(match, pkgs[i].full)
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
        fmt.Printf(" type %2d  for: %s\n", i, pkgs[i].full)
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

    fmt.Printf(" chosen main-package: %s\n\n", pkgs[choice].full)

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
    filter func(string)bool
}

func (c *collector) VisitDir(pathname string, d *os.FileInfo) bool {
    return true
}

func (c *collector) VisitFile(pathname string, d *os.FileInfo) {
    if c.filter != nil {
        if c.filter(pathname) {
            c.files = append(c.files, pathname)
        }
    }else{
        c.files = append(c.files, pathname)
    }
}

func emptyDir(pathname string) bool {
    if ! isDir(pathname) {
        return false
    }
    errs := make(chan os.Error)
    collect := &collector{make([]string, 0), nil}
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


func copyGzipStringBuffer(from string, to string, gzipFile bool) {
    buf := bytes.NewBufferString(from)
    copyGzipReader(buf, to, gzipFile)
}

func copyGzipByteBuffer(from []byte, to string, gzipFile bool){
    buf := bytes.NewBuffer(from)
    copyGzipReader(buf, to, gzipFile)
}

func copyGzip(from, to string, gzipFile bool) {

    var err os.Error
    var fromFile *os.File

    fromFile, err = os.Open(from)
    quitter(err)

    defer fromFile.Close()

    copyGzipReader(fromFile, to, gzipFile)
}

func copyGzipReader(fromReader io.Reader, to string, gzipFile bool) {

    var err os.Error
    var toFile io.WriteCloser

    toFile, err = os.Create(to)
    quitter(err)

    if gzipFile {
        toFile, err = gzip.NewWriterLevel(toFile, gzip.BestCompression)
        quitter(err)
    }

    defer toFile.Close()

    _, err = io.Copy(toFile, fromReader)

    quitter(err)
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
    name, full, output string
    osified            bool
    files              []string
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
    argv = append(argv, includeDir)

    if root != "" {
        argv = append(argv, "-I")
        argv = append(argv, root)
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

    if quiet {
        say = Say(false)
    }

    switch {
    case help:
        flag.Usage()
    case clean:
        delete(packages)
    case external:
        goinstall()
    default:
        compile(packages)
        link(packages)
    }

}
`
