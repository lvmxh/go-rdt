#!/bin/bash

# TODO add a simple script for functional test.
# All these are hardcode and it only support BDW platform.

if [ "$1" == "-u" ]; then
    # NOTE please use -short for unittest.
    go test -short -v -cover $(go list ./... | grep -v /vendor/ | grep -v /test/)
    exit 0
fi

if [ "$1" != "-i" -a "$1" != "-s" ]; then
    # NOTE please use -short for unittest.
    go test -short -v -cover $(go list ./... | grep -v /vendor/ | grep -v /test/)
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
# Set a unused random port
CHECK="do while"

while [[ ! -z $CHECK ]]; do
    PORT=$(( ( RANDOM % 60000 )  + 1025 ))
    CHECK=$(sudo netstat -ap | grep $PORT)
done

DATA=""
# TODO will also  support -data 'stdout=true,tasks=["ovs*","dpdk"]'
if [ "$1" == "-s" ]; then
    if [ "$2" == "-nocert" ]; then
        DATA="\"clientauth\":\"no\", \"tlsport\":$PORT"
    else
        DATA="\"tlsport\":$PORT"
    fi
else
    DATA="\"debugport\":$PORT"
fi

go run ./cmd/gen_conf.go -path /tmp/rdtagent.toml -data "{$DATA}"

if [ $? -ne 0 ]; then
    echo "Failed to generate configure file. Exit."
    exit 1
fi

cp etc/rdtagent/policy.toml /tmp/

# TODO need to remove these sed command.
# Set DB transport to avoid change the system DB
sed -i -e 's/\(transport = \)\(.*\)/\1"\/tmp\/rmd.db"/g' $CONFFILE
# Set log stdout
sed -i -e 's/\(stdout = \)\(.*\)/\1false/g' $CONFFILE

cat $CONFFILE

if [ "$1" == "-s" -a "$2" == "-nocert" ]; then

    # setup PAM files
    PAMSRCFILE="etc/rdtagent/pam/test/rmd"
    PAMDIR="/etc/pam.d"
    if [ -d $PAMDIR ]; then
        cp $PAMSRCFILE $PAMDIR
    fi

    # setup PAM test user
    BERKELYDBFILENAME="rmd_users.db"
    echo "user" >> users
    openssl passwd -crypt "user1" >> users
    db_load -T -t hash -f users "/tmp/"$BERKELYDBFILENAME
    if [ $? -ne 0 ]; then
        rm -rf users
        echo "Failed to setup pam files"
        exit 1
    fi
    rm -rf users
fi

# Use godep to build rmd binary instead of using dependicies of user's
# GOPATH
# TODO change it to rmd
godep go install openstackcore-rdtagent && sudo cp -r etc/rdtagent /etc
if [ $? -ne 0 ]; then
    echo "Failed to build rmd, please correct build issue."
    exit 1
fi

if [ "$1" == "-s" ]; then
    sudo ${GOPATH}/bin/openstackcore-rdtagent --conf-dir ${CONFFILE%/*} --log-dir "/tmp/rdagent.log" &
else
    sudo ${GOPATH}/bin/openstackcore-rdtagent --conf-dir ${CONFFILE%/*} --log-dir "/tmp/rdagent.log" --debug &
fi


sleep 1

if [ "$1" == "-s" ]; then
    if [ "$2" == "-nocert" ]; then
        CONF=$CONFFILE ginkgo -v -tags "integration_https" --focus="PAMAuth" ./test/integration_https/...
    else
        CONF=$CONFFILE ginkgo -v -tags "integration_https" --focus="CertAuth" ./test/integration_https/...
    fi
else
    CONF=$CONFFILE ginkgo -v -tags "integration" ./test/integration/...
fi

rev=$?

# cleanup
sudo kill -TERM `cat $PID`
sudo umount /sys/fs/resctrl
rm ${GOPATH}/bin/openstackcore-rdtagent

# cleanup PAM files
if [ "$1" == "-s" -a "$2" == "-nocert" ]; then
    rm -rf "/tmp/"$BERKELYDBFILENAME
    rm -rf $PAMDIR"/rmd"
fi

if [[ $rev -ne 0 ]]; then
    echo ":( <<< Functional testing fail, retual value $rev ."
else
    echo ":) >>> Functional testing passed ."
fi
exit $rev
