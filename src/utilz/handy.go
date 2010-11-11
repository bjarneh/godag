// © Knug Industries 2009 all rights reserved
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package handy

import (
    "os"
    "log"
)



// some utility functions

func StdExecve(argv []string, stopOnTrouble bool) (ok bool) {

    var fdesc []*os.File
    ok = true

    fdesc = make([]*os.File, 3)
    fdesc[0] = os.Stdin
    fdesc[1] = os.Stdout
    fdesc[2] = os.Stderr

    pid, err := os.ForkExec(argv[0], argv, os.Environ(), "", fdesc)

    if err != nil {
        if stopOnTrouble {
            log.Exitf("[ERROR] %s\n", err)
        } else {
            log.Printf("[ERROR] %s\n", err)
        }
        ok = false

    } else {

        wmsg, werr := os.Wait(pid, 0)

        if werr != nil || wmsg.WaitStatus.ExitStatus() != 0 {

            if werr != nil {
                log.Printf("[ERROR] %s\n", werr)
            }

            if stopOnTrouble {
                os.Exit(1)
            }

            ok = false
        }
    }

    return ok
}


// more or less stolen from a pastebin posted on #go-nuts
// http://pastebin.com/V0CULJWt by yiyus
// looked kind of handy, so it was placed here :-)
func Fopen(name, mode string, perms uint32) (file *os.File, err os.Error) {

    var imode int // int mode

    switch mode {
        case "r"   : imode = os.O_RDONLY
        case "r+"  : imode = os.O_RDWR
        case "w"   : imode = os.O_WRONLY|os.O_CREAT|os.O_TRUNC
        case "w+"  : imode = os.O_RDWR|os.O_CREAT|os.O_TRUNC
        case "a"   : imode = os.O_WRONLY|os.O_CREAT|os.O_APPEND
        case "a+"  : imode = os.O_RDWR|os.O_CREAT|os.O_APPEND
        default    : panic("Fopen: illegal mode -> "+mode)
    }

    return os.Open(name, imode, perms) // 0644 default
}
