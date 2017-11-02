#!/usr/bin/env bash

if [[ $EUID != 0 ]]; then
  echo "This script can be run only by root"
  exit 1
fi

BERKELYDBDIR="/etc/rmd/pam"
BERKELYDBFILENAME="rmd_users.db"
echo 'Setup Berkely db users'
while true
do
    echo 'Enter username or 0 to stop'
    read u
    if [ $u == "0" ]; then
        break
    fi
    echo $u >> users
    echo 'Enter password:'
    read -s p
    openssl passwd -crypt $p >> users
done

# If input file was created
if [ -f "users" ]; then
    mkdir -p $BERKELYDBDIR
    # Berkely DB is access restricted to root only
    db_load -T -t hash -f users $BERKELYDBDIR"/"$BERKELYDBFILENAME
    rm -rf users
fi
