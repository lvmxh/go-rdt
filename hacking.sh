#!/bin/bash
# Author: Eli Qiao <qiaoliyong@gmail.com>

ret=$?
files=$(git diff HEAD --stat | awk '{if ($1 ~ /\.go$/) {print $1}}')
arr=($files)
# hacking command list
# don't check shift
cmds=("golint" "go fmt" "go tool vet -shift=false")

function hack {
    local cmd=$1
    local file=$2
    local rev

    rev=$($cmd "$file")
    if [[ ! -z $rev ]]; then
        echo "$rev"
    else
        echo "0"
    fi
}

ret=0

for f in "${arr[@]}"
do
    if [ -f "${f}" ]; then
        for ((i = 0; i < ${#cmds[@]}; i++)) do

            echo "checking "${cmds[$i]}" $f ..."
            rev=$(hack "${cmds[$i]}" $f)

            if [[ ! -z ${rev} && ${rev} != "0" ]]; then
                echo ${rev}
                ret=-1
            fi
        done
    fi
done

if [[ $ret -ne 0 ]]; then
    echo ":( <<< Please address coding style mistakes."
else
    echo ":) >>> No errors for coding style"
fi

exit $ret
