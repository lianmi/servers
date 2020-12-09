# 官方sdk
https://github.com/aliyun/aliyun-oss-go-sdk

# STS临时授权

see: 
    STS临时授权 RoleArn在哪找 https://developer.aliyun.com/ask/2549

    STS Python_SDK授权临时用户读写OSS资源 https://developer.aliyun.com/article/756616

    官方 https://help.aliyun.com/document_detail/100624.html

## 只读 list 不能删除

其中: lianmi-ipfs 是bucket名称 
```
{
    "Version": "1",
    "Statement": [
     {
           "Effect": "Allow",
           "Action": [
             "oss:PutObject",
             "oss:GetObject"
             "oss:ListObjects"
           ],
           "Resource": [
             "acs:oss:*:*:lianmi-ipfs",
             "acs:oss:*:*:lianmi-ipfs/*"
           ]
     }
    ]
}
```


## 读写 list 不能删除

```
{
"Version": "1",
"Statement": [
 {
   "Effect": "Allow",
   "Action": [
     "oss:ListParts",
     "oss:AbortMultipartUpload",
     "oss:PutObject"
   ],
   "Resource": [
     "acs:oss:*:*:ram-test-app",
     "acs:oss:*:*:ram-test-app/*"
   ]
 }
]
}
```

## 读写 list 删除 目录限制在用户目录 

```
{
"Version": "1",
"Statement": [
 {
   "Effect": "Allow",
   "Action": [
     "oss:DeleteObject",
     "oss:ListParts",
     "oss:AbortMultipartUpload",
     "oss:PutObject"
   ],
   "Resource": [
     "acs:oss:*:*:ram-test-app",
     "acs:oss:*:*:ram-test-app/user001/*"
   ]
 }
]
}
```