#!/bin/sh 
  
echo "======== start clean docker containers logs ========"  
  
logs=$(find /store/service/docker/containers/ -name *-json.log)  
  
for log in $logs  
        do  
                echo "clean logs : $log"  
                cat /dev/null > $log  
        done  

echo "======== end clean docker containers logs ========"  
