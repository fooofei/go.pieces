

## 参数对应关系 

```
ssh -o "ProxyCommand client 127.0.0.1 8888 %h %p" root@12.12.12.12
```

调用到 client 程序中时，参数对应为

```
// 本地代理的地址
os.Args[1] = 127.0.0.1
os.Args[2] = 8888   

// ssh 目标主机
os.Args[3] = 12.12.12.12
os.Args[4] = 22
```

ssh TCP payload 数据是通过 stdin PIPE 传递到 client 程序的.
