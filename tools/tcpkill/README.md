

https://github.com/Kkevsterrr/tcpkiller/blob/master/tcpkiller
利用 python 的 scrapy 抓包，
s = socket.socket(socket.PF_PACKET, socket.SOCK_RAW)
s.bind((iface, 0))
发送     tcp = TCP(sport=src_port, dport=dst_port, seq=seq, flags="R")
都是要设置回调函数的


https://github.com/ggreer/dsniff/blob/master/tcpkill.c
与上面的原理相同 通过pcap 库监听，然后发送 RST 



https://github.com/kristrev/tcp_closer
依赖支持 SOCK_DESTROY 
ss 命令实现了这个，但是我从来没使用成功过呢 也没搜索到有人使用这个标记
