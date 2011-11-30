//  Copyright © 2010 bjarneh
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

package utilz_test

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
    "utilz/stringbuffer"
    "utilz/stringset"
    "utilz/timer"
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
    ss.Add("øæå") // utf-8 multi-byte fun
    if ss.Len() != 10 {
        t.Fatal(" stringbuffer.Len() != 10\n")
    }
    if ss.String() != "entoøæå" {
        t.Fatal(" stringbuffer.String() != 'entoøæå'\n")
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
func TestWalker(t *testing.T) {

    walker.IncludeDir = func(s string) bool {
        _, dirname := filepath.Split(s)
        return dirname[0] != '.'
    }
    walker.IncludeFile = func(s string) bool {
        return strings.HasSuffix(s, ".go") &&
            !strings.HasPrefix(s, "_")
    }

    srcroot := os.Getenv("SRCROOT")

    if srcroot == "" {
        t.Fatalf("$SRCROOT variable not set\n")
    }

    ss := stringset.New()

    // this is a bit static, will cause problems if
    // stuff is added or removed == not ideal..
    ss.Add(filepath.Join(srcroot, "cmplr", "compiler.go"))
    ss.Add(filepath.Join(srcroot, "cmplr", "dag.go"))
    ss.Add(filepath.Join(srcroot, "cmplr", "gdmake.go"))
    ss.Add(filepath.Join(srcroot, "parse", "gopt.go"))
    ss.Add(filepath.Join(srcroot, "parse", "gopt_test.go"))
    ss.Add(filepath.Join(srcroot, "parse", "option.go"))
    ss.Add(filepath.Join(srcroot, "start", "main.go"))
    ss.Add(filepath.Join(srcroot, "utilz", "handy.go"))
    ss.Add(filepath.Join(srcroot, "utilz", "stringbuffer.go"))
    ss.Add(filepath.Join(srcroot, "utilz", "stringset.go"))
    ss.Add(filepath.Join(srcroot, "utilz", "utilz_test.go"))
    ss.Add(filepath.Join(srcroot, "utilz", "walker.go"))
    ss.Add(filepath.Join(srcroot, "utilz", "global.go"))
    ss.Add(filepath.Join(srcroot, "utilz", "timer.go"))
    ss.Add(filepath.Join(srcroot, "utilz", "say.go"))

    files := walker.PathWalk(filepath.Clean(srcroot))

    // make sure stringset == files

    if len(files) != ss.Len() {
        t.Fatalf("walker.Len() != files.Len()\n")
    }

    for i := 0; i < len(files); i++ {
        if !ss.Contains(files[i]) {
            t.Fatalf("walker picked up files not in SRCROOT\n")
        }
        ss.Remove(files[i])
    }

}

func TestTimer(t *testing.T) {

    timer.Start("is here")
    err := timer.Stop("not here")

    if err == nil {
        t.Fatalf("job: 'not here' is here\n")
    }

    err = timer.Stop("is here")

    if err != nil {
        t.Fatalf("job: 'is here' is not here\n")
    }

    delta, err := timer.Delta("is here")

    if err != nil {
        t.Fatalf("job: 'is here' still not here..\n")
    }

    if delta < 0 {
        t.Fatalf("delta = %d < 0 ns\n", delta)
    }

    delta = timer.Hour*4 + timer.Minute*7 + timer.Second*3 + timer.Millisecond*9

    tid := timer.Nano2Time(delta)

    if tid.Hours != 4 {
        t.Fatalf("timer.Nano2Time() 4 != %d\n", tid.Hours)
    }

    if tid.Minutes != 7 {
        t.Fatalf("timer.Nano2Time() 7 != %d\n", tid.Minutes)
    }

    if tid.Seconds != 3 {
        t.Fatalf("timer.Nano2Time() 3 != %d\n", tid.Seconds)
    }

    if tid.Milliseconds != 9 {
        t.Fatalf("timer.Nano2Time() 9 != %d\n", tid.Milliseconds)
    }

}
