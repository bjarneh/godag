// © Knug Industries 2009 all rights reserved 
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package handy

import(
    "os";
    "strings";
    "fmt";
    "path";
)

// some utility functions


func StdExecve(argv []string, stopOnTrouble bool) (ok bool){

    var fdesc []*os.File;
    ok = true;

    fdesc = make([]*os.File, 3);
    fdesc[0] = os.Stdin;
    fdesc[1] = os.Stdout;
    fdesc[2] = os.Stderr;

    pid, err := os.ForkExec(argv[0], argv, os.Environ(), "", fdesc);

    if err != nil{
        fmt.Fprintf(os.Stderr, "[ERROR] %s\n",err);
        if stopOnTrouble { os.Exit(1); }
        ok = false;
    }else{
        wmsg, werr := os.Wait(pid, 0);
        if werr != nil || wmsg.WaitStatus != 0 {
            if werr != nil {
                fmt.Fprintf(os.Stderr, "[ERROR] %s\n", werr);
            }
            if stopOnTrouble { os.Exit(1); }
            ok = false;
        }
    }

    return ok;
}


func Which(cmd string) (string){

    var abspath string;
    var dir *os.FileInfo;
    var err os.Error;

    xpath := os.Getenv("PATH");
    dirs  := strings.Split(xpath, ":", 0);

    for i := range dirs {
        abspath = path.Join(dirs[i], cmd);
        dir, err = os.Stat(abspath);
        if err == nil{
            if dir.IsRegular(){
                if isExecutable(dir.Uid, dir.Permission()) {
                    return abspath;
                }
            }
        }
    }

    return "";
}

func isExecutable(uid int, perms int) bool {

    mode := 7;
    amode := (perms & mode);
    mode = mode << 6;
    umode := (perms & mode) >> 6;


    if amode == 7 || amode == 5 {
        return true;
    }

    if int(uid) == os.Getuid() {
        if umode == 7 || umode == 5 {
            return true;
        }
    }

    return false;
}
