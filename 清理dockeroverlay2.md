# 经过一段时间后，服务端的 docker目录的 overlay2占用空间过大

```
cd /store/service/docker/overlay2

$ docker system df

TYPE                TOTAL               ACTIVE              SIZE                RECLAIMABLE
Images              20                  19                  4.297GB             309.8MB (7%)
Containers          21                  21                  80.53GB             0B (0%)
Local Volumes       12                  1                   101.7MB             76.52MB (75%)
Build Cache                                                 0B                  0B

```

