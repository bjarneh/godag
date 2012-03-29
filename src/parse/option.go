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

type Option interface {
    isSet() bool
    reset()
}

type StringOption struct {
    // all options belonging to this StringOption
    options []string
    // multiple values are allowed
    values []string
    // count the number of arguments following
    // this string option, note multiple can be
    // set with multiple definitions, i.e.,
    // -I/some/dir -I/other/dir ...
    count int
}

func newStringOption(op []string) *StringOption {
    s := new(StringOption)
    s.options = op
    s.values = make([]string, 5) // default to 5
    s.count = 0
    return s
}

func (s *StringOption) isSet() bool {
    return s.count > 0
}

func (s *StringOption) reset() {
    s.values = make([]string, 5)
    s.count = 0
}

func indexOf(s1, s2 string) int {

    if len(s1) < len(s2) {
        return -1
    }

    slice1 := s1[0:len(s2)]

    if slice1 == s2 {
        return len(s2)
    }

    return -1
}

func (s *StringOption) startsWith(opt string) string {

    var max int = -1
    var tmp int = -1
    var optstr string = ""

    for i := range s.options {
        tmp = indexOf(opt, s.options[i])
        if tmp > max {
            max = tmp
            optstr = s.options[i]
        }
    }

    return optstr
}

func (s *StringOption) addArgument(arg string) {
    if (s.count + 1) == len(s.values) {
        resize := make([]string, len(s.values)*2)
        for i := range s.values {
            resize[i] = s.values[i]
        }
        s.values = resize
    }
    s.values[s.count] = arg
    s.count++
}

func (s *StringOption) String() string {
    return "StringOption " + s.options[0]
}

///////////////////////////////////////////////////////////

type BoolOption struct {
    // all options belonging to this BoolOption
    options []string
    // value true if set
    value bool
}

func newBoolOption(op []string) *BoolOption {
    b := new(BoolOption)
    b.options = op
    b.value = false
    return b
}

func (b *BoolOption) isSet() bool {
    return b.value
}

func (b *BoolOption) reset() {
    b.value = false
}

func (b *BoolOption) setFlag() {
    b.value = !b.value // flip the bool switch
}

func (b *BoolOption) String() string {
    return "BoolOption " + b.options[0]
}
