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

package main

import (
    "cmplr/compiler"
    "cmplr/dag"
    "cmplr/gdmake"
    "fmt"
    "log"
    "os"
    "parse/gopt"
    "path/filepath"
    "runtime"
    "strings"
    "utilz/global"
    "utilz/handy"
    "utilz/say"
    "utilz/timer"
    "utilz/walker"
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
    "-all",
    "-quiet",
    "-tab",
    "-external",
    // add missing test options + alias
    "-test.short",
    "-test.v",
    "-strip",
}

// keys for the string options
// note: -I is handled seperately
var strs = []string{
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
    "-mkcomplete",
    // add missing test options + alias
    "-test.bench",
    "-test.benchtime",
    "-test.cpu",
    "-test.cpuprofile",
    "-test.memprofile",
    "-test.memprofilerate",
    "-test.timeout",
    "-test.parallel",
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
    getopt.BoolOption("-p -print --print print")
    getopt.BoolOption("-d -dryrun --dryrun dryrun")
    getopt.BoolOption("-t -test --test test")
    getopt.BoolOption("-l -list --list list")
    getopt.BoolOption("-q -quiet --quiet")
    getopt.BoolOption("-V -verbose --verbose")
    getopt.BoolOption("-f -fmt --fmt fmt")
    getopt.BoolOption("-T -tab --tab")
    getopt.BoolOption("-a -all --all")
    getopt.BoolOption("-y -strip --strip strip")
    getopt.BoolOption("-e -external --external")
    getopt.StringOption("-I -I=")
    getopt.StringOption("-mkcomplete")
    getopt.StringOptionFancy("-D --dot")
    getopt.StringOptionFancy("-L --lib")
    getopt.StringOptionFancy("-g --gdmk")
    getopt.StringOptionFancy("-w --tabwidth")
    getopt.StringOptionFancy("-r --rewrite")
    getopt.StringOptionFancy("-o --output")
    getopt.StringOptionFancy("-M --main")
    getopt.StringOptionFancy("-b --bench")
    getopt.StringOptionFancy("-m --match")
    getopt.StringOptionFancy("--test-bin")
    getopt.StringOptionFancy("-B --backend")

    // new test options and aliases
    getopt.BoolOption("-test.short --test.short")
    getopt.BoolOption("-test.v --test.v")
    getopt.StringOptionFancy("--test.bench")
    getopt.StringOptionFancy("--test.benchtime")
    getopt.StringOptionFancy("--test.cpu")
    getopt.StringOptionFancy("--test.cpuprofile")
    getopt.StringOptionFancy("--test.memprofile")
    getopt.StringOptionFancy("--test.memprofilerate")
    getopt.StringOptionFancy("--test.timeout")
    getopt.StringOptionFancy("--test.parallel")

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
    goos := handy.GOOS()

    if goos == "windows" {
        global.SetString("-test-bin", "gdtest.exe")
    } else {
        global.SetString("-test-bin", "gdtest")
    }

    global.SetString("-backend", runtime.Compiler)
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

func reportTime() {
    timer.Stop("everything")
    delta, _ := timer.Delta("everything")
    say.Printf("time used: %s\n", timer.Nano2Time(delta))
}

func main() {

    var (
        ok, up2date bool
        e           error
        argv, args  []string
        config      [4]string
    )

    timer.Start("everything")
    defer reportTime()

    // possible config locations
    config[0] = filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "godag", "gdrc")
    config[1] = filepath.Join(os.Getenv("HOME"), ".config", "godag", "gdrc")
    config[2] = filepath.Join(os.Getenv("HOME"), ".gdrc")
    config[3] = filepath.Join(os.Getenv("PWD"), ".gdrc")

    for _, conf := range config {

        argv, ok = handy.ConfigToArgv(conf)

        if ok {
            args = parseArgv(argv)
            if len(args) > 0 {
                log.Print("[WARNING] non-option arguments in config file\n")
            }
        }
    }

    // small gorun version if single go-file is given
    if len(os.Args) > 1 && strings.HasSuffix(os.Args[1], ".go") && handy.IsFile(os.Args[1]) {
        say.Mute() // be silent unless error here
        single, name := dag.ParseSingle(os.Args[1])
        compiler.InitBackend()
        compiler.CreateArgv(single)
        up2date = compiler.Compile(single)
        if handy.GOOS() == "windows" {
            name = name + ".exe"
        }
        compiler.ForkLink(name, single, nil, up2date)
        args = os.Args[1:]
        args[0] = name
        handy.StdExecve(args, true)
        os.Exit(0)
    }


    // command line arguments overrides/appends config
    args = parseArgv(os.Args[1:])

    mkcomplete := global.GetString("-mkcomplete")
    if mkcomplete != "" {
        fmt.Println(mkcomplete)
        targets := dag.GetMakeTargets(mkcomplete)
        for _, t := range targets {
            fmt.Println(t)
        }
        os.Exit(0)
    }

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
        includes[i] = os.ExpandEnv(includes[i])
    }

    // expand variables in -lib
    global.SetString("-lib", os.ExpandEnv(global.GetString("-lib")))

    // expand variables in -output
    global.SetString("-output", os.ExpandEnv(global.GetString("-output")))

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
        if !handy.IsDir("src") {
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
            if global.GetBool("-verbose") || global.GetBool("-test.v") {
                say.Printf("\n")
            }
            ok = handy.StdExecve(testArgv, false)
            handy.Delete(global.GetString("-test-bin"), false)
            if !ok {
                os.Exit(1)
            }
        }

        // if packages contain both test-files and regular files
        // test-files should not be part of the objects, i.e. init
        // functions in test-packages can cause unexpected behaviour

        if compiler.ReCompile(sorted) {
            say.Printf("recompile: --tests\n")
            compiler.Compile(sorted)
        }

    }

    // link if ! up2date
    if global.GetString("-output") != "" {
        compiler.ForkLink(global.GetString("-output"), sorted, nil, up2date)
    } else if global.GetBool("-all") {
        compiler.ForkLinkAll(sorted, up2date)
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
  -y --strip           strip symbols from executable
  -g --gdmk            create a go makefile for project
  -d --dryrun          print what gd would do (stdout)
  -c --clean           delete generated object code
  -q --quiet           silent, print only errors
  -L --lib             write objects to other dir (!src)
  -M --main            regex to select main package
  -a --all             link main pkgs to bin/nameOfMainDir
  -D --dot             create a graphviz dot file
  -I                   import package directories
  -t --test            run all unit-tests
  -m --match           regex to select unit-tests
  -b --bench           regex to select benchmarks
  -V --verbose         verbose unit-test and go install
  --test-bin           name of test-binary (default: gdtest)
  --test.*             any valid gotest option
  -f --fmt             run gofmt on src and exit
  -r --rewrite         pass rewrite rule to gofmt
  -T --tab             pass -tabs=true to gofmt
  -w --tabwidth        pass -tabwidth to gofmt (default: 4)
  -e --external        go install all external dependencies
  -B --backend         [gc,gccgo,express] (default: gc)
    `

    fmt.Println(helpMSG)
}

func printVersion() {
    fmt.Println("godag 0.3 (release-branch.go1)")
}

func printListing() {

    fmt.Println("\n Listing of options and their content:\n")
    defer fmt.Println("")

    for i := 0; i < len(bools); i++ {
        fmt.Printf(" %-20s  =>    %v\n", bools[i], global.GetBool(bools[i]))
    }

    for i := 0; i < len(strs); i++ {
        fmt.Printf(" %-20s  =>    %v\n", strs[i], global.GetString(strs[i]))
    }

    fmt.Printf(" %-20s  =>    %v\n", "-lib", includes)
}
