/* Built : Sun Sep 11 19:43:06 UTC 2011 */
//-------------------------------------------------------------------
// Auto generated code, but you are encouraged to modify it ☺
// Manual: http://godag.googlecode.com
//-------------------------------------------------------------------

package main

import(
    "io"
    "os"
    "fmt"
    "regexp"
    "exec"
    "bytes"
    "log"
    "flag"
    "compress/gzip"
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

     gd -gdmk=somefilename.go

  this happens:

  if (somefilename.go already exists) {
      transfer playground from somefilename.go to new version
      of somefilename.go, i.e. you never loose the targets you
      create inside the playground..
  }else{
      write somefilename.go with example targets
  }

********************************************************************/

// PLAYGROUND START


var targets = map[string]*Target{
    "clean": &Target{
        desc:  "rm -rf _obj/ _gdmk.[856o]",
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
        desc:  "compile and link 'gd' (alias for -o=gd)",
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
        desc:  "update golang to latest release if necessary",
        first: releaseDoFirst,
        last:  nil,
    },
    "debian": &Target{
        desc:  "create a debian package of godag",
        first: debianDoFirst,
        last:  debianDoLast,
    },
}

// debian target
func debianDoFirst() {

    // see if required executables can be found
    debianSanity()

    // normail build first, i.e. build 'gd'
    buildDoFirst()
}

var debianCopyright=`Name: godag
Maintainer: bjarne holen <bjarneh@ifi.uio.no>
Source: https://godag.googlecode.com/hg

Files: *
Copyright: 2011, bjarne holen <bjarneh@ifi.uio.no>
License: GPL-3

License: GPL-3
On Debian systems, the complete text of the GNU General Public License
version 3 can be found in '/usr/share/common-licenses/GPL-3'.
`

var debianChangelog =`godag (0.2.0) devel; urgency=low

    * The actual changelog can be found in changelog...

 -- Bjarne Holen <bjarneh@ifi.uio.no>  Thu, 05 May 2011 14:07:28 -0400
`

var debianControl=`Package: godag
Version: 0.2
Section: devel
Priority: optional
Architecture: %s
Depends:
Suggests: gccgo,golang
Conflicts:
Replaces:
Installed-Size: %d
Maintainer: Bjarne Holen <bjarneh@ifi.uio.no>
Description: Golang/Go compiler front-end.
 Godag automatically builds projects written in golang,
 by inspecting source-code imports to calculate compile order.
 Unit-testing, formatting and installation of external 
 libraries are also automated. The default back-end is gc,
 other back-ends have partial support: gccgo, express.
`

func debianDoLast() {

    var debArch string
    var dirs []string

    switch os.Getenv("GOARCH") {
    case "386":
        debArch = "i386"
    case "":
        log.Fatalf("[ERROR] GOARCH == ''\n")
    default:
        debArch = os.Getenv("GOARCH")
    }

    dirs = []string{"debian/DEBIAN",
                    "debian/usr/bin",
                    "debian/usr/share/man/man1",
                    "debian/usr/share/doc/godag",
                    "debian/etc/bash_completion.d",}

    say.Println("debian   : make structure")

    for i := 0; i < len(dirs); i++ {
        e := os.MkdirAll(dirs[i], 0755)
        quitter(e)
    }

    e := os.Rename("gd", "debian/usr/bin/gd")
    quitter(e)

    say.Println("debian   : copy files")

    // copyGzip( from, to, gzipfile )
    copyGzip("util/gd-completion.sh", "debian/etc/bash_completion.d/gd", false)
    e = os.Chmod("debian/etc/bash_completion.d/gd", 0755)
    quitter(e)
    copyGzip("util/gdmk-completion.sh", "debian/etc/bash_completion.d/gdmk", false)
    e = os.Chmod("debian/etc/bash_completion.d/gdmk", 0755)
    quitter(e)
    copyGzip("util/gd.1", "debian/usr/share/man/man1/gd.1.gz", true)
    copyGzip("util/gd.1", "debian/usr/share/man/man1/godag.1.gz", true)
    copyGzipStringBuffer(debianCopyright, "debian/usr/share/doc/godag/copyright", false)
    copyGzipStringBuffer(debianChangelog, "debian/usr/share/doc/godag/changelog.Debian.gz", true)
    copyGzipStringBuffer("/etc/bash_completion.d/gd\n/etc/bash_completion.d/gdmk\n", "debian/DEBIAN/conffiles", false)

    if isDir(".hg") {
        hglog, e := exec.Command("hg", "log").CombinedOutput()
        quitter(e)
        copyGzipByteBuffer(hglog, "debian/usr/share/doc/godag/changelog.gz", true)
    }else{// git
        gitlog, e := exec.Command("git", "log").CombinedOutput()
        quitter(e)
        copyGzipByteBuffer(gitlog, "debian/usr/share/doc/godag/changelog.gz", true)
    }

    // find size of debian package
    var totalSize int64 = int64(len(debianControl))

    errs := make(chan os.Error)
    collect := &collector{make([]string, 0), nil}
    filepath.Walk("debian", collect, errs)

    for i := 0; i < len(collect.files); i++ {
        finfo, e := os.Stat(collect.files[i])
        quitter(e)
        totalSize += finfo.Size
    }

    tmpBuf := fmt.Sprintf(debianControl, debArch, totalSize)
    copyGzipStringBuffer(tmpBuf, "debian/DEBIAN/control", false)

    run([]string{"fakeroot", "dpkg-deb", "--build", "debian"})
    e = os.Rename("debian.deb",  "godag_0.2-0_" + debArch + ".deb")
    quitter(e)

    say.Println("debian   : rm -rf ./debian")
    e = os.RemoveAll("debian")
    quitter(e)
}

func debianSanity() {

    executables := []string{"dpkg-deb", "fakeroot"}

    if isDir(".hg") {
        executables = append(executables, "hg")
    } else if isDir(".git") {
        executables = append(executables, "git")
    } else {
        log.Fatalf("[ERROR] neither .hg or .git was found (needed for changelog)\n")
    }

    for i := 0; i < len(executables); i++ {
        _, e := exec.LookPath(executables[i])
        if e != nil {
            log.Fatalf("[ERROR] missing: %s\n", executables[i])
        }
    }
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
    "csv",
    "ebnf",
    "encoding",
    "expvar",
    "fmt",
    "gob",
    "index",
    "json",
    "log",
    "mail",
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

    say.Printf("[ testok ] ")

    goroot := os.Getenv("GOROOT")

    if goroot == "" {
        log.Fatalf("[ERROR] missing GOROOT variable\n")
    }

    dirReq := func(s string) bool { return true }

    fileReq := func(s string) bool {
        return strings.HasSuffix(s, ".go")
    }

    from := filepath.Join(goroot, "src", "pkg")
    to := fmt.Sprintf("tmp-pkgroot-%d", time.Seconds())

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

    say.Printf("[ stdlib ]  ")

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

    from := filepath.Join(goroot, "src", "pkg")
    to := fmt.Sprintf("tmp-pkgroot-%d", time.Seconds())

    say.Printf("pure go part of stdlib -> %s\n", to)

    for i := 0; i < len(golibs); i++ {
        recCopyStrip(filepath.Join(from, golibs[i]),
            filepath.Join(to, golibs[i]),
            fileReq, dirReq)
    }

    os.Exit(0)
}

// recursive copy that strips away main packages + testdata
func recCopyStrip(from, to string, fileReq, dirReq func(s string) bool) {

    if !isDir(from) {
        return
    }

    if !isDir(to) {
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
            copyGzip(next, filepath.Join(to, dirnames[i]), false)
        }
        if isDir(next) && dirReq(dirnames[i]) {
            recCopyStrip(next, filepath.Join(to, dirnames[i]), fileReq, dirReq)
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
        full:   "main",
        output: "_gdmk",
        files:  []string{"_gdmk.go"},
    },
}

func updateDoFirst() {

    initBackend()
    output = "gdmk"

    if os.Getenv("GOOS") == "windows" {
        fmt.Println("\n[ update ]\n")
        fmt.Println(" this target does not work on Windows,")
        fmt.Println(" to update 'gdmk' run these commands\n")
        fmt.Printf(" %s _gdmk.go\n", compiler)
        if linker == "8l" {
            fmt.Println(" 8l -o gdmk.exe _gdmk.8\n")
        }else{
            fmt.Println(" 6l -o gdmk.exe _gdmk.6\n")
        }
        os.Exit(1)
    }

    compile(self)
    link(self)

    os.Exit(0)
}

// hgup
func releaseDoFirst() {

    if os.Getenv("GOOS") == "windows" {
        fmt.Println("\n[ release ]\n")
        fmt.Println(" this target only works on Linux/Unix/Mac\n")
        os.Exit(1)
    }

    goroot := os.Getenv("GOROOT")

    if goroot == "" {
        log.Fatalf("[ERROR] missing GOROOT variable\n")
    }

    srcdir := filepath.Join(goroot, "src")

    say.Printf("> \u001B[32mcd $GOROOT/src\u001B[0m\n")
    e := os.Chdir(srcdir)
    quitter(e)

    current, e := exec.Command("hg", "id", "-i").CombinedOutput()
    quitter(e)

    say.Println("> \u001B[32mhg pull\u001B[0m")
    run([]string{"hg", "pull"})

    say.Println("> \u001B[32mhg update release\u001B[0m")
    run([]string{"hg", "update", "release"})

    updated, e := exec.Command("hg", "id", "-i").CombinedOutput()

    if e != nil {
        log.Fatalf("[ERROR] %s\n", e)
    }

    if string(current) != string(updated) {
        say.Println("> \u001B[32mbash make.bash\u001B[0m")
        run([]string{"bash", "make.bash"})
    } else {
        say.Println("[\u001B[32m already at release version\u001B[0m ]")
    }

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
        fmt.Println("\n gdmk - makefile in pure go\n")
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


//-------------------------------------------------------------------
// Package info collected by godag
//-------------------------------------------------------------------

var packages = []*Package{
    &Package{
        name:   "stringbuffer",
        full:    "utilz/stringbuffer",
        output: "_obj/utilz/stringbuffer",
        files:  []string{"src/utilz/stringbuffer.go"},
    },
    &Package{
        name:   "walker",
        full:    "utilz/walker",
        output: "_obj/utilz/walker",
        files:  []string{"src/utilz/walker.go"},
    },
    &Package{
        name:   "stringset",
        full:    "utilz/stringset",
        output: "_obj/utilz/stringset",
        files:  []string{"src/utilz/stringset.go"},
    },
    &Package{
        name:   "timer",
        full:    "utilz/timer",
        output: "_obj/utilz/timer",
        files:  []string{"src/utilz/timer.go"},
    },
    &Package{
        name:   "gopt",
        full:    "parse/gopt",
        output: "_obj/parse/gopt",
        files:  []string{"src/parse/gopt.go","src/parse/option.go"},
    },
    &Package{
        name:   "global",
        full:    "utilz/global",
        output: "_obj/utilz/global",
        files:  []string{"src/utilz/global.go"},
    },
    &Package{
        name:   "handy",
        full:    "utilz/handy",
        output: "_obj/utilz/handy",
        files:  []string{"src/utilz/handy.go"},
    },
    &Package{
        name:   "say",
        full:    "utilz/say",
        output: "_obj/utilz/say",
        files:  []string{"src/utilz/say.go"},
    },
    &Package{
        name:   "dag",
        full:    "cmplr/dag",
        output: "_obj/cmplr/dag",
        files:  []string{"src/cmplr/dag.go"},
    },
    &Package{
        name:   "gdmake",
        full:    "cmplr/gdmake",
        output: "_obj/cmplr/gdmake",
        files:  []string{"src/cmplr/gdmake.go"},
    },
    &Package{
        name:   "compiler",
        full:    "cmplr/compiler",
        output: "_obj/cmplr/compiler",
        files:  []string{"src/cmplr/compiler.go"},
    },
    &Package{
        name:   "main",
        full:    "start/main",
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