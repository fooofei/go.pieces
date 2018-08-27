package main

import "fmt"

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
	for name, ip := range hosts {
		fmt.Printf("%v,%v,%v,%v\n", name, ip, ip.String(), ip.String2())
	}
	//output:
	//loopback,[127 0 0 1],127.0.0.1,127.0.0.1
	//googleDNS,[8 8 8 8],8.8.8.8,8.8.8.8
}