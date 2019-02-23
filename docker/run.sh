#!/bin/sh
if [ ! -f /daemon/config/go-daemon.yaml ]
then
    echo "No yaml found, copy it!"
    cp /daemon/go-daemon.yaml /daemon/config/go-daemon.yaml
fi
./go-daemon
