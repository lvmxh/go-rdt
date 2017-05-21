#!/bin/bash

GOPATH=`go env GOPATH`
ALGORITHM_PLUGINS_DIR="../"

go tool vet ./

for i in $ALGORITHM_PLUGINS_DIR*/*.go
do
    algorithms=${i##$ALGORITHM_PLUGINS_DIR}
    plugin=${algorithms%%/*}
    source=${i##*/}

    if [ ${source%.go} != $plugin ]; then
        echo "$source is not plugin main source, ignore it."
        continue
    fi

    target=${i/%go/so}

    # FIXME, maybw we shoule move the plguin bin file to $GOPATH/bin
    go build -buildmode=plugin -o $target $i
    if [ $? != 0 ] ; then
        echo "Error to build plugin: $i"
        echo "Make sure your go-lang version is above v1.8"
    fi

done
