#!/bin/bash

# TODO add a simple script for functional test.
# All these are hardcode and it only support BDW platform.

if [ "$1" == "-u" ]; then
    # NOTE please use -short for unittest.
    go test -short -v -cover $(go list ./... | grep -v /vendor/ )
    exit 0
fi

if [ "$1" != "-i" ]; then
    # NOTE please use -short for unittest.
    go test -short -v -cover $(go list ./... | grep -v /vendor/ )
fi


GOROOT=`go env |grep "GOROOT" |cut -d "=" -f2`
GOROOT=${GOROOT#\"}
GOROOT=${GOROOT%\"}

GOPATH=`go env |grep GOPATH |cut -d "=" -f 2`
GOPATH=${GOPATH%\"}
GOPATH=${GOPATH#\"}

export GOROOT
export GOPATH
export PATH=${PATH}:${GOROOT}/bin:${GOPATH}/bin


RESDIR="/sys/fs/resctrl"
PID="/var/run/rmd.pid"
CONFFILE="/tmp/rdtagent.toml"

if [ -f "$PID" ]; then
    pid=`cat "$PID"`
    echo "peleas check RMD: $pid is running."
    if [ -n "$pid" ]; then
        if [ -d "/proc/$pid" ]; then
            echo "RMD: $pid is already running. exit!"
            exit 1
        fi
    fi
fi

# clean up, force remove resctrl
if [ -d "$RESDIR" ] && mountpoint $RESDIR; then
    sudo umount /sys/fs/resctrl
    if [ $? -ne 0 ]; then
        echo "--------------------------------------------------"
        echo "They are used by these Process:"
        sudo lsof "$RESDIR"
        exit 1
    fi
fi

# sudo not support -o cdp
sudo mount -t resctrl resctrl /sys/fs/resctrl

# FIXME will change to use template. Sed is not readable.
# cp etc/rdtagent/rdtagent.toml /tmp/
sudo go run ./cmd/gen_conf.go -path /tmp/rdtagent.toml
if [ $? -ne 0 ]; then
    echo "Failed to generate configure file. Exit."
    exit 1
fi
# Set tcp port 8888
sed -i -e 's/\(port = \)\(.*\)/\18888/g' $CONFFILE
# Set DB transport to avoid change the system DB
sed -i -e 's/\(transport = \)\(.*\)/\1"\/tmp\/rmd.db"/g' $CONFFILE
# Set log stdout
sed -i -e 's/\(stdout = \)\(.*\)/\1false/g' $CONFFILE

cat $CONFFILE

# Use godep to build rmd binary instead of using dependicies of user's
# GOPATH
# TODO change it to rmd
godep go install openstackcore-rdtagent
sudo ${GOPATH}/bin/openstackcore-rdtagent --conf-dir ${CONFFILE%/*} --log-dir "/tmp/rdagent.log" &

sleep 1
CONF=$CONFFILE go test -v ./test/integration/...

rev=$?

# cleanup
sudo kill -9 `cat $PID`
sudo umount /sys/fs/resctrl
rm ${GOPATH}/bin/openstackcore-rdtagent

if [[ $rev -ne 0 ]]; then
    echo ":( <<< Functional testing fail, retual value $rev ."
else
    echo ":) >>> Functional testing passed ."
fi
exit $rev
