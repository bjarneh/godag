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

package gopt_test

import (
    "parse/gopt"
    "strings"
    "testing"
)

func TestGetOpt(t *testing.T) {

    getopt := gopt.New()

    getopt.BoolOption("-h -help --help")
    getopt.BoolOption("-v -version --version")
    getopt.StringOption("-f -file --file --file=")
    getopt.StringOption("-num=")
    getopt.StringOption("-I")

    argv := strings.Split("-h -num=7 -version not-option -fsomething -I/dir1 -I/dir2", " ")

    args := getopt.Parse(argv)

    if !getopt.IsSet("-help") {
        t.Fatal("! getopt.IsSet('-help')\n")
    }

    if !getopt.IsSet("-version") {
        t.Fatal(" ! getopt.IsSet('-version')\n")
    }

    if !getopt.IsSet("-file") {
        t.Fatal(" ! getopt.IsSet('-file')\n")
    } else {
        if getopt.Get("-f") != "something" {
            t.Fatal(" getopt.Get('-f') != 'something'\n")
        }
    }

    if !getopt.IsSet("-num=") {
        t.Fatal(" ! getopt.IsSet('-num=')\n")
    } else {
        n, e := getopt.GetInt("-num=")
        if e != nil {
            t.Fatalf(" getopt.GetInt error = %s\n", e)
        }
        if n != 7 {
            t.Fatalf(" getopt.GetInt != 7 (%d)\n", n)
        }
    }

    if !getopt.IsSet("-I") {
        t.Fatal(" ! getopt.IsSet('-I')\n")
    } else {
        elms := getopt.GetMultiple("-I")
        if len(elms) != 2 {
            t.Fatal("getopt.GetMultiple('-I') != 2\n")
        }
        if elms[0] != "/dir1" {
            t.Fatal("getopt.GetMultiple('-I')[0] != '/dir1'\n")
        }
        if elms[1] != "/dir2" {
            t.Fatal("getopt.GetMultiple('-I')[1] != '/dir2'\n")
        }
    }

    if len(args) != 1 {
        t.Fatal("len(remaining) != 1\n")
    }

    if args[0] != "not-option" {
        t.Fatal("remaining[0] != 'not-something'\n")
    }
}
