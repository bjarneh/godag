#!/bin/bash
# Copyright (C) 2009 all rights reserved
# GNU GENERAL PUBLIC LICENSE VERSION 3.0
# Author bjarneh@ifi.uio.no
#
# Bash command completion file for godag
#
# Add this file somewhere where it gets sourced.
# If you have sudo power it can be dropped into
# /etc/bash_completion.d/. If not it can be sourced
# by one of your startup scripts (.bashrc .profile ...)

_gd(){

    local cur prev opts gd_long_opts gd_short_opts gd_short_explain gd_special
    # long options
    gd_long_opts="--help --version --list --print --sort --output --static --gdmk --dryrun --clean --quiet --lib --main --dot --test --bench --match --verbose --fmt --rewrite --tab --tabwidth --external --backend"
    # short options + explain
    gd_short_explain="-h[--help] -v[--version] -l[--list] -p[--print] -s[--sort] -o[--output] -S[--static] -g[--gdmk] -d[--dryrun] -c[--clean] -q[--quiet] -L[--lib] -M[--main] -D[--dot] -I -t[--test] -b[--bench] -m[--match] -V[--verbose] -f[--fmt] -r[--rewrite] -T[--tab] -w[--tabwidth] -e[--external] -B[--backend]"
    # short options
    gd_short_opts="-h -v -l -p -s -o -S -g -d -c -q -L -M -D -I -t -b -m -V -f -r -T -w -e -B"

    gd_special="clean test help $(ls)"

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
    if [[ "${cur}" == c* || "${cur}" == t* || "${cur}" == h* ]]; then
        COMPREPLY=( $(compgen -W "${gd_special}" -- "${cur}") )
    fi
    if [[ "${prev}" == -* ]]; then
        case "${prev}" in
            '-B' | '-B=' | '-backend' | '--backend' | '-backend=' | '--backend=')
                COMPREPLY=( $(compgen -W "gc gccgo express" -- "${cur}") )
                return 0
                ;;
        esac
    fi
}

## directories only -d was a bit to strict
complete -o default -F _gd gd
