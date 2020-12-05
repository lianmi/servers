# UI层的oss用户

参考：
https://help.aliyun.com/document_detail/100624.html?spm=a2c4g.11186623.2.10.4ae06627UlCeej#concept-xzh-nzk-2gb


## 1. 创建子账号。

## 2. 创建权限策略。
```
{
    "Version": "1",
    "Statement": [
     {
           "Effect": "Allow",
           "Action": [
             "oss:ListObjects",
             "oss:PutObject",
             "oss:GetObject"
           ],
           "Resource": [
             "acs:oss:*:*:ram-test",
             "acs:oss:*:*:ram-test/*"
           ]
     }
    ]
}
```