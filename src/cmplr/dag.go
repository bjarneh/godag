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

package dag

import (
    "fmt"
    "go/ast"
    "go/parser"
    "go/token"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "regexp"
    "strings"
    "sync"
    "time"
    "utilz/global"
    "utilz/handy"
    "utilz/say"
    "utilz/stringbuffer"
    "utilz/stringset"
)

var locker = new(sync.Mutex)
var oldPkgFound bool // false

type Dag map[string]*Package // package-name -> Package object

type Package struct {
    Indegree        int
    Name, ShortName string   // absolute path, basename
    Argv            []string // command needed to compile package
    Files           []string // relative path of files
    dependencies    *stringset.StringSet
    children        []*Package // packages that depend on this
    waiter          *sync.WaitGroup
    needsCompile    bool
    lock            *sync.Mutex
}

type TestCollector struct {
    TestFuncs    []string
    BenchFuncs   []string
    ExampleFuncs []string
}

type initCollector struct {
    hasInit bool
}

func New() Dag {
    return make(map[string]*Package)
}

func newPackage() *Package {
    p := new(Package)
    p.Indegree = 0
    p.Files = make([]string, 0)
    p.dependencies = stringset.New()
    p.children = make([]*Package, 0)
    p.waiter = nil
    p.needsCompile = false // yeah yeah..
    p.lock = new(sync.Mutex)
    return p
}

func newTestCollector() *TestCollector {
    t := new(TestCollector)
    t.TestFuncs = make([]string, 0)
    t.BenchFuncs = make([]string, 0)
    t.ExampleFuncs = make([]string, 0)
    return t
}

func (d Dag) Parse(root string, files []string) {

    root = addSeparatorPath(root)

    var e, pkgname string

    for i := 0; i < len(files); i++ {
        e = files[i]
        tree := getSyntaxTreeOrDie(e, parser.ImportsOnly)
        dir, _ := filepath.Split(e)
        unroot := dir[len(root):len(dir)]
        shortname := tree.Name.String()

        // if package name == directory name -> assume stdlib organizing
        if len(unroot) > 1 && filepath.Base(dir) == shortname {
            pkgname = unroot[:len(unroot)-1]
        } else {
            pkgname = filepath.Join(unroot, shortname)
        }

        pkgname = filepath.ToSlash(pkgname)

        _, ok := d[pkgname]
        if !ok {
            d[pkgname] = newPackage()
            d[pkgname].Name = pkgname
            d[pkgname].ShortName = shortname
        }

        ast.Walk(d[pkgname], tree)
        d[pkgname].Files = append(d[pkgname].Files, e)
    }
}

func (d Dag) addEdge(from, to string) {
    fromNode := d[from]
    toNode := d[to]
    fromNode.children = append(fromNode.children, toNode)
    toNode.Indegree++
}
// note that nothing is done in order to check if dependencies
// are valid if they are not part of the actual source-tree.

func (d Dag) GraphBuilder() {

    for k, v := range d {
        for dep := range v.dependencies.Iter() {
            if d.localDependency(dep) {
                d.addEdge(dep, k)
                ///fmt.Printf("local:  %s \n", dep);
            }
        }
    }
}

func (d Dag) Alien() (set *stringset.StringSet) {

    set = stringset.New()

    for _, v := range d {
        for dep := range v.dependencies.Iter() {
            if !d.localDependency(dep) {
                set.Add(dep)
            }
        }
    }

    for u := range set.Iter() {
        if !seemsExternal(u) {
            set.Remove(u)
        }
    }

    return set
}

func (d Dag) External() {

    var err error
    var argv []string
    var tmp string
    var set *stringset.StringSet
    var i int = 0

    set = d.Alien()

    argv = make([]string, 0)

    tmp, err = exec.LookPath("goinstall")

    if err != nil {
        log.Fatalf("[ERROR] %s\n", err)
    }

    argv = append(argv, tmp)

    if global.GetBool("-verbose") {
        argv = append(argv, "-v=true")
    }

    argv = append(argv, "-u=true")
    argv = append(argv, "-clean=true")

    i = len(argv)
    argv = append(argv, "dummy")

    for u := range set.Iter() {
        argv[i] = u
        if global.GetBool("-dryrun") {
            fmt.Printf("%s || exit 1\n", strings.Join(argv, " "))
        } else {
            say.Printf("goinstall: %s\n", u)
            handy.StdExecve(argv, true)
        }
    }

}

// If import starts with one of these, it seems legal...
//
//  bitbucket.org/
//  github.com/
//  [^.]+\.googlecode\.com/
//  launchpad.net/
func seemsExternal(imprt string) bool {

    if strings.HasPrefix(imprt, "bitbucket.org/") {
        return true
    } else if strings.HasPrefix(imprt, "github.com/") {
        return true
    } else if strings.HasPrefix(imprt, "launchpad.net/") {
        return true
    }

    ok, _ := regexp.MatchString("[^.]\\.googlecode\\.com\\/.*", imprt)

    return ok
}

func (d Dag) MakeDotGraph(filename string) {

    var file *os.File
    var fileinfo *os.FileInfo
    var e error
    var sb *stringbuffer.StringBuffer

    fileinfo, e = os.Stat(filename)

    if e == nil {
        if fileinfo.IsRegular() {
            e = os.Remove(fileinfo.Name)
            if e != nil {
                log.Fatalf("[ERROR] failed to remove: %s\n", filename)
            }
        }
    }

    sb = stringbuffer.NewSize(500)
    file, e = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)

    if e != nil {
        log.Fatalf("[ERROR] %s\n", e)
    }

    sb.Add("digraph depgraph {\n\trankdir=LR;\n")

    for _, v := range d {
        v.DotGraph(sb)
    }

    sb.Add("}\n")

    file.WriteString(sb.String())

    file.Close()

}

func (d Dag) MakeMainTest(root string) ([]*Package, string, string) {

    var (
        i         int
        lname     string
        sname     string
        tmplib    string
        tmpdir    string
        tmpstub   string
        tmpfile   string
        collector *TestCollector
    )

    sbImports := stringbuffer.NewSize(300)
    imprtSet := stringset.New()
    sbTests := stringbuffer.NewSize(1000)
    sbBench := stringbuffer.NewSize(1000)
    sbExample := stringbuffer.NewSize(1000)

    sbImports.Add("\n// autogenerated code\n\n")
    sbImports.Add("package main\n\n")
    imprtSet.Add("import \"regexp\"\n")
    imprtSet.Add("import \"testing\"\n")

    sbTests.Add("\n\nvar tests = []testing.InternalTest{\n")
    sbBench.Add("\n\nvar benchmarks = []testing.InternalBenchmark{\n")
    sbExample.Add("\n\nvar examples = []testing.InternalExample{\n")

    for _, v := range d {

        sname = v.ShortName
        lname = v.ShortName

        collector = newTestCollector()

        if strings.HasSuffix(v.ShortName, "_test") {

            for i = 0; i < len(v.Files); i++ {
                tree := getSyntaxTreeOrDie(v.Files[i], parser.ParseComments)
                ast.Walk(collector, tree)
            }

        } else {

            for i = 0; i < len(v.Files); i++ {
                if strings.HasSuffix(v.Files[i], "_test.go") {
                    tree := getSyntaxTreeOrDie(v.Files[i], parser.ParseComments)
                    ast.Walk(collector, tree)
                }
            }
        }

        if collector.FoundAnything() {

            if hasSlash(v.Name) {
                lname = removeSlashAndDot(v.Name)
                imprtSet.Add(fmt.Sprintf("import %s \"%s\"\n", lname, v.Name))
            } else {
                imprtSet.Add(fmt.Sprintf("import \"%s\"\n", v.Name))
            }

            // add tests
            for i := 0; i < len(collector.TestFuncs); i++ {
                fn := collector.TestFuncs[i]
                sbTests.Add(fmt.Sprintf(
                    "testing.InternalTest{\"%s.%s\", %s.%s },\n",
                    sname, fn, lname, fn))
            }

            // add benchmarks
            for i := 0; i < len(collector.BenchFuncs); i++ {
                fn := collector.BenchFuncs[i]
                sbBench.Add(fmt.Sprintf(
                    "testing.InternalBenchmark{\"%s.%s\", %s.%s },\n",
                    sname, fn, lname, fn))
            }

            // add examples ( not really )
            for i := 0; i < len(collector.ExampleFuncs); i++ {
                // fn := collector.ExampleFuncs[i] //TODO add comment which seems to be what we compare against..
                // sbExample.Add(fmt.Sprintf("testing.InternalExample{\"%s.%s\", %s.%s },\n", sname, fn, lname, fn))
            }
        }

    }

    sbTests.Add("};\n")
    sbBench.Add("};\n")
    sbExample.Add("};\n")

    for im := range imprtSet.Iter() {
        sbImports.Add(im)
    }

    sbTotal := stringbuffer.NewSize(sbImports.Len() +
        sbTests.Len() +
        sbBench.Len() + 100)
    sbTotal.Add(sbImports.String())
    sbTotal.Add(sbTests.String())
    sbTotal.Add(sbBench.String())
    sbTotal.Add(sbExample.String())

    sbTotal.Add("func main(){\n")
    sbTotal.Add("testing.Main(regexp.MatchString, tests, benchmarks, examples);\n}\n\n")

    tmpstub = fmt.Sprintf("tmp%d", time.Seconds())
    tmpdir = filepath.Join(root, tmpstub)
    if global.GetString("-lib") != "" {
        tmplib = filepath.Join(global.GetString("-lib"), tmpstub)
    }

    dir, e1 := os.Stat(tmpdir)

    if e1 == nil && dir.IsDirectory() {
        log.Printf("[ERROR] directory: %s already exists\n", tmpdir)
    } else {
        e_mk := os.Mkdir(tmpdir, 0777)
        if e_mk != nil {
            log.Fatal("[ERROR] failed to create directory for testing")
        }
    }

    tmpfile = filepath.Join(tmpdir, "_main.go")

    fil, e2 := os.OpenFile(tmpfile, os.O_WRONLY|os.O_CREATE, 0777)

    if e2 != nil {
        log.Fatalf("[ERROR] %s\n", e2)
    }

    n, e3 := fil.WriteString(sbTotal.String())

    if e3 != nil {
        log.Fatalf("[ERROR] %s\n", e3)
    } else if n != sbTotal.Len() {
        log.Fatal("[ERROR] failed to write test")
    }

    fil.Close()

    p := newPackage()
    p.Name = filepath.Join(tmpstub, "main")
    p.ShortName = "main"
    p.Files = append(p.Files, tmpfile)

    vec := make([]*Package, 1)
    vec[0] = p
    return vec, tmpdir, tmplib
}

func (d Dag) Topsort() []*Package {

    var node, child *Package
    var cnt int = 0

    zero := make([]*Package, 0)
    done := make([]*Package, 0)

    for _, v := range d {
        if v.Indegree == 0 {
            zero = append(zero, v)
        }
    }

    for len(zero) > 0 {

        node = zero[0]
        zero = zero[1:] // Pop

        for i := 0; i < len(node.children); i++ {
            child = node.children[i]
            child.Indegree--
            if child.Indegree == 0 {
                zero = append(zero, child)
            }
        }
        cnt++
        done = append(done, node)
    }

    if cnt < len(d) {
        log.Fatal("[ERROR] loop in dependency graph\n")
    }

    return done
}

func (d Dag) localDependency(dep string) bool {
    _, ok := d[dep]
    return ok
}

func (d Dag) PrintInfo() {

    var i int

    fmt.Println("--------------------------------------")
    fmt.Println("Packages and Dependencies")
    fmt.Println("p = package, f = file, d = dependency ")
    fmt.Println("--------------------------------------\n")

    for k, v := range d {
        fmt.Println("p ", k)
        for i = 0; i < len(v.Files); i++ {
            fmt.Println("f ", v.Files[i])
        }
        for ds := range v.dependencies.Iter() {
            fmt.Println("d ", ds)
        }
        fmt.Println("")
    }
}

func (p *Package) DotGraph(sb *stringbuffer.StringBuffer) {

    if p.dependencies.Len() == 0 {

        sb.Add(fmt.Sprintf("\t\"%s\";\n", p.Name))

    } else {

        for dep := range p.dependencies.Iter() {
            sb.Add(fmt.Sprintf("\t\"%s\" -> \"%s\";\n", p.Name, dep))
        }
    }
}

// if a package contains test-files and regular files, and on top
// of that contains an 'init' function inside the test-files; we
// have to recompile that package and its recursive dependencies
// to avoid dragging the test-code into the produced binaries and
// libraries that depends on this package. thanks to seth.bunce@gm..
// for reporting this issue.
func (p *Package) HasTestAndInit() (recompile bool) {

    var (
        testFile   bool = false
        plainFile  bool = false
        plainArgv  []string
        plainFiles []string
    )

    p.Indegree = 0

    for y := 0; y < len(p.Files); y++ {
        if strings.HasSuffix(p.Files[y], "_test.go") {
            testFile = true
        } else {
            plainFile = true
        }
    }

    if testFile && plainFile {
        for j := 0; j < len(p.Files); j++ {
            if strings.HasSuffix(p.Files[j], "_test.go") {
                collector := &initCollector{hasInit: false}
                tree := getSyntaxTreeOrDie(p.Files[j], 0)
                ast.Walk(collector, tree)
                if collector.hasInit {
                    recompile = true
                }
            }
        }
    }

    // strip test-files from package during recompile, we
    // 'touch' a file inside the package in order to make godag
    // recompile the package (it won't be up2date any longer)
    if recompile {

        plainArgv = make([]string, 0)
        plainFiles = make([]string, 0)

        for j := 0; j < len(p.Argv); j++ {
            if !strings.HasSuffix(p.Argv[j], "_test.go") {
                plainArgv = append(plainArgv, p.Argv[j])
            }
        }

        for j := 0; j < len(p.Files); j++ {
            if !strings.HasSuffix(p.Files[j], "_test.go") {
                plainFiles = append(plainFiles, p.Files[j])
                handy.Touch(p.Files[j])
            }
        }

        p.Argv = plainArgv
        p.Files = plainFiles
    }

    return recompile
}

func (p *Package) Rep() string {

    sb := make([]string, 0)
    sb = append(sb, "&Package{")
    sb = append(sb, "    name:   \""+p.ShortName+"\",")
    sb = append(sb, "    full:    \""+p.Name+"\",")
    sb = append(sb, "    output: \"_obj/"+p.Name+"\",")

    // special case: build from PWD (srcdir == .)
    files := make([]string, len(p.Files))
    for i := 0; i < len(p.Files); i++ {
        files[i] = p.Files[i]
    }

    pwd, e := os.Getwd()
    if e == nil {
        pwd = pwd + string(filepath.Separator)
        for i := 0; i < len(files); i++ {
            if strings.HasPrefix(files[i], pwd) {
                files[i] = files[i][len(pwd):]
            }
        }
    }

    fs := make([]string, 0)
    for i := 0; i < len(p.Files); i++ {
        fs = append(fs, "\""+filepath.ToSlash(files[i])+"\"")
    }

    sb = append(sb, "    files:  []string{"+strings.Join(fs, ",")+"},")
    sb = append(sb, "},\n")

    for i := 0; i < len(sb); i++ {
        sb[i] = "    " + sb[i]
    }

    return strings.Join(sb, "\n")
}

func (p *Package) UpToDate() bool {

    if p.Argv == nil {
        log.Fatalf("[ERROR] missing dag.Package.Argv\n")
    }

    var e error
    var finfo *os.FileInfo
    var compiledModifiedTime int64
    var last, stop, i int
    var resultingFile string

    last = len(p.Argv) - 1
    stop = last - len(p.Files)
    resultingFile = p.Argv[stop]

    finfo, e = os.Stat(resultingFile)

    if e != nil {
        return false
    } else {
        compiledModifiedTime = finfo.Mtime_ns
    }

    for i = last; i > stop; i-- {
        finfo, e = os.Stat(p.Argv[i])
        if e != nil {
            panic(fmt.Sprintf("Missing go file: %s\n", p.Argv[i]))
        } else {
            if finfo.Mtime_ns > compiledModifiedTime {
                return false
            }
        }
    }

    // package contains _test.go and -test => not UpToDate
    if global.GetBool("-test") {
        testpkgs := 0
        for i = 0; i < len(p.Files); i++ {
            if strings.HasSuffix(p.Files[i], "_test.go") {
                testpkgs++
            }
        }
        if testpkgs > 0 && testpkgs != len(p.Files) {
            return false
        }
    }

    return true
}

func (p *Package) Ready(local, compiled *stringset.StringSet) bool {

    for dep := range p.dependencies.Iter() {
        if local.Contains(dep) && !compiled.Contains(dep) {
            return false
        }
    }

    return true
}

func (p *Package) ResetIndegree() {
    for i := 0; i < len(p.children); i++ {
        p.children[i].Indegree++
    }
}

func (p *Package) InitWaitGroup() {
    p.waiter = new(sync.WaitGroup)
    p.waiter.Add(p.Indegree)
}

func (p *Package) Decrement(compile bool) {
    p.lock.Lock()
    p.needsCompile = compile
    p.waiter.Done()
    p.lock.Unlock()
}

func (p *Package) Compile(ch chan int) {

    var doCompile bool

    p.waiter.Wait()

    if p.needsCompile || !p.UpToDate() {
        oldPkgIsFound()
        doCompile = true
    } else {
        say.Printf("up 2 date: %s\n", p.Name)
    }
    if doCompile {
        say.Printf("compiling: %s\n", p.Name)
        handy.StdExecve(p.Argv, true)
    }
    for _, child := range p.children {
        child.Decrement(doCompile)
    }
    ch <- 1
}

func (p *Package) Visit(node ast.Node) (v ast.Visitor) {

    switch node.(type) {
    case *ast.BasicLit:
        bl, ok := node.(*ast.BasicLit)
        if ok {
            stripped := string(bl.Value[1 : len(bl.Value)-1])
            p.dependencies.Add(stripped)
        }
    default: // nothing to do if not BasicLit
    }
    return p
}

//TODO make this examples stuff work, if someone asks for it..
//TODO check that types are ok as well..
func (t *TestCollector) Visit(node ast.Node) (v ast.Visitor) {
    switch fn := node.(type) {
    case *ast.FuncDecl:

        if fn.Recv == nil { // node is a function
            if strings.HasPrefix(fn.Name.Name, "Test") {
                if fn.Type.Params != nil && fn.Type.Params.NumFields() == 1 {
                    t.TestFuncs = append(t.TestFuncs, fn.Name.Name)
                }
            }
            if strings.HasPrefix(fn.Name.Name, "Benchmark") {
                if fn.Type.Params != nil && fn.Type.Params.NumFields() == 1 {
                    t.BenchFuncs = append(t.BenchFuncs, fn.Name.Name)
                }
            }
            if strings.HasPrefix(fn.Name.Name, "Example") {
                if fn.Type.Params != nil && fn.Type.Params.NumFields() == 0 {
                    t.ExampleFuncs = append(t.ExampleFuncs, fn.Name.Name)
                }
            }
        }

///     case *ast.Comment: fmt.Printf("Comment: %s\n", fn.Text)

    default: // nothing to do if not FuncDecl,Comment
    }
    return t
}

func (t *TestCollector) String() string {
    return fmt.Sprintf("&TestCollector{\n\tt: %v\n\tb: %v\n\te: %v\n}\n",
        t.TestFuncs, t.BenchFuncs, t.ExampleFuncs)
}

func (t *TestCollector) FoundAnything() bool {
    tot := len(t.TestFuncs) + len(t.BenchFuncs) + len(t.ExampleFuncs)
    return tot > 0
}

func (i *initCollector) Visit(node ast.Node) (v ast.Visitor) {
    switch t := node.(type) {
    case *ast.FuncDecl:
        if t.Name.Name == "init" {
            if (t.Type.Params == nil || t.Type.Params.NumFields() == 0) &&
                (t.Type.Results == nil || t.Type.Results.NumFields() == 0) {
                i.hasInit = true
            }
        }
    default: // nothing to do if not FuncDecl
    }
    return i
}

func addSeparatorPath(root string) string {
    if root[len(root)-1:] != "/" {
        root = root + "/"
    }
    return root
}

func hasSlash(s string) bool {
    return strings.Index(s, "/") != -1
}

func removeSlashAndDot(s string) string {
    noslash := strings.Replace(s, "/", "", -1)
    return strings.Replace(noslash, ".", "", -1)
}

func getSyntaxTreeOrDie(file string, mode uint) *ast.File {
    absSynTree, err := parser.ParseFile(token.NewFileSet(), file, nil, mode)
    if err != nil {
        log.Fatalf("%s\n", err)
    }
    return absSynTree
}

func OldPkgYet() (res bool) {
    locker.Lock()
    res = oldPkgFound
    locker.Unlock()
    return res
}

func oldPkgIsFound() {
    locker.Lock()
    oldPkgFound = true
    locker.Unlock()
}
