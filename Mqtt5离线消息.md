Mqtt5 - 会话保留与离线消息接收（完整）

https://blog.csdn.net/luo15242208310/article/details/103971457

Mqtt5支持离线消息接收的几个核心设置：
ClientId
CleanStart: false
SessionExpiry
Qos:2
CONNACK中的session present flag

ClientId用于唯一标识用户session。
CleanStart设置为0，表示创建一个持久会话，在客户端断开连接时，会话仍然保持并保存离线消息，直到会话超时注销。CleanStart设置为1，表示创建一个新的临时会话，在客户端断开时，会话自动销毁。
SessionExpiry即指定在CleanStart为0时，会话的保存时长，如果客户端未在用户定义的时间段内连接，则可以丢弃状态（例如，订阅和缓冲的消息）而无需进行清理。
Qos即消息的Quality of Service，若要支持离线消息，需要订阅端、发布端Qos >= 1
session present即在connect到mqtt服务器的返回结果ConnAck中，包含session present标识，该标识表示当前clientId是否存在之前的持久会话（persistent session），若之前已存在session（此时千万不要再次重复订阅topic，若再次订阅则之前的消息都将收不到），则session会保留之前的订阅关系、客户端离线时的消息（Qos>=1）、未ack的消息。重点说明一下session present的使用，在客户端连接到mqtt服务器并获取到connack中的isSessionPresent标识时，若isSessionPresent=true则已存在会话，此时无需再重复订阅topic（订阅关系已保存到session中，若再重复订阅则收不到之前的离线消息），可通过全局接收来处理离线消息和之后的新消息；若isSessionPresent=false则不存在session（又或者session已超期），此时需要重新订阅topic，且之前离线的消息都已接收不到，只能通过其他方式获取离线消息（例如IM后端服务的全量同步服务）。

xxx
图片截取自：mqtt-essentials-part-3-client-broker-connection-establishment

在这里插入图片描述
图片截取自：mqtt-essentials-part-7-persistent-session-queuing-messages

如ClientId=1, CleanStart=false, SessionExpiry=3600s, Qos=2即指定clientId=1的会话为持久会话，用户在离线后3600s的的离线消息都会被Mqtt服务器保存，用户在离线时间不超过3600s且再次以ClientId=1重新上线时，是可以收到离线期间消息的补充推送的，同时Qos=2（exactly once）保证消息只会被客户端收到一次且一定一次。



## 以上的几个核心设置：
clientId,
cleanStart=fasle,
sessionExpiry > 0,
Qos>=1，
CONNACK session present处理，
缺一不可，少一项设置便无法实现离线消息的接受。