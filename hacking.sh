#!/bin/bash
# Author: Eli Qiao <qiaoliyong@gmail.com>

echo "calling go fmt"
go fmt

files=$(git diff HEAD~1 --stat | awk '{if ($1 ~ /\.go$/) {print $1}}')

arr=($files)

ret=0
echo "calling golint on all changed files"
for f in "${arr[@]}"
do
    echo "calling golint $f ..."
    rev=$(golint "$f")
    if [[ ! -z $rev ]]; then
        ret=-1
    fi
done

if [[ $ret -ne 0 ]]; then
    echo ":( <<< Please address coding style mistakes."
else
    echo ":) >>> No errors for coding style"
fi
