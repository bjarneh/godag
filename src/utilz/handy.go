// © Knug Industries 2009 all rights reserved
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package handy

import (
    "os"
    "log"
    "io/ioutil"
    "regexp"
    "strings"
    "exec"
)


// some utility functions

func StdExecve(argv []string, stopOnTrouble bool) bool {

    var err os.Error
    var cmd *exec.Cmd

    switch len(argv) {
    case 0:
        if stopOnTrouble {
            log.Fatalf("[ERROR] len(argv) == 0\n")
        }
        return false
    case 1:
        cmd = exec.Command(argv[0])
    default:
        cmd = exec.Command(argv[0], argv[1:]...)
    }

    // pass-through
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Stdin = os.Stdin

    err = cmd.Start()

    if err != nil {
        if stopOnTrouble {
            log.Fatalf("[ERROR] %s\n", err)
        } else {
            log.Printf("[ERROR] %s\n", err)
            return false
        }
    }

    err = cmd.Wait()

    if err != nil {
        if stopOnTrouble {
            log.Fatalf("[ERROR] %s\n", err)
        } else {
            log.Printf("[ERROR] %s\n", err)
            return false
        }
    }

    return true
}


// More or less taken from a pastebin posted on #go-nuts
// http://pastebin.com/V0CULJWt by yiyus
// looked kind of handy, so it was placed here :-)
func Fopen(name, mode string, perms uint32) (*os.File, os.Error) {

    var imode int // int mode

    switch mode {
    case "r":
        imode = os.O_RDONLY
    case "r+":
        imode = os.O_RDWR
    case "w":
        imode = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
    case "w+":
        imode = os.O_RDWR | os.O_CREATE | os.O_TRUNC
    case "a":
        imode = os.O_WRONLY | os.O_CREATE | os.O_APPEND
    case "a+":
        imode = os.O_RDWR | os.O_CREATE | os.O_APPEND
    default:
        panic("Fopen: illegal mode -> " + mode)
    }

    return os.OpenFile(name, imode, perms) // 0644 default
}


// Config files can be as simple as writing command line arguments,
// after all that's all they are anyway, options we give every time.
// This function takes a pathname which possibly contains a config
// file and returns an array (ARGV)
func ConfigToArgv(pathname string) (argv []string, ok bool) {

    fileInfo, e := os.Stat(pathname)

    if e != nil {
        return nil, false
    }

    if !fileInfo.IsRegular() {
        return nil, false
    }

    b, e := ioutil.ReadFile(pathname)

    if e != nil {
        log.Print("[WARNING] failed to read config file\n")
        log.Printf("[WARNING] %s \n", e)
        return nil, false
    }

    comStripRegex := regexp.MustCompile("#[^\n]*\n?")
    blankRegex := regexp.MustCompile("[\n\t \r]+")

    rmComments := comStripRegex.ReplaceAllString(string(b), "")
    rmNewLine := blankRegex.ReplaceAllString(rmComments, " ")

    pureOptions := strings.TrimSpace(rmNewLine)

    if pureOptions == "" {
        return nil, false
    }

    argv = strings.Split(pureOptions, " ", -1)

    return argv, true
}

// Exit if pathname ! dir

func DirOrExit(pathname string) {

    fileInfo, err := os.Stat(pathname)

    if err != nil {
        log.Fatalf("[ERROR] %s\n", err)
    } else if !fileInfo.IsDirectory() {
        log.Fatalf("[ERROR] %s: is not a directory\n", pathname)
    }
}

// Mkdir if not dir

func DirOrMkdir(pathname string) bool {

    fileInfo, err := os.Stat(pathname)

    if err == nil && fileInfo.IsDirectory() {
        return true
    } else {
        err = os.MkdirAll(pathname, 0777)
        if err != nil {
            log.Fatalf("[ERROR] %s\n", err)
        }
    }
    return false
}

func IsDir(pathname string) bool {
    fileInfo, err := os.Stat(pathname)
    if err != nil || !fileInfo.IsDirectory() {
        return false
    }
    return true
}

func IsFile(pathname string) bool {
    fileInfo, err := os.Stat(pathname)
    if err != nil || !fileInfo.IsRegular() {
        return false
    }
    return true
}

func Delete(pathname string, die bool) (ok bool) {
    ok = true
    e := os.Remove(pathname)
    if e != nil {
        log.Printf("[ERROR]: %s\n", e)
        if die {
            os.Exit(1)
        }
        ok = false
    }
    return
}

func RmRf(pathname string, die bool) (ok bool) {
    ok = true
    e := os.RemoveAll(pathname)
    if e != nil {
        log.Printf("[ERROR]: %s\n", e)
        if die {
            os.Exit(1)
        }
        ok = false
    }
    return
}

func ModifyTimestamp(pathname string) (ts int64) {
    finfo, e := os.Stat(pathname)
    if e != nil {
        log.Fatalf("[ERROR]: %s - not a file\n", pathname)
    } else {
        ts = finfo.Mtime_ns
    }
    return
}
