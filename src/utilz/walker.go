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

package walker /* texas ranger */

import (
    "os"
    "path/filepath"
)

// This package does something along the lines of: find PATH -type f
// Filters can be added on both directory and filenames in order to filter
// the resulting slice of pathnames.

// reassign to filter pathwalk
var IncludeDir = func(p string) bool { return true }
var IncludeFile = func(p string) bool { return true }

func PathWalk(root string) (files []string) {

    fn := func(p string, d os.FileInfo, e error) error {

        if d.IsDir() && !IncludeDir(p) {
            return filepath.SkipDir
        }

        if !d.IsDir() && IncludeFile(p) {
            files = append(files, p)
        }

        return e
    }

    filepath.Walk(root, fn)

    return files
}

func helper(root string, ch chan string) {

    fn := func(p string, d os.FileInfo, e error) error {

        if d.IsDir() && !IncludeDir(p) {
            return filepath.SkipDir
        }

        if !d.IsDir() && IncludeFile(p) {
            ch <- p
        }

        return e
    }

    filepath.Walk(root, fn)

    close(ch)
}

func ChanWalk(root string) (files chan string) {
    ch := make(chan string)
    go helper(root, ch)
    return ch
}
