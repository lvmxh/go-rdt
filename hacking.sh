#!/bin/bash
# Author: Eli Qiao <qiaoliyong@gmail.com>

echo "calling go fmt"
go fmt

ret=$?
files=$(git diff HEAD --stat | awk '{if ($1 ~ /\.go$/) {print $1}}')

arr=($files)

echo "calling golint on all changed files"
for f in "${arr[@]}"
do
    if [ -f "${f}" ]; then
        echo "calling golint $f ..."
        rev=$(golint "$f")
        if [[ ! -z $rev ]]; then
            echo "$rev"
            ret=-1
        fi
    fi
done

if [[ $ret -ne 0 ]]; then
    echo ":( <<< Please address coding style mistakes."
else
    echo ":) >>> No errors for coding style"
fi

exit $ret
