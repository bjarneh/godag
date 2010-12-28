package utilz_test

import (
    "testing"
    "path"
    "strings"
    "os"
    "utilz/stringset"
    "utilz/stringbuffer"
    "utilz/walker"
)

func TestStringSet(t *testing.T) {

    ss := stringset.New()

    ss.Add("en")

    if ss.Len() != 1 {
        t.Fatal("stringset.Len() != 1\n")
    }

    ss.Add("to")

    if ss.Len() != 2 {
        t.Fatal("stringset.Len() != 2\n")
    }

    if !ss.Contains("en") {
        t.Fatal("! stringset.Contains('en')\n")
    }

    if !ss.Contains("to") {
        t.Fatal("! stringset.Contains('to')\n")
    }

    if ss.Contains("not here") {
        t.Fatal(" stringset.Contains('not here')\n")
    }
}

func TestStringBuffer(t *testing.T) {

    ss := stringbuffer.New()
    ss.Add("en")
    if ss.String() != "en" {
        t.Fatal(" stringbuffer.String() != 'en'\n")
    }
    ss.Add("to")
    if ss.String() != "ento" {
        t.Fatal(" stringbuffer.String() != 'ento'\n")
    }
    if ss.Len() != 4 {
        t.Fatal(" stringbuffer.Len() != 4\n")
    }
    ss.Add("øæå"); // utf-8 multi-byte fun
    if ss.Len() != 10 {
        t.Fatal(" stringbuffer.Len() != 10\n");
    }
    if ss.String() != "entoøæå" {
        t.Fatal(" stringbuffer.String() != 'entoøæå'\n");
    }
    ss.ClearSize(2)
    if ss.Len() != 0 {
        t.Fatal(" stringbuffer.Len() != 0\n")
    }
    for i := 0; i < 20; i++ {
        if ss.Len() != i {
            t.Fatal(" stringbuffer.Len() != i")
        }
        ss.Add("a")
    }
    if ss.String() != "aaaaaaaaaaaaaaaaaaaa" {
        t.Fatal(" stringbuffer.String() != a * 20")
    }
}

// SRCROOT variable is set during testing
func TestWalker(t *testing.T){

    walker.IncludeDir = func(s string) bool {
        _, dirname := path.Split(s)
        return dirname[0] != '.'
    }
    walker.IncludeFile = func(s string) bool {
        return strings.HasSuffix(s, ".go")
    }

    srcroot := os.Getenv("SRCROOT")

    if srcroot == "" {
        t.Fatalf("$SRCROOT variable not set\n")
    }

    ss := stringset.New()

    // this is a bit static, will cause problems if
    // stuff is added or removed == not ideal..
    ss.Add(path.Join(srcroot, "cmplr", "compiler.go"))
    ss.Add(path.Join(srcroot, "cmplr", "dag.go"))
    ss.Add(path.Join(srcroot, "parse", "gopt.go"))
    ss.Add(path.Join(srcroot, "parse", "gopt_test.go"))
    ss.Add(path.Join(srcroot, "parse", "option.go"))
    ss.Add(path.Join(srcroot, "start", "main.go"))
    ss.Add(path.Join(srcroot, "utilz", "handy.go"))
    ss.Add(path.Join(srcroot, "utilz", "stringbuffer.go"))
    ss.Add(path.Join(srcroot, "utilz", "stringset.go"))
    ss.Add(path.Join(srcroot, "utilz", "utilz_test.go"))
    ss.Add(path.Join(srcroot, "utilz", "walker.go"))
    ss.Add(path.Join(srcroot, "utilz", "global.go"))

    files   := walker.PathWalk(srcroot)

    // make sure stringset == files

    if files.Len() != ss.Len() {
        t.Fatalf("walker.Len() != files.Len()\n");
    }

    for i := 0; i < files.Len(); i++ {
        if ! ss.Contains( files.At(i) ){
            t.Fatalf("walker picked up files not in SRCROOT\n")
        }
        ss.Remove( files.At(i) )
    }

}
