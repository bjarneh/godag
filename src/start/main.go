// © Knug Industries 2009 all rights reserved
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package main

import (
    "os"
    "fmt"
    "path"
    "log"
    "strings"
    "io/ioutil"
    "regexp"
    "runtime"
    "exec"
    "utilz/walker"
    "cmplr/compiler"
    "cmplr/dag"
    "parse/gopt"
    "utilz/handy"
    "container/vector"
    "utilz/global"
)


// option parser object (struct)
var getopt *gopt.GetOpt

// list of files to compile
var files *vector.StringVector

// libraries other than $GOROOT/pkg/PLATFORM
var includes []string = nil

// source root
var srcdir string = "src"


// keys for the bool options
var bools = []string{
    "-help",
    "-clean",
    "-static",
    "-version",
    "-sort",
    "-print",
    "-dryrun",
    "-test",
    "-list",
    "-verbose",
    "-fmt",
    "-no-comments",
    "-tab",
    "-external",
}

// keys for the string options
// note: -I is handled seperately
var strs  = []string{
    "-arch",
    "-dot",
    "-tabwidth",
    "-rew-rule",
    "-output",
    "-benchmarks",
    "-match",
    "-test-bin",
}


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
    getopt.BoolOption("-e -external --external")
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

    for _, bkey := range bools {
        global.SetBool(bkey, false)
    }

    for _, skey := range strs {
        global.SetString(skey, "")
    }

    global.SetString("-test-bin", "gdtest")
    global.SetString("-I", "")

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
    if global.GetBool("-list") {
        printListing()
        os.Exit(0)
    }

    if global.GetBool("-help") {
        printHelp()
        os.Exit(0)
    }

    if global.GetBool("-version") {
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
    if global.GetBool("-clean") {
        rm865(srcdir)
        os.Exit(0)
    }

    files = findFiles(srcdir)

    // gofmt on all files gathered
    if global.GetBool("-fmt") {
        formatFiles(files)
        os.Exit(0)
    }

    // parse the source code, look for dependencies
    dgrph := dag.New()
    dgrph.Parse(srcdir, files)

    // print collected dependency info
    if global.GetBool("-print") {
        dgrph.PrintInfo()
        os.Exit(0)
    }

    // draw graphviz dot graph
    if global.GetString("-dot") != "" {
        dgrph.MakeDotGraph( global.GetString("-dot"))
        os.Exit(0)
    }

    gotRoot() //?

    // build all external dependencies
    if global.GetBool("-external") {
        dgrph.External()
    }

    // sort graph based on dependencies
    dgrph.GraphBuilder()
    sorted := dgrph.Topsort()

    // print packages sorted
    if global.GetBool("-sort") {
        for i := 0; i < sorted.Len(); i++ {
            rpkg, _ := sorted.At(i).(*dag.Package)
            fmt.Printf("%s\n", rpkg.Name)
        }
        os.Exit(0)
    }

    // compile
    kompiler := compiler.New(srcdir, includes)
    kompiler.CreateArgv(sorted)
    if runtime.GOMAXPROCS(-1) > 1 && ! global.GetBool("-dryrun") {
        kompiler.ParallelCompile(sorted)
    } else {
        kompiler.SerialCompile(sorted)
    }

    // test
    if global.GetBool("-test") {
        os.Setenv("SRCROOT", srcdir)
        testMain, testDir := dgrph.MakeMainTest(srcdir)
        kompiler.CreateArgv(testMain)
        kompiler.SerialCompile(testMain)
        kompiler.ForkLink(testMain, global.GetString("-test-bin"))
        kompiler.DeletePackages(testMain)
        rmError := os.Remove(testDir)
        if rmError != nil {
            log.Printf("[ERROR] failed to remove testdir: %s\n", testDir)
        }
        testArgv := createTestArgv()
        if ! global.GetBool("-dryrun") {
            tstring := "testing  : "
            if global.GetBool("-verbose") {
                tstring += "\n"
            }
            fmt.Printf(tstring)
            ok = handy.StdExecve(testArgv, false)
            e = os.Remove(global.GetString("-test-bin"))
            if e != nil {
                log.Printf("[ERROR] %s\n", e)
            }
            if !ok {
                os.Exit(1)
            }
        }

    }

    if global.GetString("-output") != "" {
        kompiler.ForkLink(sorted, global.GetString("-output"))
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

    for _, bkey := range bools {
        if getopt.IsSet(bkey) {
            global.SetBool(bkey, true)
        }
    }

    for _, skey := range strs {
        if getopt.IsSet(skey) {
            global.SetString(skey, getopt.Get(skey))
        }
    }

    if getopt.IsSet("-test") || getopt.IsSet("-fmt") {
        findTestFilesAlso()
    }

    if getopt.IsSet("-I") {
        if includes == nil {
            includes = getopt.GetMultiple("-I")
        } else {
            includes = append(includes, getopt.GetMultiple("-I")...)
        }
    }

    getopt.Reset()
    return args
}

func createTestArgv() []string {

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

func formatFiles(files *vector.StringVector) {

    var i int = 0
    var argvLen int = 0
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

func rm865(srcdir string) {

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
  -V --verbose         verbose unit-test and goinstall
  --test-bin           name of test-binary (default: gdtest)
  -f --fmt             run gofmt on src and exit
  --rew-rule           pass rewrite rule to gofmt
  --tab                pass -tabindent=true to gofmt
  --tabwidth           pass -tabwidth to gofmt (default:4)
  --no-comments        pass -comments=false to gofmt
  -e --external        goinstall all external dependencies
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
  -e --external        =>   %t

`
    tabRepr := "4"
    if global.GetString("-tabwidth") != "" {
        tabRepr = global.GetString("-tabwidth")
    }

    archRepr := "$GOARCH"
    if global.GetString("-arch") != "" {
        archRepr = global.GetString("-arch")
    }

    fmt.Printf(listMSG,
               global.GetBool("-help"),
               global.GetBool("-version"),
               global.GetBool("-print"),
               global.GetBool("-sort"),
               global.GetString("-output"),
               global.GetBool("-static"),
               archRepr,
               global.GetBool("-dryrun"),
               global.GetBool("-clean"),
               includes,
               global.GetString("-dot"),
               global.GetBool("-test"),
               global.GetString("-benchmarks"),
               global.GetString("-match"),
               global.GetBool("-verbose"),
               global.GetString("-test-bin"),
               global.GetBool("-fmt"),
               global.GetString("-rew-rule"),
               global.GetBool("-tab"),
               tabRepr,
               global.GetBool("-no-comments"),
               global.GetBool("-external"))
}
