// © Knug Industries 2009 all rights reserved
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package main

import (
    "os"
    "fmt"
    "path"
    "log"
    "utilz/walker"
    "cmplr/compiler"
    "cmplr/dag"
    "container/vector"
    "parse/gopt"
    "strings"
    "utilz/handy"
    "io/ioutil"
    "regexp"
    "runtime"
    "exec"
)


// option parser object (struct)
var getopt *gopt.GetOpt

// list of files to compile
var files *vector.StringVector

// libraries other than $GOROOT/pkg/PLATFORM
var includes []string = nil

// variables for the string options
var (
    dot      = ""
    arch     = ""
    gdtest   = "gdtest"
    output   = ""
    srcdir   = "src"
    bmatch   = ""
    match    = ""
    rewRule  = ""
    tabWidth = ""
)

// variables for bool options
var (
    dryrun       = false
    test         = false
    testVerbose  = false
    static       = false
    noComments   = false
    tabIndent    = false
    listing      = false
    gofmt        = false
    printInfo    = false
    sortInfo     = false
    cleanTree    = false
    needsHelp    = false
    needsVersion = false
)

func init() {

    // initialize option parser
    getopt = gopt.New()

    // add all options (bool/string)
    getopt.BoolOption("-h -help --help help")
    getopt.BoolOption("-c -clean --clean clean")
    getopt.BoolOption("-S -static --static")
    getopt.BoolOption("-v -version --version version")
    getopt.BoolOption("-s -sort --sort sort")
    getopt.BoolOption("-p -print --print")
    getopt.BoolOption("-d -dryrun --dryrun")
    getopt.BoolOption("-t -test --test")
    getopt.BoolOption("-l -list --list")
    getopt.BoolOption("-V -verbose --verbose")
    getopt.BoolOption("-f -fmt --fmt")
    getopt.BoolOption("-no-comments --no-comments")
    getopt.BoolOption("-tab --tab")
    getopt.StringOption("-a -a= -arch --arch -arch= --arch=")
    getopt.StringOption("-dot -dot= --dot --dot=")
    getopt.StringOption("-I -I=")
    getopt.StringOption("-tabwidth --tabwidth -tabwidth= --tabwidth=")
    getopt.StringOption("-rew-rule --rew-rule -rew-rule= --rew-rule=")
    getopt.StringOption("-o -o= -output --output -output= --output=")
    getopt.StringOption("-b -b= -benchmarks --benchmarks -benchmarks= --benchmarks=")
    getopt.StringOption("-m -m= -match --match -match= --match=")
    getopt.StringOption("-test-bin --test-bin -test-bin= --test-bin=")

    // override IncludeFile to make walker pick up only .go files
    walker.IncludeFile = func(s string) bool {
        return strings.HasSuffix(s, ".go") &&
            !strings.HasSuffix(s, "_test.go")
    }

    // override IncludeDir to make walker ignore 'hidden' directories
    walker.IncludeDir = func(s string) bool {
        _, dirname := path.Split(s)
        return dirname[0] != '.'
    }

}

func gotRoot() {
    if os.Getenv("GOROOT") == "" {
        log.Exit("[ERROR] missing GOROOT\n")
    }
}

func findTestFilesAlso() {
    // override IncludeFile to make walker pick _test.go files also
    walker.IncludeFile = func(s string) bool {
        return strings.HasSuffix(s, ".go")
    }
}

func main() {

    var ok bool
    var e os.Error
    var argv, args []string

    // default config location 1 $HOME/.gdrc
    argv, ok = getConfigArgv("HOME")

    if ok {
        args = parseArgv(argv)
        if len(args) > 0 {
            log.Print("[WARNING] non-option arguments in config file\n")
        }
    }

    // default config location 2 $PWD/.gdrc
    argv, ok = getConfigArgv("PWD")

    if ok {
        args = parseArgv(argv)
        if len(args) > 0 {
            log.Print("[WARNING] non-option arguments in config file\n")
        }
    }

    // command line arguments overrides/appends config
    args = parseArgv(os.Args[1:])

    if len(args) > 0 {
        if len(args) > 1 {
            log.Print("[WARNING] len(input directories) > 1\n")
        }
        srcdir = args[0]
        if srcdir == "." {
            srcdir, e = os.Getwd()
            if e != nil {
                srcdir = os.Getenv("PWD")
                if srcdir == "" {
                    log.Exit("[ERROR] can't find working directory\n")
                }
            }
        }

    }


    // stuff that can be done without $GOROOT
    if needsHelp {
        printHelp()
        os.Exit(0)
    }

    if listing {
        printListing()
        os.Exit(0)
    }

    if needsVersion {
        printVersion()
        os.Exit(0)
    }

    if len(args) == 0 {
        // give nice feedback if missing input dir
        possibleSrc := path.Join(os.Getenv("PWD"), "src")
        _, e = os.Stat(possibleSrc)
        if e != nil {
            fmt.Printf("usage: gd [OPTIONS] src-directory\n")
            os.Exit(1)
        }
    }


    // delete all object/archive files
    if cleanTree {
        rm865(srcdir, dryrun)
        os.Exit(0)
    }

    files = findFiles(srcdir)

    // gofmt on all files gathered
    if gofmt {
        formatFiles(files, dryrun, tabIndent, noComments, rewRule, tabWidth)
        os.Exit(0)
    }

    // parse the source code, look for dependencies
    dgrph := dag.New()
    dgrph.Parse(srcdir, files)

    // print collected dependency info
    if printInfo {
        dgrph.PrintInfo()
        os.Exit(0)
    }

    // draw graphviz dot graph
    if dot != "" {
        dgrph.MakeDotGraph(dot)
        os.Exit(0)
    }

    gotRoot() //?

    // sort graph based on dependencies
    dgrph.GraphBuilder(includes)
    sorted := dgrph.Topsort()

    // print packages sorted
    if sortInfo {
        for i := 0; i < sorted.Len(); i++ {
            rpkg, _ := sorted.At(i).(*dag.Package)
            fmt.Printf("%s\n", rpkg.Name)
        }
        os.Exit(0)
    }

    // compile
    kompiler := compiler.New(srcdir, arch, dryrun, includes)
    kompiler.CreateArgv(sorted)
    if runtime.GOMAXPROCS(-1) > 1 && !dryrun {
        kompiler.ParallelCompile(sorted)
    } else {
        kompiler.SerialCompile(sorted)
    }

    // test
    if test {
        os.Setenv("SRCROOT", srcdir)
        testMain, testDir := dgrph.MakeMainTest(srcdir)
        kompiler.CreateArgv(testMain)
        kompiler.SerialCompile(testMain)
        kompiler.ForkLink(testMain, gdtest, false)
        kompiler.DeletePackages(testMain)
        rmError := os.Remove(testDir)
        if rmError != nil {
            log.Printf("[ERROR] failed to remove testdir: %s\n", testDir)
        }
        testArgv := createTestArgv(gdtest, bmatch, match, testVerbose)
        if !dryrun {
            tstring := "testing  : "
            if testVerbose {
                tstring += "\n"
            }
            fmt.Printf(tstring)
            ok = handy.StdExecve(testArgv, false)
            e = os.Remove(gdtest)
            if e != nil {
                log.Printf("[ERROR] %s\n", e)
            }
            if !ok {
                os.Exit(1)
            }
        }

    }

    if output != "" {
        kompiler.ForkLink(sorted, output, static)
    }

}


// syntax for config files are identical to command line
// options, i.e., write command line options to the config file 
// and everything should work, comments start with a '#' sign.
// config files are either $HOME/.gdrc or $PWD/.gdrc
func getConfigArgv(where string) (argv []string, ok bool) {

    location := os.Getenv(where)

    if location == "" {
        return nil, false
    }

    configFile := path.Join(location, ".gdrc")
    configDir, e := os.Stat(configFile)

    if e != nil {
        return nil, false
    }

    if !configDir.IsRegular() {
        return nil, false
    }

    b, e := ioutil.ReadFile(configFile)

    if e != nil {
        log.Print("[WARNING] failed to read config file\n")
        log.Printf("[WARNING] %s \n", e)
        return nil, false
    }

    comStripRegex := regexp.MustCompile("#[^\n]*\n?")
    blankRegex := regexp.MustCompile("[\n\t \r]+")

    rmComments := comStripRegex.ReplaceAllString(string(b), "")
    rmNewLine := blankRegex.ReplaceAllString(rmComments, " ")

    pureOptions := strings.TrimSpace(rmNewLine)

    if pureOptions == "" {
        return nil, false
    }

    argv = strings.Split(pureOptions, " ", -1)

    return argv, true
}


func parseArgv(argv []string) (args []string) {

    args = getopt.Parse(argv)

    if getopt.IsSet("-help") {
        needsHelp = true
    }

    if getopt.IsSet("-list") {
        listing = true
    }

    if getopt.IsSet("-version") {
        needsVersion = true
    }

    if getopt.IsSet("-dryrun") {
        dryrun = true
    }

    if getopt.IsSet("-print") {
        printInfo = true
    }

    if getopt.IsSet("-sort") {
        sortInfo = true
    }

    if getopt.IsSet("-static") {
        static = true
    }

    if getopt.IsSet("-clean") {
        cleanTree = true
    }

    if getopt.IsSet("-arch") {
        arch = getopt.Get("-a")
    }

    if getopt.IsSet("-dot") {
        dot = getopt.Get("-dot")
    }

    if getopt.IsSet("-I") {
        if includes == nil {
            includes = getopt.GetMultiple("-I")
        } else {
            tmp := getopt.GetMultiple("-I")
            joined := make([]string, (len(includes) + len(tmp)))

            var i, j int

            for i = 0; i < len(includes); i++ {
                joined[i] = includes[i]
            }
            for j = 0; j < len(tmp); j++ {
                joined[i+j] = tmp[j]
            }

            includes = joined
        }
    }

    if getopt.IsSet("-output") {
        output = getopt.Get("-o")
    }

    // for gotest
    if getopt.IsSet("-test") {
        test = true
        findTestFilesAlso()
    }

    if getopt.IsSet("-benchmarks") {
        bmatch = getopt.Get("-b")
    }

    if getopt.IsSet("-match") {
        match = getopt.Get("-m")
    }

    if getopt.IsSet("-verbose") {
        testVerbose = true
    }

    if getopt.IsSet("-test-bin") {
        gdtest = getopt.Get("-test-bin")
    }

    // for gofmt
    if getopt.IsSet("-fmt") {
        gofmt = true
    }

    if getopt.IsSet("-no-comments") {
        noComments = true
    }

    if getopt.IsSet("-rew-rule") {
        rewRule = getopt.Get("-rew-rule")
    }

    if getopt.IsSet("-tab") {
        tabIndent = true
    }

    if getopt.IsSet("-tabwidth") {
        tabWidth = getopt.Get("-tabwidth")
    }

    getopt.Reset()
    return args
}

func createTestArgv(prg, bmatch, match string, tverb bool) []string {
    var numArgs int = 1
    pwd, e := os.Getwd()
    if e != nil {
        log.Exit("[ERROR] could not locate working directory\n")
    }
    arg0 := path.Join(pwd, prg)
    if bmatch != "" {
        numArgs += 2
    }
    if match != "" {
        numArgs += 2
    }
    if tverb {
        numArgs++
    }

    var i = 1
    argv := make([]string, numArgs)
    argv[0] = arg0
    if bmatch != "" {
        argv[i] = "-benchmarks"
        i++
        argv[i] = bmatch
        i++
    }
    if match != "" {
        argv[i] = "-match"
        i++
        argv[i] = match
        i++
    }
    if tverb {
        argv[i] = "-v"
    }
    return argv
}

func findFiles(pathname string) *vector.StringVector {
    okDirOrDie(pathname)
    return walker.PathWalk(path.Clean(pathname))
}

func okDirOrDie(pathname string) {

    var dir *os.FileInfo
    var staterr os.Error

    dir, staterr = os.Stat(pathname)

    if staterr != nil {
        log.Exitf("[ERROR] %s\n", staterr)
    } else if !dir.IsDirectory() {
        log.Exitf("[ERROR] %s: is not a directory\n", pathname)
    }
}

func formatFiles(files *vector.StringVector, dryrun, tab, noC bool, rew, tw string) {

    var i int = 0
    var argvLen int = 0
    var argv []string
    var tabWidth string = "-tabwidth=4"
    var useTabs string = "-tabindent=false"
    var comments string = "-comments=true"
    var rewRule string = ""
    var fmtexec string
    var err os.Error

    fmtexec, err = exec.LookPath("gofmt")

    if err != nil {
        log.Exit("[ERROR] could not find 'gofmt' in $PATH")
    }

    if tw != "" {
        tabWidth = "-tabwidth=" + tw
    }
    if noC {
        comments = "-comments=false"
    }
    if rew != "" {
        rewRule = rew
        argvLen++
    }
    if tab {
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
        argv[i] = "-r=" + rewRule
        i++
    }

    for y := 0; y < files.Len(); y++ {
        argv[i] = files.At(y)
        if !dryrun {
            fmt.Printf("gofmt : %s\n", files.At(y))
            _ = handy.StdExecve(argv, true)
        } else {
            fmt.Printf(" %s\n", strings.Join(argv, " "))
        }
    }

}

func rm865(srcdir string, dryrun bool) {

    // override IncludeFile to make walker pick up only .[865] files
    walker.IncludeFile = func(s string) bool {
        return strings.HasSuffix(s, ".8") ||
            strings.HasSuffix(s, ".6") ||
            strings.HasSuffix(s, ".5") ||
            strings.HasSuffix(s, ".a")

    }

    okDirOrDie(srcdir)

    compiled := walker.PathWalk(path.Clean(srcdir))

    for i := 0; i < compiled.Len(); i++ {

        if !dryrun {

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

func printHelp() {
    var helpMSG string = `
  Godag is a compiler front-end for golang,
  its main purpose is to help build projects
  which are pure Go-code without Makefiles.
  Hopefully it simplifies testing as well.

  usage: gd [OPTIONS] src-directory

  options:

  -h --help            print this message and quit
  -v --version         print version and quit
  -l --list            list option values and quit
  -p --print           print package info collected
  -s --sort            print legal compile order
  -o --output          link main package -> output
  -S --static          statically link binary
  -a --arch            architecture (amd64,arm,386)
  -d --dryrun          print what gd would do (stdout)
  -c --clean           rm *.[a865] from src-directory
  -dot                 create a graphviz dot file
  -I                   import package directories
  -t --test            run all unit-tests
  -b --benchmarks      pass argument to unit-test
  -m --match           pass argument to unit-test
  -V --verbose         pass argument '-v' to unit-test
  --test-bin           name of test-binary (default: gdtest)
  -f --fmt             run gofmt on src and exit
  --rew-rule           pass rewrite rule to gofmt
  --tab                pass -tabindent=true to gofmt
  --tabwidth           pass -tabwidth to gofmt (default:4)
  --no-comments        pass -comments=false to gofmt
    `

    fmt.Println(helpMSG)
}

func printVersion() {
    fmt.Println("godag 0.1")
}

func printListing() {
    var listMSG string = `
  Listing of options and their content:

  -h --help            =>   %t
  -v --version         =>   %t
  -p --print           =>   %t
  -s --sort            =>   %t
  -o --output          =>   '%s'
  -S --static          =>   %t
  -a --arch            =>   %v
  -d --dryrun          =>   %t
  -c --clean           =>   %t
  -I                   =>   %v
  -dot                 =>   '%s'
  -t --test            =>   %t
  -b --benchmarks      =>   '%s'
  -m --match           =>   '%s'
  -V --verbose         =>   %t
  --test-bin           =>   '%s'
  -f --fmt             =>   %t
  --rew-rule           =>   '%s'
  --tab                =>   %t
  --tabwidth           =>   %s
  --no-comments        =>   %t

`
    tabRepr := "4"
    if tabWidth != "" {
        tabRepr = tabWidth
    }

    archRepr := "$GOARCH"
    if arch != "" {
        archRepr = arch
    }

    fmt.Printf(listMSG, needsHelp, needsVersion, printInfo,
        sortInfo, output, static, archRepr, dryrun, cleanTree,
        includes, dot, test, bmatch, match, testVerbose, gdtest,
        gofmt, rewRule, tabIndent, tabRepr, noComments)
}
