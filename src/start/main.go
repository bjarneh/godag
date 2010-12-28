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
    "runtime"
    "utilz/walker"
    "cmplr/compiler"
    "cmplr/dag"
    "parse/gopt"
    "utilz/handy"
    "utilz/global"
)


// option parser object (struct)
var getopt *gopt.GetOpt

// list of files to compile
var files []string

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


func main() {

    var ok bool
    var e os.Error
    var argv, args []string
    var config1, config2 string

    // default config location 1 $HOME/.gdrc
    config1 = path.Join(os.Getenv("HOME"), ".gdrc")
    argv, ok = handy.ConfigToArgv(config1)

    if ok {
        args = parseArgv(argv)
        if len(args) > 0 {
            log.Print("[WARNING] non-option arguments in config file\n")
        }
    }

    // default config location 2 $PWD/.gdrc
    config2 = path.Join(os.Getenv("PWD"), ".gdrc")
    argv, ok = handy.ConfigToArgv(config2)

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
        cwd, e := os.Getwd()
        if e != nil {
            cwd = os.Getenv("PWD")
        }
        possibleSrc := path.Join(cwd, "src")
        _, e = os.Stat(possibleSrc)
        if e != nil {
            fmt.Printf("usage: gd [OPTIONS] src-directory\n")
            os.Exit(1)
        }
    }


    // delete all object/archive files
    if global.GetBool("-clean") {
        compiler.Remove865a(srcdir)
        os.Exit(0)
    }

    handy.DirOrExit(srcdir)
    files = walker.PathWalk(path.Clean(srcdir))

    // gofmt on all files gathered
    if global.GetBool("-fmt") {
        compiler.FormatFiles(files)
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
        for i := 0; i < len(sorted); i++ {
            fmt.Printf("%s\n", sorted[i].Name)
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
        testArgv := compiler.CreateTestArgv()
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
        // override IncludeFile to make walker pick _test.go files
        walker.IncludeFile = func(s string) bool {
            return strings.HasSuffix(s, ".go")
        }
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
