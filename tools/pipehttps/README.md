


# pipeHTTPS

pipe https will receive client's HTTP/HTTPS request, and send to the real server.

we can dump the request and response from the middle.








    ┌──────────┐          ┌───────────┐      ┌──────────────┐
    │          │ ───►HTTP │           │      │              │
    │ Client   │          │   tools  ─┼───►HTTPS  Server    │
    │          │ ───►HTTPS│           │      │              │
    └──────────┘          └───────────┘      └──────────────┘
                        Dump HTTP Req&Resp








                                                               ▼
