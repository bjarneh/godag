// © Knug Industries 2009 all rights reserved
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package main

import (
    "os"
    "fmt"
    "log"
    "strings"
    "path/filepath"
    "utilz/walker"
    "cmplr/compiler"
    "cmplr/dag"
    "cmplr/gdmake"
    "parse/gopt"
    "utilz/handy"
    "utilz/global"
    "utilz/timer"
    "utilz/say"
)

// option parser object (struct)
var getopt *gopt.GetOpt

// list of files to compile
var files []string

// libraries other than $GOROOT/pkg/PLATFORM
var includes []string = make([]string, 0)

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
    "-quiet",
    "-tab",
    "-external",
}

// keys for the string options
// note: -I is handled seperately
var strs = []string{
    "-arch",
    "-dot",
    "-tabwidth",
    "-rewrite",
    "-output",
    "-bench",
    "-match",
    "-test-bin",
    "-lib",
    "-main",
    "-backend",
    "-gdmk",
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
    getopt.BoolOption("-t -test --test test")
    getopt.BoolOption("-l -list --list")
    getopt.BoolOption("-q -quiet --quiet")
    getopt.BoolOption("-V -verbose --verbose")
    getopt.BoolOption("-f -fmt --fmt")
    getopt.BoolOption("-T -tab --tab")
    getopt.BoolOption("-e -external --external")
    getopt.StringOption("-a -a= -arch --arch -arch= --arch=")
    getopt.StringOption("-D -D= -dot -dot= --dot --dot=")
    getopt.StringOption("-L -L= -lib -lib= --lib --lib=")
    getopt.StringOption("-I -I=")
    getopt.StringOption("-g -g= -gdmk -gdmk= --gdmk --gdmk=")
    getopt.StringOption("-w -w= -tabwidth --tabwidth -tabwidth= --tabwidth=")
    getopt.StringOption("-r -r= -rewrite --rewrite -rewrite= --rewrite=")
    getopt.StringOption("-o -o= -output --output -output= --output=")
    getopt.StringOption("-M -M= -main --main -main= --main=")
    getopt.StringOption("-b -b= -bench --bench -bench= --bench=")
    getopt.StringOption("-m -m= -match --match -match= --match=")
    getopt.StringOption("-test-bin --test-bin -test-bin= --test-bin=")
    getopt.StringOption("-B -B= -backend --backend -backend= --backend=")

    // override IncludeFile to make walker pick up only .go files
    walker.IncludeFile = noTestFilesFilter

    // override IncludeDir to make walker ignore 'hidden' directories
    walker.IncludeDir = func(s string) bool {
        _, dirname := filepath.Split(s)
        return dirname[0] != '.'
    }

    for _, bkey := range bools {
        global.SetBool(bkey, false)
    }

    for _, skey := range strs {
        global.SetString(skey, "")
    }

    // Testing on Windows requires .exe ending
    if os.Getenv("GOOS") == "windows" {
        global.SetString("-test-bin", "gdtest.exe")
    } else {
        global.SetString("-test-bin", "gdtest")
    }

    global.SetString("-backend", "gc")
    global.SetString("-I", "")

}

// utility func for walker: *.go unless start = '_' || end = _test.go
func noTestFilesFilter(s string) bool {
    return strings.HasSuffix(s, ".go") &&
        !strings.HasSuffix(s, "_test.go") &&
        !strings.HasPrefix(filepath.Base(s), "_")
}

// utility func for walker: *.go unless start = '_'
func allGoFilesFilter(s string) bool {
    return strings.HasSuffix(s, ".go") &&
        !strings.HasPrefix(filepath.Base(s), "_")
}

// ignore GOROOT for gccgo and express
func gotRoot() {
    if global.GetString("-backend") == "gc" {
        if os.Getenv("GOROOT") == "" {
            log.Fatal("[ERROR] missing GOROOT\n")
        }
    }
}

func reportTime() {
    timer.Stop("everything")
    delta, _ := timer.Delta("everything")
    say.Printf("time used: %s\n", timer.Nano2Time(delta))
}

func main() {

    var ok, up2date bool
    var e os.Error
    var argv, args []string
    var config1, config2 string

    timer.Start("everything")
    defer reportTime()

    // default config location 1 $HOME/.gdrc
    config1 = filepath.Join(os.Getenv("HOME"), ".gdrc")
    argv, ok = handy.ConfigToArgv(config1)

    if ok {
        args = parseArgv(argv)
        if len(args) > 0 {
            log.Print("[WARNING] non-option arguments in config file\n")
        }
    }

    // default config location 2 $PWD/.gdrc
    config2 = filepath.Join(os.Getenv("PWD"), ".gdrc")
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
                log.Fatal("[ERROR] can't find working directory\n")
            }
        }
    }

    // expand variables in includes
    for i := 0; i < len(includes); i++ {
        includes[i] = os.ShellExpand(includes[i])
    }

    // expand variables in -lib
    global.SetString("-lib", os.ShellExpand(global.GetString("-lib")))

    // expand variables in -output
    global.SetString("-output", os.ShellExpand(global.GetString("-output")))

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
        possibleSrc := filepath.Join(cwd, "src")
        _, e = os.Stat(possibleSrc)
        if e != nil {
            fmt.Printf("usage: gd [OPTIONS] src-directory\n")
            os.Exit(1)
        }
    }

    if global.GetBool("-quiet") {
        say.Mute()
    }

    handy.DirOrExit(srcdir)
    files = walker.PathWalk(filepath.Clean(srcdir))

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
        dgrph.MakeDotGraph(global.GetString("-dot"))
        os.Exit(0)
    }

    gotRoot() //? (only matters to gc, gccgo and express ignores it)

    // build &| update all external dependencies
    if global.GetBool("-external") {
        dgrph.External()
        os.Exit(0)
    }

    // sort graph based on dependencies
    dgrph.GraphBuilder()
    sorted := dgrph.Topsort()

    // clean only what we possibly could have generated…
    if global.GetBool("-clean") {
        compiler.DeleteObjects(srcdir, sorted)
        os.Exit(0)
    }

    // print packages sorted
    if global.GetBool("-sort") {
        for i := 0; i < len(sorted); i++ {
            fmt.Printf("%s\n", sorted[i].Name)
        }
        os.Exit(0)
    }

    // compile argv
    compiler.Init(srcdir, includes)
    if global.GetString("-lib") != "" {
        compiler.CreateLibArgv(sorted)
    } else {
        compiler.CreateArgv(sorted)
    }

    // gdmk
    if global.GetString("-gdmk") != "" {
        gdmake.Make(global.GetString("-gdmk"), sorted, dgrph.Alien().Slice())
        os.Exit(0)
    }

    // compile; up2date == true => 0 packages modified
    if global.GetBool("-dryrun") {
        compiler.Dryrun(sorted)
    } else {
        up2date = compiler.Compile(sorted) // updated parallel
    }

    // test
    if global.GetBool("-test") {
        os.Setenv("SRCROOT", srcdir)
        testMain, testDir, testLib := dgrph.MakeMainTest(srcdir)
        if global.GetString("-lib") != "" {
            compiler.CreateLibArgv(testMain)
        } else {
            compiler.CreateArgv(testMain)
        }
        if !global.GetBool("-dryrun") {
            compiler.Compile(testMain)
        }
        switch global.GetString("-backend") {
        case "gc", "express":
            compiler.ForkLink(global.GetString("-test-bin"), testMain, nil, false)
        case "gccgo", "gcc":
            compiler.ForkLink(global.GetString("-test-bin"), testMain, sorted, false)
        default:
            log.Fatalf("[ERROR] '%s' unknown back-end\n", global.GetString("-backend"))
        }
        compiler.DeletePackages(testMain)
        handy.Delete(testDir, false)
        if testLib != "" {
            handy.Delete(testLib, false)
        }
        testArgv := compiler.CreateTestArgv()
        if global.GetBool("-dryrun") {
            testArgv[0] = filepath.Base(testArgv[0])
            say.Printf("%s\n", strings.Join(testArgv, " "))
        } else {
            say.Printf("testing  : ")
            if global.GetBool("-verbose") {
                say.Printf("\n")
            }
            ok = handy.StdExecve(testArgv, false)
            handy.Delete(global.GetString("-test-bin"), false)
            if !ok {
                os.Exit(1)
            }
        }
    }

    // link if ! up2date
    if global.GetString("-output") != "" {
        compiler.ForkLink(global.GetString("-output"), sorted, nil, up2date)
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

    if getopt.IsSet("-test") || getopt.IsSet("-fmt") || getopt.IsSet("-clean") {
        // override IncludeFile to make walker pick _test.go files
        walker.IncludeFile = allGoFilesFilter
    }

    if getopt.IsSet("-gdmk") {
        global.SetString("-lib", "_obj")
        // gdmk does not support testing
        walker.IncludeFile = noTestFilesFilter
    }

    if getopt.IsSet("-I") {
        includes = append(includes, getopt.GetMultiple("-I")...)
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
  -g --gdmk            create a go makefile for project
  -d --dryrun          print what gd would do (stdout)
  -c --clean           delete generated object code
  -q --quiet           silent, print only errors
  -L --lib             write objects to other dir (!src)
  -M --main            regex to select main package
  -D --dot             create a graphviz dot file
  -I                   import package directories
  -t --test            run all unit-tests
  -m --match           regex to select unit-tests
  -b --bench           regex to select benchmarks
  -V --verbose         verbose unit-test and goinstall
  --test-bin           name of test-binary (default: gdtest)
  -f --fmt             run gofmt on src and exit
  -r --rewrite         pass rewrite rule to gofmt
  -T --tab             pass -tabindent=true to gofmt
  -w --tabwidth        pass -tabwidth to gofmt (default: 4)
  -e --external        goinstall all external dependencies
  -B --backend         [gc,gccgo,express] (default: gc)
    `

    fmt.Println(helpMSG)
}

func printVersion() {
    fmt.Println("godag 0.2 (r.60.2)")
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
  -g --gdmk            =>   '%s'
  -d --dryrun          =>   %t
  -c --clean           =>   %t
  -q --quiet           =>   %t
  -L --lib             =>   '%s'
  -M --main            =>   '%s'
  -I                   =>   %v
  -D --dot             =>   '%s'
  -t --test            =>   %t
  -m --match           =>   '%s'
  -b --bench           =>   '%s'
  -V --verbose         =>   %t
  --test-bin           =>   '%s'
  -f --fmt             =>   %t
  -r --rewrite         =>   '%s'
  -T --tab             =>   %t
  -w --tabwidth        =>   %s
  -e --external        =>   %t
  -B --backend         =>   '%s'

`
    tabRepr := "4"
    if global.GetString("-tabwidth") != "" {
        tabRepr = global.GetString("-tabwidth")
    }

    fmt.Printf(listMSG,
        global.GetBool("-help"),
        global.GetBool("-version"),
        global.GetBool("-print"),
        global.GetBool("-sort"),
        global.GetString("-output"),
        global.GetBool("-static"),
        global.GetString("-gdmk"),
        global.GetBool("-dryrun"),
        global.GetBool("-clean"),
        global.GetBool("-quiet"),
        global.GetString("-lib"),
        global.GetString("-main"),
        includes,
        global.GetString("-dot"),
        global.GetBool("-test"),
        global.GetString("-match"),
        global.GetString("-bench"),
        global.GetBool("-verbose"),
        global.GetString("-test-bin"),
        global.GetBool("-fmt"),
        global.GetString("-rew-rule"),
        global.GetBool("-tab"),
        tabRepr,
        global.GetBool("-external"),
        global.GetString("-backend"))
}
