// © Knug Industries 2009 all rights reserved 
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package main

import(
    "os";
    "fmt";
    "path";
    "utilz/walker";
    "cmplr/compiler";
    "cmplr/dag";
    "container/vector";
    "parse/gopt";
    "strings";
    "utilz/handy";
)


func init(){

    // override IncludeFile to make walker pick up only .go files
    walker.IncludeFile = func(s string)bool{
        return strings.HasSuffix(s,".go") &&
             ! strings.HasSuffix(s, "_test.go");
    };

    // override IncludeDir to make walker ignore 'hidden' directories
    walker.IncludeDir = func(s string)bool{
        _, dirname := path.Split(s);
        return dirname[0] != '.';
    };

}

func gotRoot(){
    if os.Getenv("GOROOT") == "" {
        fmt.Fprintf(os.Stderr,"[ERROR] missing GOROOT\n");
        os.Exit(1);
    }
}

func findTestFilesAlso(){
    // override IncludeFile to make walker pick up only .go files
    walker.IncludeFile = func(s string)bool{
        return strings.HasSuffix(s,".go");
    };
}

func main(){

    var files *vector.StringVector;

    var arch, output, srcdir, bmatch, match, rewRule, tabWidth string;
    var dryrun, test, testVerbose, static, noComments, tabIndent bool;
    var includes []string = nil;

    getopt := gopt.New();

    getopt.BoolOption("-h -help --help help");
    getopt.BoolOption("-c -clean --clean clean");
    getopt.BoolOption("-S -static --static");
    getopt.BoolOption("-v -version --version version");
    getopt.BoolOption("-s -sort --sort sort");
    getopt.BoolOption("-p -print --print");
    getopt.BoolOption("-d -dryrun --dryrun");
    getopt.BoolOption("-t -test --test");
    getopt.BoolOption("-V -verbose --verbose");
    getopt.BoolOption("-f -fmt --fmt");
    getopt.BoolOption("-nc --nc");
    getopt.BoolOption("-tab --tab");
    getopt.StringOption("-a -arch --arch -arch= --arch=");
    getopt.StringOption("-I");
    getopt.StringOption("-tw --tw -tw= --tw=");
    getopt.StringOption("-rw --rw -rw= --rw=");
    getopt.StringOption("-o -output --output -output= --output=");
    getopt.StringOption("-b -benchmarks --benchmarks -benchmarks= --benchmarks=");
    getopt.StringOption("-m -match --match -match= --match=");

    args := getopt.Parse(os.Args[1:]);

    if len(args) == 0{
        srcdir = "src";
    }else{
        if len(args) > 1 {
            fmt.Fprintf(os.Stderr,"[WARNING] len(input directories) > 1\n");
        }
        srcdir = args[0];
    }

    if getopt.IsSet("-help") { printHelp(); os.Exit(0); }
    if getopt.IsSet("-version") { printVersion(); os.Exit(0); }
    if getopt.IsSet("-clean") { rm865(srcdir); os.Exit(0); }
    if getopt.IsSet("-dryrun"){ dryrun = true; }
    if getopt.IsSet("-static"){ static = true; }
    if getopt.IsSet("-verbose"){ testVerbose = true; }
    if getopt.IsSet("-test"){
        test = true;
        findTestFilesAlso();
    }

    gotRoot();//?

    if getopt.IsSet("-arch"){ arch = getopt.Get("-a"); }
    if getopt.IsSet("-output"){ output = getopt.Get("-o"); }
    if getopt.IsSet("-benchmarks"){ bmatch = getopt.Get("-b"); }
    if getopt.IsSet("-match"){ match = getopt.Get("-m"); }
    if getopt.IsSet("-I"){ includes = getopt.GetMultiple("-I"); }


    files = findFiles(srcdir);


    if getopt.IsSet("-fmt"){
        if getopt.IsSet("-nc"){ noComments = true; }
        if getopt.IsSet("-rw"){ rewRule = getopt.Get("-rw"); }
        if getopt.IsSet("-tab"){ tabIndent = true; }
        if getopt.IsSet("-tw"){ tabWidth = getopt.Get("-tw"); }
        formatFiles(files, tabIndent, noComments, rewRule, tabWidth);
    }

    dgrph := dag.New();
    dgrph.Parse(srcdir, files);

    if getopt.IsSet("-print") {
        dgrph.PrintInfo();
        os.Exit(0);
    }

    dgrph.GraphBuilder(includes);
    sorted := dgrph.Topsort();

    if getopt.IsSet("-sort") {
        for pkg := range sorted.Iter() {
            rpkg, _ := pkg.(*dag.Package);
            fmt.Printf("%s\n", rpkg.Name);
        }
        os.Exit(0);
    }

    kompiler  := compiler.New(srcdir, arch, dryrun, includes);
    kompiler.ForkCompile(sorted);

    if test {
        testMain, testDir := dgrph.MakeMainTest(srcdir);
        kompiler.ForkCompile(testMain);
        kompiler.ForkLink(testMain, "gdtest", false);
        kompiler.DeletePackages(testMain);
        rmError := os.Remove(testDir);
        if rmError != nil {
            fmt.Fprintf(os.Stderr,"[ERROR] failed to remove testdir: %s\n",testDir);
        }
        testArgv := createTestArgv("gdtest", bmatch, match, testVerbose);
        tstring := "testing  : ";
        if testVerbose { tstring += "\n"; }
        fmt.Printf(tstring);
        ok := handy.StdExecve(testArgv, false);
        e := os.Remove("gdtest");
        if e != nil{
            fmt.Fprintf(os.Stderr,"[ERROR] %s\n",e);
        }
        if ! ok {
            os.Exit(1);
        }
    }

    if output != "" {
        kompiler.ForkLink(sorted, output, static);
    }

}

func createTestArgv(prg, bmatch, match string, tverb bool) ([]string) {
    var numArgs int = 1;
    pwd, e := os.Getwd();
    if e != nil {
        fmt.Fprintf(os.Stderr,"[ERROR] could not locate working directory\n");
        os.Exit(1);
    }
    arg0 := path.Join(pwd, prg);
    if bmatch != "" { numArgs += 2; }
    if match  != "" { numArgs += 2; }
    if tverb        { numArgs++;    }

    var i = 1;
    argv := make([]string, numArgs);
    argv[0] = arg0;
    if bmatch != "" {
        argv[i] = "-benchmarks"; i++;
        argv[i] = bmatch; i++;
    }
    if match != "" {
        argv[i] = "-match"; i++;
        argv[i] = match; i++;
    }
    if tverb {
        argv[i] = "-v";
    }
    return argv;
}

func findFiles(pathname string) *vector.StringVector{
    okDirOrDie(pathname);
    return walker.PathWalk(path.Clean(pathname));
}

func okDirOrDie(pathname string){

    var dir *os.FileInfo;
    var staterr  os.Error;

    dir, staterr = os.Stat(pathname);

    if staterr != nil {
        fmt.Fprintf(os.Stderr,"[ERROR] %s\n", staterr);
        os.Exit(1);
    }else if ! dir.IsDirectory() {
        fmt.Fprintf(os.Stderr,"[ERROR] %s: is not a directory\n", pathname);
        os.Exit(1);
    }
}

func formatFiles(files *vector.StringVector, tab, noC bool, rew, tw string){

    var i int = 0;
    var argvLen int = 0;
    var argv []string;
    var tabWidth string  = "-tabwidth=4";
    var useTabs  string  = "-tabindent=false";
    var comments string  = "-comments=true";
    var rewRule  string  = "";
    var fmtexec  string  = handy.Which("gofmt");

    if tw != "" { tabWidth = "-tabwidth="+tw; }
    if noC { comments = "-comments=false"; }
    if rew != "" { rewRule = rew; argvLen++ ; }
    if tab { useTabs = "-tabindent=true"; }

    argv = make([]string, 6 + argvLen);

    if fmtexec == "" {
        fmt.Fprintf(os.Stderr,"[ERROR] could not find: gofmt\n");
        os.Exit(1);
    }

    argv[i] = fmtexec;   i++;
    argv[i] = "-w=true"; i++;
    argv[i] = tabWidth;  i++;
    argv[i] = useTabs;   i++;
    argv[i] = comments;  i++;

    if rewRule != "" {
        argv[i] = "-r="+rewRule; i++;
    }

    for fileName := range files.Iter() {
        argv[i] = fileName;
        fmt.Printf("gofmt    : %s\n", fileName);
        _ = handy.StdExecve(argv, true);
    }

}

func rm865(srcdir string){

    // override IncludeFile to make walker pick up only .[865] files
    walker.IncludeFile = func(s string)bool{
        return strings.HasSuffix(s,".8") ||
               strings.HasSuffix(s,".6") ||
               strings.HasSuffix(s,".5") ||
               strings.HasSuffix(s,".a");

    };

    okDirOrDie(srcdir);

    compiled := walker.PathWalk(path.Clean(srcdir));

    for s := range compiled.Iter() {
        fmt.Printf("rm: %s\n", s);
        e := os.Remove(s);
        if e != nil {
            fmt.Fprintf(os.Stderr,"[ERROR] could not delete file: %s\n",s);
        }
    }
}

func printHelp(){
    var helpMSG string =`
  godag is a compiler front-end for golang,
  hopefully you can avoid Makefiles.

  usage: gd [OPTIONS] src-directory

  options:

  -h --help            print this message and quit
  -v --version         print version and quit
  -p --print           print package info collected
  -s --sort            print legal compile order
  -o --output          link main package -> output
  -S --static          statically link binary
  -a --arch            architecture (amd64,arm,386)
  -d --dryrun          print what gd would do (stdout)
  -c --clean           rm *.[a865] from src-directory
  -I                   import package directories
  -t --test            run all unit-tests
  -b --benchmarks      pass argument to unit-test
  -m --match           pass argument to unit-test
  -V --verbose         pass argument '-v' to unit-test
  -rw                  pass rewrite rule to gofmt
  -nc                  pass -comments=false to gofmt
  -tw                  pass -tabwidth to gofmt (default:4)
  -tab                 pass -tabindent=true to gofmt
    `;

    fmt.Println(helpMSG);
}

func printVersion(){
    fmt.Println("godag 0.1");
}
