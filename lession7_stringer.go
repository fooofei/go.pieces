package main


import "fmt"

type IPAddr [4]byte

func (self  IPAddr)String() (r string) {
	r = fmt.Sprintf("%v.%v.%v.%v", self[0],self[1],self[2],self[3])
	return r
}

// Warning  pointer is not right

/*
func (self * IPAddr)String() (r string) {
	r = fmt.Sprintf("%v.%v.%v.%v", self[0],self[1],self[2],self[3])
	return r
}
*/

func testAddr1(){
	hosts := map[string]IPAddr{
		"loopback":  {127, 0, 0, 1},
		"googleDNS": {8, 8, 8, 8},
	}
	for name, ip := range hosts {
		fmt.Printf("%v: %v\n", name, ip)
	}
}

func main() {
	testAddr1()
}