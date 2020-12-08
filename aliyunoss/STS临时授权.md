# STS临时授权

see: https://developer.aliyun.com/ask/2549

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