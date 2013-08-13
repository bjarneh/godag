#!/bin/bash
#
#  Copyright (C) 2009 bjarneh
#
#  This program is free software: you can redistribute it and/or modify
#  it under the terms of the GNU General Public License as published by
#  the Free Software Foundation, either version 3 of the License, or
#  (at your option) any later version.
#
#  This program is distributed in the hope that it will be useful,
#  but WITHOUT ANY WARRANTY; without even the implied warranty of
#  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#  GNU General Public License for more details.
#
#  You should have received a copy of the GNU General Public License
#  along with this program.  If not, see <http://www.gnu.org/licenses/>.
#
# -----------------------------------------------------------------------
#
#  Bash command completion file for godag
# 
#  Add this file somewhere where it gets sourced.
#  If you have sudo power it can be dropped into
#  /etc/bash_completion.d/. If not it can be sourced
#  by one of your startup scripts (.bashrc .profile ...)
#

_gd(){

    local cur prev opts gd_long_opts gd_short_opts gd_short_explain gd_special
    # long options
    gd_long_opts="--help --version --list --print --sort --output --static --gdmk --dryrun --clean --quiet --lib --main --dot --test --bench --match --verbose --fmt --rewrite --tab --tabwidth --external --update-external --backend --test-bin --test.short --test.v --test.bench --test.benchtime --test.cpu --test.cpuprofile --test.memprofile --test.memprofilerate --test.timeout --strip"
    # short options + explain
    gd_short_explain="-h[--help] -v[--version] -l[--list] -p[--print] -s[--sort] -o[--output] -S[--static] -g[--gdmk] -d[--dryrun] -c[--clean] -q[--quiet] -L[--lib] -M[--main] -D[--dot] -I -t[--test] -b[--bench] -m[--match] -V[--verbose] -f[--fmt] -r[--rewrite] -T[--tab] -w[--tabwidth] -e[--external] -u[--update--external]  -B[--backend] -y[--strip]"
    # short options
    gd_short_opts="-h -v -l -p -s -o -S -g -d -c -q -L -M -D -I -t -b -m -V -f -r -T -w -e -u -B -y"

    gd_special="clean test help fmt strip print dryrun list"

    COMPREPLY=()

    prev="${COMP_WORDS[COMP_CWORD-1]}"
    cur="${COMP_WORDS[COMP_CWORD]}"

    if [[ "${cur}" == --* ]]; then
        COMPREPLY=( $(compgen -W "${gd_long_opts}" -- "${cur}") )
        return 0
    fi

    if [[ "${cur}" == -* ]]; then
        COMPREPLY=( $(compgen -W "${gd_short_opts}" -- "${cur}") )
        if [ "${#COMPREPLY[@]}" -gt 1 ]; then
            COMPREPLY=( $(compgen -W "${gd_short_explain}" -- "${cur}") )
        fi
        return 0
    fi

    case "${cur}" in
        c* | t* | h* | f* | s* | p* | d* | l*)
          COMPREPLY=( $(compgen -W "${gd_special}" -- "${cur}") )
          if [ "${#COMPREPLY[@]}" -gt 1 ]; then
              return 0
          fi
        ;;
    esac

    # use godag to parse makefile and look for targets
    if [[ "${prev}" == "mk.go" ]]; then
        if [ "${COMP_CWORD}" -eq 2 ]; then
            TARGETS=$(gd -mkcomplete mk.go)
            COMPREPLY=( $(compgen -W "$TARGETS" -- "${cur}" ) )
            return 0
        fi
    fi

    if [[ "${prev}" == -* ]]; then
        case "${prev}" in
            '-B' |'-B=' | '-backend' |'--backend' |'-backend=' |'--backend=')
                COMPREPLY=( $(compgen -W "gc gccgo express" -- "${cur}") )
                return 0
                ;;
        esac
    fi
}

## directories only -d was a bit to strict
complete -o default -F _gd gd
