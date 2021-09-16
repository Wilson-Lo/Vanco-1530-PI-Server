#!/bin/sh

echo "update server"
unzip -o /home/pi/golang/controller/tmp/fw.zip -d /home/pi/app/ &

wait
echo "update end"
sync
echo "sync end"
