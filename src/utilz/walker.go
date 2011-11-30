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

    fn := func(p string, d *os.FileInfo, e error) error {

        if d.IsDirectory() && !IncludeDir(p) {
            return filepath.SkipDir
        }

        if d.IsRegular() && IncludeFile(p) {
            files = append(files, p)
        }

        return e
    }

    filepath.Walk(root, fn)

    return files
}

//////////////////////////////////TODO: fix the chan-walker
/// type collect struct {
/// 	files []string
/// }
/// 
/// func newCollect() *collect {
/// 	c := new(collect)
/// 	c.files = make([]string, 0)
/// 	return c
/// }
/// 
/// func (c *collect) VisitDir(path string, d *os.FileInfo) bool {
/// 	return IncludeDir(path)
/// }
/// 
/// func (c *collect) VisitFile(path string, d *os.FileInfo) {
/// 	if IncludeFile(path) {
/// 		c.files = append(c.files, path)
/// 	}
/// }

/// func PathWalk(root string) []string {
/// 	c := newCollect()
/// 	errs := make(chan error)
/// 	filepath.Walk(root, c, errs)
/// 	return c.files
/// }

// ChanWalk is a type of PathWalk which returns immediately and
// spits out path-names through a channel, it requires a new
// type; this is it :-)

/// type chanCollect struct {
/// 	files chan string
/// }
/// 
/// func newChanCollect() *chanCollect {
/// 	c := new(chanCollect)
/// 	c.files = make(chan string)
/// 	return c
/// }
/// 
/// func (c *chanCollect) VisitDir(path string, d *os.FileInfo) bool {
/// 	return IncludeDir(path)
/// }
/// 
/// func (c *chanCollect) VisitFile(path string, d *os.FileInfo) {
/// 	if IncludeFile(path) {
/// 		c.files <- path
/// 	}
/// }
/// 
/// func helper(root string, cc *chanCollect) {
/// 	errs := make(chan error)
/// 	filepath.Walk(root, cc, errs)
/// 	close(cc.files)
/// }
/// 
/// // Same as PathWalk part from returning path names in a channel,
/// // note that this function returns immediatlely, most likely this is
/// // what you want unless you need all path names at once..
/// func ChanWalk(root string) chan string {
/// 	cc := newChanCollect()
/// 	go helper(root, cc)
/// 	return cc.files
/// }
