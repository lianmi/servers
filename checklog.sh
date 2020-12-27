#!/bin/sh

echo "======== docker containers logs file size ========"  

logs=$(find /store/service/docker/containers/ -name *-json.log)  

for log in $logs  
        do  
             ls -lh $log   
        done  

