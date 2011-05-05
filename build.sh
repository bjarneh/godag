#!/bin/bash
# Copyright (C) 2009 all rights reserved 
# GNU GENERAL PUBLIC LICENSE VERSION 3.0
# Author bjarneh@ifi.uio.no

COMPILER=""
LINKY=""
D=`dirname "$0"`
B=`basename "$0"`
FULL="`cd \"$D\" 2>/dev/null && pwd || echo \"$D\"`/$B"
HERE=$(dirname "$FULL")
IDIR=$HERE/src
CPROOT=`date +"tmp-pkgroot-%s"`
SRCROOT="$GOROOT/src/pkg"
UP_ONE=""

# array to store packages which are pure go
declare -a package;

# this is done statically for now, no grepping
# to figure out which packges are actually pure go..
package=(
'archive'
'compress'
'container'
'flag'
'go'
'html'
'http'
'image'
'mime'
'patch'
'rpc'
'strconv'
'tabwriter'
'template'
'io'
## packages above this line cannot be tested without modification
'asn1'
'bufio'
'cmath'
'ebnf'
'encoding'
'expvar'
'fmt'
'gob'
'hash'
'index'
'json'
'log'
'netchan'
'rand'
'reflect'
'regexp'
'scanner'
'smtp'
'sort'
'strings'
'syslog'
'testing'
'try'
'unicode'
'unsafe'
'utf16'
'utf8'
'websocket'
'xml'
)


function build(){
    echo -n "build "
    cd src/utilz && $COMPILER walker.go || exit 1
    $COMPILER handy.go || exit 1
    $COMPILER stringset.go || exit 1
    $COMPILER stringbuffer.go || exit 1
    $COMPILER global.go || exit 1
    $COMPILER timer.go || exit 1
    $COMPILER say.go || exit 1
    cd $HERE/src/parse && $COMPILER -o gopt.$OBJ option.go gopt.go || exit 1
    cd $HERE/src/cmplr && $COMPILER -I $IDIR dag.go || exit 1
    $COMPILER -I $IDIR compiler.go || exit 1
    cd $HERE/src/start && $COMPILER -I $IDIR main.go || exit 1
    cd $HERE && $LINKY -o gd -L src src/start/main.? || exit 1
    echo "...done"
}

function clean(){
    echo -n "clean"
    cd $HERE
    rm -rf src/utilz/walker.?
    rm -rf src/utilz/stringset.?
    rm -rf src/utilz/stringbuffer.?
    rm -rf src/utilz/global.?
    rm -rf src/utilz/utilz_test.?
    rm -rf src/utilz/handy.?
    rm -rf src/utilz/timer.?
    rm -rf src/utilz/say.?
    rm -rf src/cmplr/dag.?
    rm -rf src/cmplr/compiler.?
    rm -rf src/parse/gopt.?
    rm -rf src/parse/gopt_test.?
    rm -rf src/parse/option.?
    rm -rf src/start/main.?
    rm -rf gd
    rm -rf "$HOME/bin/gd"
    rm -rf "$GOBIN/gd"
    echo " ...done"
}

function move(){
    cd "$HERE"
    if [ -f "gd" ]; then
        echo -n "move"
        if [ -d "${HOME}/bin" ]; then
            cd "$HERE"
            mv gd "$HOME/bin"
        else
            if [ -d "$GOBIN" ]; then
                cd "$HERE"
                mv gd "$GOBIN"
            else
                echo -e "\n[ERROR] \$HOME/bin: not a directory"
                echo -e "[ERROR] \$GOBIN   : not set\n"
            fi
        fi
        echo "  ...done"
    else
        echo "'gd' not found, nothing to move"
        exit 1
    fi
}

function phelp(){
cat <<EOH

build.sh - utility script for godag

targets:

  help    : print this menu and exit
  clean   : rm *.[865a] from src + rm gd \$HOME/bin/gd \$GOBIN/gd
  build   : compile source code in ./src
  move    : move 'gd' to \$HOME/bin (\$GOBIN fallback)
  install : clean + build + move (DEFAULT)
  cproot  : copy modified (pure go) part of \$GOROOT/src/pkg
  stdlib  : copy original (pure go) part of \$GOROOT/src/pkg
  testok  : copy partial stdlib that can be tested without modification
  debian  : build a debian package (godag_0.2-0_${GOARCH}.deb)

EOH
}

function die(){
    echo "$1"
    exit 1
}


# recursively copy all the $GOROOT/src/pkg to $CPROOT,
# with a *.go filter, any test that includes testdata will fail.
# NOTE main packages are also removed, these are used for testing
# and since too many of these end up in the same name-space, they
# are all removed here..
function recursive_copy(){

    mkdir "$2"

    for i in $(ls "$1");
    do
        if [ -f "$1/$i" ]; then
            case "$i" in *.go)
                grep "^package main$" -q "$1/$i" || cp "$1/$i" "$2/$i"
            esac
        fi

        if [ -d "$1/$i" ]; then
            if [ ! "$i" == "testdata" ];then
                recursive_copy "$1/$i" "$2/$i"
            fi
        fi
    done

    return 1
}


# move all go packages up one level, and give them
# a fitting header based on directory..
function up_one_level(){

    for element in $(ls $1);
    do
        if [ -f "$1/$element" ]; then
            mv "$1/$element" "${1}/${2}_${element}"
            mv "${1}/${2}_${element}" "${1}/.."
        fi

        if [ -d "$1/$element" ]; then
            up_one_level "$1/$element" "$element"
        fi

    done

    return 1
}

function cproot(){

    mkdir "$CPROOT";
    echo "cp *.go: \$GOROOT/src/pkg  ->  $CPROOT"
    echo "this may take some time..."

    for p in "${package[@]}";
    do
        recursive_copy "$SRCROOT/$p" "$CPROOT/$p"
    done

    if [ "$UP_ONE" ];then
        up_one_level "$CPROOT" "$CPROOT"
    fi

    # delete empty directories from $CPROOT
    find -depth -type d -empty -exec rmdir {} \;

    exit 0
}

function testok(){
    cnt=15
    for((i = 0; i < cnt; i++))
    do
        unset package[$i]
    done

    cproot
}

# this target does not show up in help message :-)
function rmstdlib(){
    for p in "${package[@]}";
    do
        echo "rm -rf ${GOROOT}/pkg/${GOOS}_${GOARCH}/${p}.*"
        rm -rf "${GOROOT}/pkg/${GOOS}_${GOARCH}/${p}".*
    done
}

# default target clean + build + move
function triple(){
    clean
    build
    move
}

# Make sure we have all binaries needed in order to build debian package
function sanity(){
    pathfind 'hg'       || die "[ERROR] missing 'hg (mercurial)'"
    pathfind 'gzip'     || die "[ERROR] missing 'gzip'"
    pathfind 'md5sum'   || die "[ERROR] missing 'md5sum'"
    pathfind 'dpkg-deb' || die "[ERROR] missing 'dpkg-deb'"
    pathfind 'fakeroot' || die "[ERROR] missing 'fakeroot'"
    # Not too many systems lacking these (coreutils) but still
    pathfind 'cp'       || die "[ERROR] missing 'cp'"
    pathfind 'mkdir'    || die "[ERROR] missing 'mkdir'"
    pathfind 'mv'       || die "[ERROR] missing 'mv'"
    pathfind 'du'       || die "[ERROR] missing 'du'"
    pathfind 'chmod'    || die "[ERROR] missing 'chmod'"
    pathfind 'printf'   || die "[ERROR] missing 'printf'"
}

# Taken from Debian Developers Reference Chapter 6
function pathfind(){
     OLDIFS="$IFS"
     IFS=:
     for p in $PATH; do
         if [ -x "$p/$*" ]; then
             IFS="$OLDIFS"
             return 0
         fi
     done
     IFS="$OLDIFS"
     return 1
}

# create a debian package of godag, not the prettiest function..
function debian() {

# do you have what it takes to build a deb?

    sanity

# NOTE: does not depend on anything yet :-)
DEBCONTROL="Package: godag
Version: 0.2
Section: devel
Priority: optional
Architecture: %s
Depends:
Suggests: gccgo,golang
Conflicts:
Replaces:
Installed-Size: %d
Maintainer: Bjarne Holen <bjarneh@ifi.uio.no>
Description: Golang/Go compiler front-end.
 Godag automatically builds projects written in golang,
 by inspecting source-code imports to calculate compile order.
 Unit-testing, formatting and installation of external 
 libraries are also automated. The default back-end is gc,
 other back-ends have partial support: gccgo, express.
"

DEBCOPYRIGHT="Name: godag
Maintainer: bjarne holen <bjarneh@ifi.uio.no>
Source: https://godag.googlecode.com/hg

Files: *
Copyright: 2011, bjarne holen <bjarneh@ifi.uio.no>
License: GPL-3

License: GPL-3
On Debian systems, the complete text of the GNU General Public License
version 3 can be found in '/usr/share/common-licenses/GPL-3'.
"

DEBCHANGELOG="godag (0.2.0) devel; urgency=low

    * The actual changelog can be found in changelog...

 -- Bjarne Holen <bjarneh@ifi.uio.no>  Thu, 05 May 2011 14:07:28 -0400
"

    if [ "$GOARCH" = "386" ]; then
        DEBARCH="i386"
    else
        DEBARCH="$GOARCH"
    fi

    build

    mkdir -p ./debian/DEBIAN
    mkdir -p ./debian/usr/bin
    mkdir -p ./debian/usr/share/man/man1
    mkdir -p ./debian/usr/share/doc/godag
    mkdir -p ./debian/etc/bash_completion.d

    mv gd ./debian/usr/bin
    cp ./util/gd-completion.sh ./debian/etc/bash_completion.d/gd
    cat ./util/gd.1 | gzip --best - > ./debian/usr/share/man/man1/gd.1.gz
    cat ./util/gd.1 | gzip --best - > ./debian/usr/share/man/man1/godag.1.gz
    echo "$DEBCOPYRIGHT" > ./debian/usr/share/doc/godag/copyright
    if [ -d ".hg" ];then
        hg log | gzip --best - > ./debian/usr/share/doc/godag/changelog.gz
    else # we are git
        git log | gzip --best - > ./debian/usr/share/doc/godag/changelog.gz
    fi
    echo "$DEBCHANGELOG" | gzip --best - > ./debian/usr/share/doc/godag/changelog.Debian.gz
    arr=($(du -s ./debian))
    printf "$DEBCONTROL" "$DEBARCH" "${arr[0]}" > ./debian/DEBIAN/control
    echo "/etc/bash_completion.d/gd" > ./debian/DEBIAN/conffiles
    fakeroot dpkg-deb --build ./debian
    mv debian.deb "godag_0.2-0_$DEBARCH.deb"
    rm -rf ./debian

}

# main
{
[ "$GOROOT" ] || die "[ERROR] missing \$GOROOT"
[ "$GOARCH" ] || die "[ERROR] missing \$GOARCH"
[ "$GOOS" ]   || die "[ERROR] missing \$GOOS"
[ "$GOBIN" ]  || die "[ERROR] missing \$GOBIN"

case "$GOARCH" in
    '386')
    COMPILER="8g"
    LINKY="8l"
	OBJ="8"
    ;;
    'arm')
    COMPILER="5g"
    LINKY="5l"
	OBJ="5"
    ;;
    'amd64')
    COMPILER="6g"
    LINKY="6l"
	OBJ="6"
    ;;
    *)
    echo "architecture not: 'amd64' '386' 'arm'"
    echo "architecture was ${GOARC}"
    exit 1
    ;;
esac


case "$1" in
     'help' | '-h' | '--help' | '-help')
      phelp
      ;;
      'cproot' | '--cproot' | '-cproot')
      UP_ONE="yes"
      time cproot
      ;;
      'stdlib' | '-stdlib' | '--stdlib')
      time cproot
      ;;
      'testok' | '-testok' | '--testok')
      time testok
      ;;
      'clean' | 'c' | '-c' | '--clean' | '-clean')
      time clean
      ;;
      'build' | 'b' | '-b' | '--build' | '-build')
      time build
      ;;
      'move' | 'm' | '-m' | '--move' | '-move')
      time move
      ;;
      'del' | '--del' | '-del')
      rmstdlib
      ;;
      'debian' | '--debian' | '-debian')
      debian
      ;;
      *)
      time triple
      ;;
esac
}
