/* Built : Thu Jul  7 08:32:39 UTC 2011 */
//-------------------------------------------------------------------
// Auto generated code, but you are encouraged to modify it ☺
// Manual: http://godag.googlecode.com
//-------------------------------------------------------------------

package main

import(
    "os"
    "fmt"
    "regexp"
    "exec"
    "log"
    "flag"
    "io/ioutil"
    "path/filepath"
    "time"
    "strings"
)


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

// PLAYGROUND START


var targets = map[string]*Target{
    "clean": &Target{
        desc:  "rm -rf _obj/ util/gdmake.[856o]",
        first: cleanDoFirst,
        last:  nil,
    },
    "uninstall": &Target{
        desc:  "target:clean + rf -f $HOME/bin/gd $GOBIN/gd",
        first: uninstallDoFirst,
        last:  nil,
    },
    "install": &Target{
        desc:  "target:build + mv 'gd' $HOME/bin || $GOBIN",
        first: buildDoFirst,
        last:  installDoLast,
    },
    "update": &Target{
        desc:  "recompile yourself (if you add targets etc.)",
        first: updateDoFirst,
        last:  nil,
    },
    "build": &Target{
        desc:  "compile code in src and link 'gd'",
        first: buildDoFirst,
        last:  nil,
    },
    "stdlib": &Target{
        desc:  "copy pure go part of standard library",
        first: stdlibDoFirst,
        last:  nil,
    },
    "testok": &Target{
        desc:  "copy pure go testable part of standard library",
        first: testokDoFirst,
        last:  nil,
    },
    "release": &Target{
        desc:  "update godag to latest release",
        first: releaseDoFirst,
        last:  nil,
    },
}

// build target
func buildDoFirst() {
    clean = false // just in case
    output = "gd"
}

// install target
func installDoLast() {

    var tmp, name, home string

    home = os.Getenv("HOME")

    if home != "" {
        tmp = filepath.Join(home, "bin")
        if isDir(tmp) {
            name = filepath.Join(tmp, "gd")
        }
    }

    if name == "" {
        name = filepath.Join(os.Getenv("GOBIN"), "gd")
    }

    say.Printf("move gd  : %s\n", name)

    os.Rename("gd", name)
}


// needed for stdlib, testok targets
var golibs = []string{
    "archive",
    "compress",
    "container",
    "flag",
    "html",
    "http",
    "image",
    "mime",
    "patch",
    "rpc",
    "strconv",
    "tabwriter",
    "template",
    "io",
// packages above this line cannot be tested without modification
    "asn1",
    "bufio",
    "cmath",
    "ebnf",
    "encoding",
    "expvar",
    "fmt",
    "gob",
    "hash",
    "index",
    "json",
    "log",
    "netchan",
    "rand",
    "reflect",
    "regexp",
    "scanner",
    "smtp",
    "sort",
    "strings",
    "syslog",
    "testing",
    "try",
    "unicode",
    "unsafe",
    "utf16",
    "utf8",
    "websocket",
    "xml",
}

// testok target
func testokDoFirst() {

    goroot := os.Getenv("GOROOT")

    if goroot == "" {
        log.Fatalf("[ERROR] missing GOROOT variable\n")
    }

    dirReq := func(s string) bool { return true }

    fileReq := func(s string) bool {
        return strings.HasSuffix(s, ".go")
    }

    from := filepath.Join(goroot,"src","pkg")
    to   := fmt.Sprintf("tmp-pkgroot-%d", time.Seconds())

    say.Printf("testable part of stdlib -> %s\n", to)

    testable := golibs[14:]

    for i := 0; i < len(testable); i++ {
        recCopyStrip(filepath.Join(from, testable[i]),
                     filepath.Join(to, testable[i]),
                     fileReq, dirReq)
    }

    os.Exit(0)

}

// stdlib target
func stdlibDoFirst() {

    goroot := os.Getenv("GOROOT")

    if goroot == "" {
        log.Fatalf("[ERROR] missing GOROOT variable\n")
    }

    dirReq := func(s string) bool {
        return s != "testdata"
    }

    fileReq := func(s string) bool {
        return strings.HasSuffix(s, ".go") && !strings.HasSuffix(s, "_test.go")
    }

    from := filepath.Join(goroot,"src","pkg")
    to   := fmt.Sprintf("tmp-pkgroot-%d", time.Seconds())

    say.Printf("pure go part of stdlib -> %s\n", to)

    for i := 0; i < len(golibs); i++ {
        recCopyStrip(filepath.Join(from, golibs[i]),
                     filepath.Join(to, golibs[i]),
                     fileReq, dirReq)
    }

    os.Exit(0)
}

// recursive copy that strips away main packages + testdata
func recCopyStrip(from, to string, fileReq, dirReq func(s string)bool) {

    if ! isDir(from){
        return
    }

    if ! isDir(to){
        e := os.MkdirAll(to, 0777)
        quitter(e)
    }

    fromFile, e := os.Open(from)
    quitter(e)
    defer fromFile.Close()

    dirnames, e := fromFile.Readdirnames(-1)
    quitter(e)

    for i := 0; i < len(dirnames); i++ {
        next := filepath.Join(from, dirnames[i])
        if isFile(next) && fileReq(next) {
            cp(next, filepath.Join(to, dirnames[i]))
        }
        if isDir(next) && dirReq(dirnames[i]) {
            recCopyStrip(next, filepath.Join(to, dirnames[i]), fileReq,dirReq)
        }
    }
}


// clean target
func cleanDoFirst() {

    delete(packages)
    delete(self)

    os.Exit(0)
}

// uninstall target
func uninstallDoFirst() {

    delete(packages)

    p1 := filepath.Join(os.Getenv("GOBIN"), "gd")

    if isFile(p1) {
        say.Printf("rm: %s\n", p1)
        e := os.Remove(p1)
        quitter(e)
    }

    p2 := filepath.Join(os.Getenv("HOME"), "bin", "gd")

    if isFile(p2) {
        say.Printf("rm: %s\n", p2)
        e := os.Remove(p2)
        quitter(e)
    }

    os.Exit(0)

}

// update target
var self = []*Package{
    &Package{
        name:   "main",
        loc:    "main",
        output: "util/main",
        files:  []string{"util/gdmake.go"},
    },
}

func updateDoFirst() {

    initBackend()
    output = "gdmake"

    if os.Getenv("GOOS") == "windows" {
        output += ".exe"
    }

    compile(self)
    link(self)

    os.Exit(0)
}

// hgup
func releaseDoFirst(){

    pull := []string{"hg", "pull"}
    upRelease := []string{"hg", "update", "release"}

    say.Println("[hg pull]")
    run(pull)

    say.Println("[hg update release]")
    run(upRelease)

    os.Exit(0)
}


// PLAYGROUND STOP

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


//-------------------------------------------------------------------
// External dependencies
//-------------------------------------------------------------------

var alien = []string{}


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


//-------------------------------------------------------------------
// Package info collected by godag
//-------------------------------------------------------------------

var packages = []*Package{
    &Package{
        name:   "stringbuffer",
        loc:    "utilz/stringbuffer",
        output: "_obj/utilz/stringbuffer",
        files:  []string{"src/utilz/stringbuffer.go"},
    },
    &Package{
        name:   "walker",
        loc:    "utilz/walker",
        output: "_obj/utilz/walker",
        files:  []string{"src/utilz/walker.go"},
    },
    &Package{
        name:   "stringset",
        loc:    "utilz/stringset",
        output: "_obj/utilz/stringset",
        files:  []string{"src/utilz/stringset.go"},
    },
    &Package{
        name:   "timer",
        loc:    "utilz/timer",
        output: "_obj/utilz/timer",
        files:  []string{"src/utilz/timer.go"},
    },
    &Package{
        name:   "gopt",
        loc:    "parse/gopt",
        output: "_obj/parse/gopt",
        files:  []string{"src/parse/gopt.go","src/parse/option.go"},
    },
    &Package{
        name:   "global",
        loc:    "utilz/global",
        output: "_obj/utilz/global",
        files:  []string{"src/utilz/global.go"},
    },
    &Package{
        name:   "handy",
        loc:    "utilz/handy",
        output: "_obj/utilz/handy",
        files:  []string{"src/utilz/handy.go"},
    },
    &Package{
        name:   "say",
        loc:    "utilz/say",
        output: "_obj/utilz/say",
        files:  []string{"src/utilz/say.go"},
    },
    &Package{
        name:   "dag",
        loc:    "cmplr/dag",
        output: "_obj/cmplr/dag",
        files:  []string{"src/cmplr/dag.go"},
    },
    &Package{
        name:   "gdmake",
        loc:    "cmplr/gdmake",
        output: "_obj/cmplr/gdmake",
        files:  []string{"src/cmplr/gdmake.go"},
    },
    &Package{
        name:   "compiler",
        loc:    "cmplr/compiler",
        output: "_obj/cmplr/compiler",
        files:  []string{"src/cmplr/compiler.go"},
    },
    &Package{
        name:   "main",
        loc:    "start/main",
        output: "_obj/start/main",
        files:  []string{"src/start/main.go"},
    },

}

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
