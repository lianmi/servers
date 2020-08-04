# Ubuntu下 Mosquitto的配置
https://www.cnblogs.com/guyk/p/12405938.html

## 配置文件
```
# 消息持久存储
persistence true
persistence_location /home/lishijia/developments/lianmi/servers/mqttserver/persistence

# 日志文件
log_dest file /home/lishijia/developments/lianmi/servers/mqttserver/log/mosquitto.log

# 其他配置
include_dir /home/lishijia/developments/lianmi/servers/mqttserver/conf.d

# 匿名访问 false为禁止
allow_anonymous true

# 认证配置
password_file /home/lishijia/developments/lianmi/servers/mqttserver/pwfile

# 权限配置
acl_file /home/lishijia/developments/lianmi/servers/mqttserver/aclfile
```

## 运行
```
./mosquitto -c ./mosquitto.conf -d
```