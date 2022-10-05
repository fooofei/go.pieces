package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func splitResultList(itemList []*requestItem) ([]*requestItem, []*requestItem) {
	var validItemList = make([]*requestItem, 0)
	var invalidItemList = make([]*requestItem, 0)
	for _, item := range itemList {
		if item.Result != "" {
			validItemList = append(validItemList, item)
		} else {
			invalidItemList = append(invalidItemList, item)
		}
	}
	return validItemList, invalidItemList
}

// 关于获取自己的 IP 这个需求，有使用 DNS 的方式的，命令为
// dig +short TXT o-o.myaddr.l.google.com @114.114.114.114
// 但是获取到的IP 与 http 获取到的不同
// https://unix.stackexchange.com/questions/22615/how-can-i-get-my-external-ip-address-in-a-shell-script
// https://poplite.xyz/post/2018/05/19/how-to-get-your-public-ip-by-dns-lookup.html

// https://api.myip.com/ 美国
// https://myexternalip.com/json 美国
// https://ipapi.co/json 美国
// https://ident.me/.json 英国
// https://get.geojs.io/v1/ip.json 美国
func main() {
	var waitTimeout time.Duration
	//
	flag.DurationVar(&waitTimeout, "wait", 6*time.Second, "wait for timeout(5s, 1h)")
	flag.Parse()
	// no need to use https://api.ipify.org/?format=json
	pubSrvList := [...]*requestItem{
		{Uri: "https://ip.nf/me.json", ParseFunc: getIpInJsonIPIP},
		{Uri: "http://ip-api.com/json", ParseFunc: getIpInJsonQuery},
		{Uri: "https://wtfismyip.com/json", ParseFunc: getIpInJsonYourFuck},
		{Uri: "https://api.ipify.org", ParseFunc: getIpInPlainText},
		{Uri: "https://ip.seeip.org", ParseFunc: getIpInPlainText},
		{Uri: "https://ifconfig.me/ip", ParseFunc: getIpInPlainText},
		{Uri: "https://ifconfig.co/ip", ParseFunc: getIpInPlainText},
		// taobao 的服务不稳定
		{Uri: "http://ip.taobao.com/service/getIpInfo2.php?ip=myip",
			ParseFunc: getIpInJsonTaobao},
		{Uri: "http://members.3322.org/dyndns/getip", ParseFunc: getIpInPlainText},
		{Uri: "http://ip.cip.cc", ParseFunc: getIpInPlainText},
		{Uri: "https://api.ip.sb/ip", ParseFunc: getIpInPlainText},
		{Uri: "http://ifcfg.cn/echo", ParseFunc: getIpInJsonIP},
		{Uri: "http://eth0.me", ParseFunc: getIpInPlainText},
		{Uri: "http://ip.360.cn/IPShare/info", ParseFunc: getIpInJsonIP},
	}
	// log init
	log.SetFlags(log.LstdFlags)
	log.SetPrefix(fmt.Sprintf("pid= %v ", os.Getpid()))

	var waitCtx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if waitTimeout > 0 {
		waitCtx, cancel = context.WithTimeout(waitCtx, waitTimeout)
	}

	log.Printf("enter to send many requests")
	sendAll(waitCtx, pubSrvList[:])
	log.Printf("leave to send many requests")
	//
	var itemList, invalidItemList = splitResultList(pubSrvList[:])
	var sortItemList = sortResultList(itemList)
	//
	log.Printf("fetch result cnt= %v from give %v uris", len(itemList), len(pubSrvList))
	for _, sortItem := range sortItemList {
		fmt.Printf("%s \n", sortItem.Result)
		for i, item := range sortItem.ItemList {
			fmt.Printf("  [%v]%s %s\n", i+1, item.Uri, item.TakeTime.String())
		}
	}
	for _, item := range invalidItemList {
		fmt.Printf("failed get ip from %s takeTime %s error %v\n", item.Uri, item.TakeTime.String(), item.Err)
	}
	log.Printf("main exit")
}
