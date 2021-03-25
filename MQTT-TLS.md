# 运行 
docker run -it \
	-p 1883:1883 \
	-p 9001:9001 \
	-v /Users/mac/developments/lianmi/ssl/mqtt/lianmica/mosquitto.conf:/mosquitto/config/mosquitto.conf \
	-v /Users/mac/developments/lianmi/ssl/mqtt/lianmica/data:/mosquitto/data \
	-v /Users/mac/developments/lianmi/ssl/mqtt/lianmica/log:/mosquitto/log \
	-v /Users/mac/developments/lianmi/ssl/mqtt/lianmica/ca:/mosquitto/ca \
	eclipse-mosquitto

# 发布 

./mosquitto_pub -t test -m dsada -h 192.168.1.125\
                 -p 1883   \
                 --cafile /Users/mac/developments/lianmi/ssl/mqtt/lianmica/ca/ca.crt \
                 --cert /Users/mac/developments/lianmi/ssl/mqtt/lianmica/ca/mqtt.lianmi.cloud.crt  \
                 --key /Users/mac/developments/lianmi/ssl/mqtt/lianmica/ca/mqtt.lianmi.cloud.key


# 订阅

./mosquitto_sub -t test  -v \
--cafile /Users/mac/developments/lianmi/ssl/mqtt/lianmica/ca/ca.crt \
--cert /Users/mac/developments/lianmi/ssl/mqtt/lianmica/ca/mqtt.lianmi.cloud.crt  \
--key /Volumes/wd320/wujehy/devroot/Projects/mqservice/ca/export/mqtt.lianmi.cloud.key \
-h 192.168.1.125 -p 1883




#  如何理解并配置日志输出：

http://www.steves-internet-guide.com/mosquitto-logging/

命令行：

./mosquitto_pub  -m test -t lianmi/lwt -u 6c4d14f5-1950-4e8a-8f61-19413839576d -P eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkZXZpY2VJRCI6IjZjNGQxNGY1LTE5NTAtNGU4YS04ZjYxLTE5NDEzODM5NTc2ZCIsImV4cCI6MTYxOTE2NTQ2Niwib3JpZ19pYXQiOjE2MTY1NzM0NjYsInVzZXJOYW1lIjoiaWQ0IiwidXNlclJvbGVzIjoiW3tcImlkXCI6NCxcInVzZXJfaWRcIjo0LFwidXNlcl9uYW1lXCI6XCJpZDRcIixcInZhbHVlXCI6XCJcIn1dIn0.SqFH87mtVrv9Mg4LViaGYOlaqfhLiBPe9xF5BR011A8  -V mqttv5  --cafile /home/wujehy/devroot/cpp/mosquitto_auth_plugin/output_server/docker_mqtt_plugin/ca/ca.crt --cert /home/wujehy/devroot/cpp/mosquitto_auth_plugin/output_server/docker_mqtt_plugin/ca/mqtt.lianmi.cloud.crt  --key /home/wujehy/devroot/cpp/mosquitto_auth_plugin/output_server/docker_mqtt_plugin/ca/mqtt.lianmi.cloud.key --will-topic lianmi/lwt --will-payload disconnected:6c4d14f5-1950-4e8a-8f61-19413839576d -h mqtt.lianmi.cloud -p 1883


##  订阅mosquitto系统topic


## 1. 连接总数
```
mosquitto_sub  -t '$SYS/broker/clients/connected' -v
```

## 2. 断开的连接数
```
mosquitto_sub  -t '$SYS/broker/clients/disconnected' -v
```

## 2. 所有连接数（活动的和非活动的）
```
mosquitto_sub  -t '$SYS/broker/clients/total' -v
```