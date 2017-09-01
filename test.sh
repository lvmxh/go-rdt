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
if [ -d "$RESDIR" ]; then
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
cp etc/rdtagent/rdtagent.toml /tmp/
# Set tcp port 8088
sed -i -e 's/\(port = \)\(.*\)/\18088/g' $CONFFILE
# Set DB transport to avoid change the system DB
sed -i -e 's/\(transport = \)\(.*\)/\1"\/tmp\/rmd.db"/g' $CONFFILE
# Set log stdout
sed -i -e 's/\(stdout = \)\(.*\)/\1false/g' $CONFFILE
# Set OSGroup cacheways = 1
sed -i -e '/\[OSGroup\]*/{$!{N;s/\(cacheways = \)\(.*\)/\11/}}' $CONFFILE
# Set InfraGroup cacheways = 10
sed -i -e '/\[InfraGroup\]*/{$!{N;s/\(cacheways = \)\(.*\)/\110/}}' $CONFFILE
# set guarantee pool = 6
sed -i -e 's/\(guarantee = \)\(.*\)/\16/g' $CONFFILE
# set besteffort pool = 3
sed -i -e 's/\(besteffort = \)\(.*\)/\13/g' $CONFFILE
# set shared pool = 1
sed -i -e 's/\(shared = \)\(.*\)/\11/g' $CONFFILE

cat $CONFFILE

# Use godep to build rmd binary instead of using dependicies of user's
# GOPATH
# TODO change it to rmd
godep go install openstackcore-rdtagent
sudo ${GOPATH}/bin/openstackcore-rdtagent --conf-dir ${CONFFILE%/*} --log-dir "/tmp/rdagent.log" &

sleep 1
CONF=$CONFFILE go test -v ./test/integration/...

# cleanup
sudo kill -9 `cat $PID`
unmount /sys/fs/resctrl
rm ${GOPATH}/bin/openstackcore-rdtagent
