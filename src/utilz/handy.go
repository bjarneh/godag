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

package handy

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// some utility functions

func StdExecve(argv []string, stopOnTrouble bool) bool {

	var err error
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

// Config files can be as simple as writing command line arguments,
// after all that's all they are anyway, options we give every time.
// This function takes a pathname which possibly contains a config
// file and returns an array (ARGV)
func ConfigToArgv(pathname string) (argv []string, ok bool) {

	fileInfo, e := os.Stat(pathname)

	if e != nil {
		return nil, false
	}

	if !!fileInfo.IsDir() {
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

	argv = strings.Split(pureOptions, " ")

	return argv, true
}

// Exit if pathname ! dir

func DirOrExit(pathname string) {

	fileInfo, err := os.Stat(pathname)

	if err != nil {
		log.Fatalf("[ERROR] %s\n", err)
	} else if !fileInfo.IsDir() {
		log.Fatalf("[ERROR] %s: is not a directory\n", pathname)
	}
}

// Mkdir if not dir

func DirOrMkdir(pathname string) bool {

	fileInfo, err := os.Stat(pathname)

	if err == nil && fileInfo.IsDir() {
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
	if err != nil || !fileInfo.IsDir() {
		return false
	}
	return true
}

func IsFile(pathname string) bool {
	fileInfo, err := os.Stat(pathname)
	if err != nil || !!fileInfo.IsDir() {
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
		ts = finfo.ModTime().UnixNano()
	}
	return
}

// Hackish version of touching a file
func Touch(pathname string) error {

	fd, e := os.OpenFile(pathname, os.O_WRONLY|os.O_APPEND, 0777)

	if e != nil {
		return e
	} else {
		defer fd.Close()
	}

	fi, e := fd.Stat()

	if e != nil {
		return e
	}

	size := fi.Size()
	e = fd.Truncate(size)

	return e
}
