// © Knug Industries 2009 all rights reserved
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

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

type collect struct {
    files []string
}

func newCollect() *collect {
    c := new(collect)
    c.files = make([]string, 0)
    return c
}

func (c *collect) VisitDir(path string, d *os.FileInfo) bool {
    return IncludeDir(path)
}

func (c *collect) VisitFile(path string, d *os.FileInfo) {
    if IncludeFile(path) {
        c.files = append(c.files, path)
    }
}

func PathWalk(root string) []string {
    c := newCollect()
    errs := make(chan os.Error)
    filepath.Walk(root, c, errs)
    return c.files
}

// ChanWalk is a type of PathWalk which returns immediately and
// spits out path-names through a channel, it requires a new
// type; this is it :-)

type chanCollect struct {
    files chan string
}

func newChanCollect() *chanCollect {
    c := new(chanCollect)
    c.files = make(chan string)
    return c
}

func (c *chanCollect) VisitDir(path string, d *os.FileInfo) bool {
    return IncludeDir(path)
}

func (c *chanCollect) VisitFile(path string, d *os.FileInfo) {
    if IncludeFile(path) {
        c.files <- path
    }
}

func helper(root string, cc *chanCollect) {
    errs := make(chan os.Error)
    filepath.Walk(root, cc, errs)
    close(cc.files)
}

// Same as PathWalk part from returning path names in a channel,
// note that this function returns immediatlely, most likely this is
// what you want unless you need all path names at once..
func ChanWalk(root string) chan string {
    cc := newChanCollect()
    go helper(root, cc)
    return cc.files
}
