#!/bin/bash
# Copyright (C) 2011 all rights reserved
# GNU GENERAL PUBLIC LICENSE VERSION 3.0
# Author bjarneh@ifi.uio.no
#
# Bash command completion file for gdmk
#
# Add this file somewhere where it gets sourced.
# If you have sudo power it can be dropped into
# /etc/bash_completion.d/. If not it can be sourced
# by one of your startup scripts (.bashrc .profile ...)

_gdmk(){

    local cur gdmk_targets gdmk_binary

    gdmk_binary="${PWD}/gdmk"

    if [ -x "$gdmk_binary" ];then
        gdmk_targets=$(${gdmk_binary} --list)
    else
        gdmk_targets=""
    fi

    COMPREPLY=()

    cur="${COMP_WORDS[COMP_CWORD]}"

    COMPREPLY=( $(compgen -W "${gdmk_targets}" -- "${cur}") )
}

complete -o default -F _gdmk gdmk
