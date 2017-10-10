#!/usr/bin/env bash

if [[ $EUID != 0 ]]; then
  echo "This script can be run only by root"
  exit 1
fi

USER="rmd"
sudo useradd $USER || echo "User rmd already exists."

BERKELYDBFILE="/etc/rdtagent/pam/rmd_users.db"
echo 'Setup berkely db users'
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

db_load -T -t hash -f users $BERKELYDBFILE
sudo chown $USER:$USER $BERKELYDBFILE
rm -rf users