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

package gopt

/*

The flag package provided in the go standard
library will only allow options before regular
input arguments, which is not ideal. So this
is another take on the old getopt.

Most notable difference:

  - Multiple 'option-strings' for a single option ('-r -rec -R')
  - Non-option arguments can come anywhere in argv
  - Option arguments can be in juxtaposition with flag
  - Only two types of options: string, bool


Usage:

 getopt := gopt.New();

 getopt.BoolOption("-h -help --help");
 getopt.BoolOption("-v -version --version");
 getopt.StringOption("-f -file --file --file=");
 getopt.StringOption("-l -list --list");
 getopt.StringOption("-I");

 args := getopt.Parse(os.Args[1:]);

 // getopt.IsSet("-h") == getopt.IsSet("-help") ..

 if getopt.IsSet("-help"){ println("-help"); }
 if getopt.IsSet("-v")   { println("-version"); }
 if getopt.IsSet("-file"){ println("--file ",getopt.Get("-f")); }
 if getopt.IsSet("-list"){ println("--list ",getopt.Get("-list")); }

 if getopt.IsSet("-I"){
     elms := getopt.GetMultiple("-I");
     for y := range elms { println("-I ",elms[y]);  }
 }

 for i := range args{
     println("remaining:",args[i]);
 }


*/

import (
    "log"
    "strconv"

    "strings"
)

type GetOpt struct {
    options []Option
    cache   map[string]Option
}

func New() *GetOpt {
    g := new(GetOpt)
    g.options = make([]Option, 0)
    g.cache = make(map[string]Option)
    return g
}

func (g *GetOpt) isOption(o string) Option {
    _, ok := g.cache[o]
    if ok {
        element, _ := g.cache[o]
        return element
    }
    return nil
}

func (g *GetOpt) getStringOption(o string) *StringOption {

    opt := g.isOption(o)

    if opt != nil {
        sopt, ok := opt.(*StringOption)
        if ok {
            return sopt
        } else {
            log.Fatalf("[ERROR] %s: is not a string option\n", o)
        }
    } else {
        log.Fatalf("[ERROR] %s: is not an option at all\n", o)
    }

    return nil
}

func (g *GetOpt) Get(o string) string {

    sopt := g.getStringOption(o)

    switch sopt.count {
    case 0:
        log.Fatalf("[ERROR] %s: is not set\n", o)
    case 1: // fine do nothing
    default:
        log.Printf("[WARNING] option %s: has more arguments than 1\n", o)
    }
    return sopt.values[0]
}

func (g *GetOpt) GetFloat32(o string) (float32, error) {
    f, e := strconv.ParseFloat(g.Get(o), 32)
    if e != nil {
        return 0.0, e
    }
    return float32(f), nil
}

func (g *GetOpt) GetFloat64(o string) (float64, error) {
    return strconv.ParseFloat(g.Get(o), 64)
}

func (g *GetOpt) GetInt(o string) (int, error) {
    return strconv.Atoi(g.Get(o))
}

func (g *GetOpt) Reset() {
    for _, v := range g.cache {
        v.reset()
    }
}

func (g *GetOpt) GetMultiple(o string) []string {

    sopt := g.getStringOption(o)

    if sopt.count == 0 {
        log.Fatalf("[ERROR] %s: is not set\n", o)
    }

    return sopt.values[0:sopt.count]
}

func (g *GetOpt) Parse(argv []string) (args []string) {

    args = make([]string, 0)

    for i := 0; i < len(argv); i++ {

        opt := g.isOption(argv[i])

        if opt != nil {

            switch opt.(type) {
            case *BoolOption:
                bopt, _ := opt.(*BoolOption)
                bopt.setFlag()
            case *StringOption:
                sopt, _ := opt.(*StringOption)
                if i+1 >= len(argv) {
                    log.Fatalf("[ERROR] missing argument for: %s\n", argv[i])
                } else {
                    sopt.addArgument(argv[i+1])
                    i++
                }
            }

        } else {

            // arguments written next to options
            start, ok := g.juxtaStringOption(argv[i])

            if ok {
                stropt := g.getStringOption(start)
                stropt.addArgument(argv[i][len(start):])
            } else {

                boolopts, ok := g.juxtaBoolOption(argv[i])

                if ok {

                    for i := 0; i < len(boolopts); i++ {
                        boolopt, _ := g.isOption(boolopts[i]).(*BoolOption)
                        boolopt.setFlag()
                    }

                } else {
                    args = append(args, argv[i])
                }
            }
        }
    }

    return args
}

func (g *GetOpt) juxtaStringOption(opt string) (string, bool) {

    var tmpmax string = ""

    for i := 0; i < len(g.options); i++ {

        sopt, ok := g.options[i].(*StringOption)

        if ok {
            s := sopt.startsWith(opt)
            if s != "" {
                if len(s) > len(tmpmax) {
                    tmpmax = s
                }
            }
        }
    }

    if tmpmax != "" {
        return tmpmax, true
    }

    return "", false
}

// convert: -abc => -a -b -c
func (g *GetOpt) juxtaBoolOption(opt string) ([]string, bool) {

    var tmp string

    if !strings.HasPrefix(opt, "-") || len(opt) < 3 {
        return nil, false
    }

    bopts := make([]string, 0)
    couldBe := strings.Split(opt[1:], "")

    for i := 0; i < len(couldBe); i++ {

        tmp = "-" + couldBe[i]
        opt := g.isOption(tmp)

        if opt != nil {
            _, ok := opt.(*BoolOption)
            if ok {
                bopts = append(bopts, tmp)
            } else {
                return nil, false
            }
        } else {
            return nil, false
        }
    }

    return bopts, true
}

func (g *GetOpt) IsSet(o string) bool {
    _, ok := g.cache[o]
    if ok {
        element, _ := g.cache[o]
        return element.isSet()
    } else {
        log.Fatalf("[ERROR] %s not an option\n", o)
    }
    return false
}

func (g *GetOpt) BoolOption(optstr string) {
    ops := strings.Split(optstr, " ")
    boolopt := newBoolOption(ops)
    for i := range ops {
        g.cache[ops[i]] = boolopt
    }
    g.options = append(g.options, boolopt)
}

func (g *GetOpt) StringOption(optstr string) {
    ops := strings.Split(optstr, " ")
    stringopt := newStringOption(ops)
    for i := range ops {
        g.cache[ops[i]] = stringopt
    }
    g.options = append(g.options, stringopt)
}

func (g *GetOpt) StringOptionFancy(optstr string) {
    ops := Convert(optstr)
    stringopt := newStringOption(ops)
    for i := range ops {
        g.cache[ops[i]] = stringopt
    }
    g.options = append(g.options, stringopt)
}

// '-f' -> [-f, -f=]
// '--file' -> [-file, -file=, --file, --file=]
// '-f --file' -> [-f, -f=, -file, -file=, --file, --file=]
func Convert(optstr string) (opts []string) {

    ops := strings.Split(optstr, " ")
    convOps := make([]string, len(ops))
    copy(convOps, ops)

    for i := 0; i < len(ops); i++ {
        // could be wierd UTF-8 char i.e. -ø -ł -Ħ ...
        point := []rune(ops[i])
        if len(point) == 2 && point[0] == rune('-') {
            convOps = append(convOps, ops[i]+"=")
        } else if len(point) > 3 && ops[i][:2] == "--" {
            convOps = append(convOps, ops[i]+"=")
            convOps = append(convOps, ops[i][1:])
            convOps = append(convOps, ops[i][1:]+"=")
        }
    }
    return convOps
}
