//  Copyright © 2011 bjarneh
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

package say // Perl6 inspiration here :-)

import (
    "fmt"
    "io"
)

// package to turn of all print statements

var mute bool = false

func Mute() {
    mute = true
}

func Sound() {
    mute = false
}

func Print(args ...interface{}) (int, error) {
    if !mute {
        return fmt.Print(args...)
    }
    return 0, nil
}

func Println(args ...interface{}) (int, error) {
    if !mute {
        return fmt.Println(args...)
    }
    return 0, nil
}

func Printf(f string, args ...interface{}) (int, error) {
    if !mute {
        return fmt.Printf(f, args...)
    }
    return 0, nil
}

func Fprint(w io.Writer, args ...interface{}) (int, error) {
    if !mute {
        return fmt.Fprint(w, args...)
    }
    return 0, nil
}

func Fprintln(w io.Writer, args ...interface{}) (int, error) {
    if !mute {
        return fmt.Fprintln(w, args...)
    }
    return 0, nil
}

func Fprintf(w io.Writer, f string, args ...interface{}) (int, error) {
    if !mute {
        fmt.Fprintf(w, f, args...)
    }
    return 0, nil
}
