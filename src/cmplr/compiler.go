// © Knug Industries 2009 all rights reserved 
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package compiler

import(
    "os";
    "container/vector";
    "fmt";
    "utilz/handy";
    "cmplr/dag";
    "path";
)


type Compiler struct{
    root, arch, suffix, executable string;
    dryrun bool;
    includes []string;
}

func New(root, arch string, dryrun bool,include []string) *Compiler{
    c                := new(Compiler);
    c.root           = root;
    c.arch, c.suffix = archNsuffix(arch);
    c.executable     = findCompiler(c.arch);
    c.dryrun         = dryrun;
    c.includes       = include;
    return c;
}

func findCompiler(arch string) string{

    var lookingFor string;
    switch arch {
        case "arm"  : lookingFor = "5g";
        case "amd64": lookingFor = "6g";
        case "386"  : lookingFor = "8g";
    }

    real := handy.Which(lookingFor);
    if real == "" {
        die("[ERROR] could not find compiler\n");
    }
    return real;
}

func findLinker(arch string) string{

    var lookingFor string;
    switch arch {
        case "arm"  : lookingFor = "5l";
        case "amd64": lookingFor = "6l";
        case "386"  : lookingFor = "8l";
    }

    real := handy.Which(lookingFor);
    if real == "" {
        die("[ERROR] could not find linker\n");
    }
    return real;
}


func archNsuffix(arch string)(a, s string){

    if arch == "" {
        a = os.Getenv("GOARCH");
    }else{
        a = arch;
    }

    switch a {
        case "arm"  : s = ".5";
        case "amd64": s = ".6";
        case "386"  : s = ".8";
        default     : die("[ERROR] unknown architecture: %s\n",a);
    }

    return a, s;
}

func (c *Compiler) String() string{
    s := "Compiler{ root=%s, arch=%s, suffix=%s, executable=%s }";
    return fmt.Sprintf(s, c.root, c.arch, c.suffix, c.executable);
}


func (c *Compiler) ForkCompile(pkgs *vector.Vector){

    includeLen := c.extraPkgIncludes();

    for p := range pkgs.Iter() {
        pkg, _ := p.(*dag.Package);//safe cast, only Packages there

        argv := make([]string, 5 + pkg.Files.Len() + (includeLen*2));
        i    := 0;
        argv[i] = c.executable; i++;
        argv[i] = "-I"; i++;
        argv[i] = c.root; i++;
        if includeLen > 0 {
            for y := 0; y < includeLen; y ++ {
                argv[i] = "-I"; i++;
                argv[i] = c.includes[y]; i++;
            }
        }
        argv[i] = "-o"; i++;
        argv[i] = path.Join(c.root, pkg.Name) + c.suffix; i++;

        for f := range pkg.Files.Iter() {
            argv[i] = f;
            i++;
        }

        if c.dryrun {
            dryRun(argv);
        }else{
            fmt.Println("compiling:",pkg.Name);
            handy.StdExecve(argv, true);
        }
    }
}
// for removal of temoprary packages created for testing and so on..
func (c *Compiler) DeletePackages(pkgs *vector.Vector) bool{
    var ok = true;
    var e os.Error;

    for p := range pkgs.Iter() {
        pkg, _ := p.(*dag.Package);//safe cast, only Packages there

        for f := range pkg.Files.Iter(){
            e = os.Remove(f);
            if e != nil{
                ok = false;
                fmt.Fprintf(os.Stderr,"[ERROR] %s\n",e);
            }
        }
        pcompile := path.Join(c.root, pkg.Name) + c.suffix;
        e = os.Remove(pcompile);
        if e != nil{
            ok = false;
            fmt.Fprintf(os.Stderr,"[ERROR] %s\n",e);
        }
    }

    return ok;
}

func (c *Compiler) ForkLink(pkgs *vector.Vector, output string, static bool){

    gotMain := new(vector.Vector);

    for p := range pkgs.Iter() {
        pk, _ := p.(*dag.Package);
        if pk.ShortName == "main" {
            gotMain.Push( pk );
        }
    }

    if gotMain.Len() == 0 {
        die("[ERROR] (linking) no main package found\n");
    }

    if gotMain.Len() > 1 {
        die("[ERROR] (linking) more than one main package found\n");
    }

    includeLen := c.extraPkgIncludes();
    staticXtra := 0;
    if static { staticXtra++; }

    pkg, _ := gotMain.Pop().(*dag.Package);

    linker := findLinker(c.arch);
    compiled := path.Join(c.root, pkg.Name) + c.suffix;

    argv := make([]string, 6 + (includeLen*2) + staticXtra);
    i    := 0;
    argv[i] = linker; i++;
    argv[i] = "-o"; i++;
    argv[i] = output; i++;
    argv[i] = "-L"; i++;
    argv[i] = c.root; i++;
    if static { argv[i] = "-d"; i++; }
    if includeLen > 0{
        for y := 0; y < includeLen; y++ {
            argv[i] = "-L"; i++;
            argv[i] = c.includes[y]; i++;
        }
    }
    argv[i] = compiled; i++;

    if c.dryrun {
        dryRun(argv);
    }else{
        fmt.Println("linking  :",output);
        handy.StdExecve(argv, true);
    }
}

func die(strfmt string, v ...interface{}){
    fmt.Fprintf(os.Stderr, strfmt, v);
    os.Exit(1);
}


func dryRun(argv []string){
    var cmd string;

    for i := 0; i < len(argv); i++ {
        cmd = fmt.Sprintf("%s %s ", cmd, argv[i]);
    }

    fmt.Printf("%s || exit 1\n",cmd);
}

func (c *Compiler) extraPkgIncludes() int{
    if c.includes != nil && len(c.includes) > 0 {
        return len(c.includes);
    }
    return 0;
}
