// © Knug Industries 2009 all rights reserved 
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package walker /* texas ranger */

import(
    "os";
    "container/vector";
    "path";
)

// reassign to filter pathwalk
var IncludeDir  = func(p string) bool{ return true; }
var IncludeFile = func(p string) bool{ return true; }

type collect struct{
    Files *vector.StringVector;
};

func newCollect() *collect{
    c := new(collect);
    c.Files = new(vector.StringVector);
    return c;
}

func (c *collect) VisitDir (path string, d *os.FileInfo) bool{
    return IncludeDir(path);
}

func (c *collect) VisitFile(path string, d *os.FileInfo){
    if IncludeFile(path) {
        c.Files.Push(path);
    }
}

func PathWalk(root string) *vector.StringVector{
    c    :=  newCollect();
    errs :=  make(chan os.Error);
    path.Walk(root, c, errs);
    return c.Files;
}
