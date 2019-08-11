# xping

xping contains tcping udping httping.


## TCPing

A choice of tcping maybe https://github.com/KirillShlenskiy/tcping

The bad is the time of start request pkts not Precision.

I dump some pkts when tcping:
```
 No.     Time                 Source        sport  Destination    dport  Protocol Length Info
 4  2019-03-24 19:07:32.895642 192.168.0.108 61917  183.131.192.23  443 TCP  78  61917 → 443 [SYN] 
10  2019-03-24 19:07:33.905305 192.168.0.108 61918  183.131.192.23  443 TCP  78  61918 → 443 [SYN] 
25  2019-03-24 19:07:34.915931 192.168.0.108 61919  183.131.192.23  443 TCP  78  61919 → 443 [SYN] 
41  2019-03-24 19:07:35.923406 192.168.0.108 61920  183.131.192.23  443 TCP  78  61920 → 443 [SYN] 
54  2019-03-24 19:07:36.931049 192.168.0.108 61921  183.131.192.23  443 TCP  78  61921 → 443 [SYN] 
65  2019-03-24 19:07:37.944162 192.168.0.108 61922  183.131.192.23  443 TCP  78  61922 → 443 [SYN] 
72  2019-03-24 19:07:38.953240 192.168.0.108 61923  183.131.192.23  443 TCP  78  61923 → 443 [SYN] 
78  2019-03-24 19:07:39.967107 192.168.0.108 61924  183.131.192.23  443 TCP  78  61924 → 443 [SYN] 
86  2019-03-24 19:07:40.979846 192.168.0.108 61925  183.131.192.23  443 TCP  78  61925 → 443 [SYN] 
92  2019-03-24 19:07:41.990226 192.168.0.108 61926  183.131.192.23  443 TCP  78  61926 → 443 [SYN] 
100 2019-03-24 19:07:43.000308 192.168.0.108 61927  183.131.192.23  443 TCP  78  61927 → 443 [SYN] 
107 2019-03-24 19:07:44.011435 192.168.0.108 61928  183.131.192.23  443 TCP  78  61928 → 443 [SYN] 
```
The pkt between 92 and 100, 19:07:42 not sent, every pkt sent delay a little.

so I make some change on my tcping.

The delay time not use `time.Second` instead of `time.Nanosecond * 999 * 998 * 999`.

or use `time.Tick`.


```shell
$ ./tcping 127.0.0.1:63347
=> TCPing 127.0.0.1:63347 for infinit=false n=4
> [0001][09:11:01] 127.0.0.1:63347: 2.38 ms
> [0002][09:11:02] 127.0.0.1:63347: 0.45 ms
> [0003][09:11:03] 127.0.0.1:63347: 0.42 ms
> [0004][09:11:04] 127.0.0.1:63347: 0.45 ms
 Sent = 4,  Received = 4 (100.0%)
 Minimum = 0.42 ms, Maximum = 2.38 ms
 Average = 0.92 ms, Median = 0.45 ms
 90% of Request <= 1.42 ms
 75% of Request <= 0.45 ms
 50% of Request <= 0.45 ms
```


## HTTPing

```shell
$ ./httping http://www.baidu.com
=> HTTPing http://www.baidu.com for infinit=false n=4
> [0001][09:07:31] http://www.baidu.com: 114.35 ms
> [0002][09:07:32] http://www.baidu.com: 87.19 ms
> [0003][09:07:33] http://www.baidu.com: 63.96 ms
> [0004][09:07:34] http://www.baidu.com: 73.34 ms
 Sent = 4,  Received = 4 (100.0%)
 Minimum = 63.96 ms, Maximum = 114.35 ms
 Average = 84.71 ms, Median = 80.27 ms
 90% of Request <= 100.77 ms
 75% of Request <= 87.19 ms
 50% of Request <= 73.34 ms
```
## UDPing