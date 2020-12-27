#!/bin/sh

echo "======== docker containers logs file size ========"  

logs=$(find /store/service/docker/overlay2/ -name *.log)  

for log in $logs  
        do  
             ls -lh $log   
        done  

