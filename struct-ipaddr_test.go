package main

import (
    "fmt"
    "sort"
)

type IpAddr [4]uint8


func (self *IpAddr) String() (string){
	r := fmt.Sprintf("%v.%v.%v.%v", self[0], self[1], self[2], self[3])
	return r
}

func (self IpAddr) String2()(string){
	r := fmt.Sprintf("%v.%v.%v.%v", self[0], self[1], self[2], self[3])
	return r
}

func ExampleIpaddr(){
	hosts := map[string]IpAddr{
		"loopback":  {127, 0, 0, 1},
		"googleDNS": {8, 8, 8, 8},
	}
	keys := make([]string, 0)
	for name,_ := range hosts {
	    keys = append(keys,name)
    }
	sort.Strings(keys)
	// loopback 和 googleDNS 出现的顺序不固定
	// 需要借助数组这个数据结构来稳定输出顺序
	for _,name := range keys{
	    ip := hosts[name]
        fmt.Printf("%v,%v,%v,%v\n", name, ip, ip.String(), ip.String2())
    }
	//output:
	//googleDNS,[8 8 8 8],8.8.8.8,8.8.8.8
    //loopback,[127 0 0 1],127.0.0.1,127.0.0.1
}