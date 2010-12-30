// © Knug Industries 2009 all rights reserved
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package walker /* texas ranger */

import (
    "os"
    "path"
)

// This package creates a list of strings from a path,
// where pathnames to all files below path given, is
// returned as a StringVector. Unwanted directories and
// files can be filtered out using the two filter functions
// IncludeDir and IncludeFile.


// reassign to filter pathwalk
var IncludeDir = func(p string) bool { return true }
var IncludeFile = func(p string) bool { return true }

type collect struct {
    Files []string
}

func newCollect() *collect {
    c := new(collect)
    c.Files = make([]string, 0)
    return c
}

func (c *collect) VisitDir(path string, d *os.FileInfo) bool {
    return IncludeDir(path)
}

func (c *collect) VisitFile(path string, d *os.FileInfo) {
    if IncludeFile(path) {
        c.Files = append(c.Files, path)
    }
}

func PathWalk(root string) []string {
    c := newCollect()
    errs := make(chan os.Error)
    path.Walk(root, c, errs)
    return c.Files
}
