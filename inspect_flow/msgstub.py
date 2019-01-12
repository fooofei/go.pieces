#coding=utf-8

import os
import sys

from socket import AF_INET
from socket import SOCK_STREAM
from socket import ntohs
from socket import htons
from socket import socket
from socket import SOL_SOCKET
from socket import SO_REUSEADDR
from time import sleep
from struct import pack as st_pack


def main():
    laddr = ('0.0.0.0',8889)

    fd = socket(AF_INET, SOCK_STREAM, 0)
    fd.setsockopt(SOL_SOCKET, SO_REUSEADDR, 1)
    fd.bind(laddr)
    fd.listen(2)
    print('bind to {}'.format(laddr))
    brk = False
    try:
        while not brk:
            f,raddr = fd.accept()
            print('accept from {}'.format(raddr))

            while not brk:
                try:
                    msg = 'helloworldssssssss'
                    ctnt = htons(len(msg))
                    ctnt = st_pack('H',ctnt)
                    ctnt += msg
                    print('tx len={}'.format(len(ctnt)))
                    f.send(ctnt)
                    sleep(1)
                except KeyboardInterrupt:
                    brk=True
                except Exception as er:
                    print(er)
                finally:
                    f.close()
                    break
    except KeyboardInterrupt:
        brk=True
    finally:
        print('main exit')
        fd.close()

if __name__ == '__main__':
    main()