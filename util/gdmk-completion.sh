#!/bin/bash
#
# Bash command completion file for gdmk
#
# Add this file somewhere where it gets sourced.
# If you have sudo power it can be dropped into
# /etc/bash_completion.d/. If not it can be sourced
# by one of your startup scripts (.bashrc .profile ...)
#
#  Copyright (C) 2011 bjarneh
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
