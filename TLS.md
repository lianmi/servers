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