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
