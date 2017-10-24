#!/usr/bin/env bash

PAMSRCFILE="etc/rdtagent/pam/rmd"
PAMDIR="/etc/pam.d"
if [ -d $PAMDIR ]; then
    cp $PAMSRCFILE $PAMDIR
fi

BERKELYDBFILE="/etc/rdtagent/pam/rmd_users.db"

if [ -f $BERKELYDBFILE ]; then
    echo "Do you want to create/update users in RMD Berkely DB file?(y/n)"
    read -r a
    if [ $a == "y" -o $a == "Y" ]; then
        ./setup_rmd_users.sh
    elif [ $a != "n" -a $a != "N" ]; then
        echo "Invalid input. No action taken."
    fi
else
    ./setup_rmd_users.sh
fi
