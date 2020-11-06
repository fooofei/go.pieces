package routetable

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"router/pkg/netaddr"
	"syscall"
	"unsafe"
)

type Route struct {
	Dest      string
	Mask      string
	NextHop   string
	Policy    uint32
	IfIndex   uint32
	IfName    string
	IfMacAddr string
	IfDesc    string
	IfUUID    string
	Type      uint32
	Proto     uint32
	Age       uint32
	NextHopAS uint32
	Metric1   uint32
	Metric2   uint32
	Metric3   uint32
	Metric4   uint32
	Metric5   uint32
}

func (r *Route) ToIpForwardRow() *IpForwardRow {
	var err error
	row := &IpForwardRow{
		ForwardPolicy:    r.Policy,
		ForwardIfIndex:   r.IfIndex,
		ForwardType:      r.Type,
		ForwardProto:     r.Proto,
		ForwardAge:       r.Age,
		ForwardNextHopAS: r.NextHopAS,
		ForwardMetric1:   r.Metric1,
		ForwardMetric2:   r.Metric2,
		ForwardMetric3:   r.Metric3,
		ForwardMetric4:   r.Metric4,
		ForwardMetric5:   r.Metric5,
	}
	_ = err
	row.ForwardDest, err = netaddr.StringToNetwork(r.Dest)
	row.ForwardMask, err = netaddr.StringToNetwork(r.Mask)
	row.ForwardNextHop, err = netaddr.StringToNetwork(r.NextHop)
	return row
}

func (r *Route) FromIpForwardRow(row *IpForwardRow) {
	// 11 fields
	var err error
	r.Policy = row.ForwardPolicy
	r.IfIndex = row.ForwardIfIndex
	r.Type = row.ForwardType
	r.Proto = row.ForwardProto
	r.Age = row.ForwardAge
	r.NextHopAS = row.ForwardNextHopAS
	r.Metric1 = row.ForwardMetric1
	r.Metric2 = row.ForwardMetric2
	r.Metric3 = row.ForwardMetric3
	r.Metric4 = row.ForwardMetric4
	r.Metric5 = row.ForwardMetric5
	_ = err
	r.Dest, err = netaddr.NetworkToString(row.ForwardDest)
	r.Mask, err = netaddr.NetworkToString(row.ForwardMask)
	r.NextHop, err = netaddr.NetworkToString(row.ForwardNextHop)
}

func (r *Route) String() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func (r *Route) Digest() string {
	return fmt.Sprintf("%v/%v %v %v/%v/%v",
	r.Dest, r.Mask, r.NextHop, r.IfName, r.IfDesc, r.IfMacAddr)
}

// GetAdapterList return the Windows interfaces's adapter info
// we want the Desc field.
// copy from https://play.golang.org/p/kJ0P7HnvDE
func GetAdapterList() (*syscall.IpAdapterInfo, error) {
	b := make([]byte, 1000)
	l := uint32(len(b))
	a := (*syscall.IpAdapterInfo)(unsafe.Pointer(&b[0]))
	// TODO(mikio): GetAdaptersInfo returns IP_ADAPTER_INFO that
	// contains IPv4 address list only. We should use another API
	// for fetching IPv6 stuff from the kernel.
	err := syscall.GetAdaptersInfo(a, &l)
	if err == syscall.ERROR_BUFFER_OVERFLOW {
		b = make([]byte, l)
		a = (*syscall.IpAdapterInfo)(unsafe.Pointer(&b[0]))
		err = syscall.GetAdaptersInfo(a, &l)
	}
	if err != nil {
		return nil, os.NewSyscallError("GetAdaptersInfo", err)
	}
	return a, nil
}

func trimName(s []byte) string {
	return string(bytes.Trim(s, " \r\n\x00"))
}

// RowsToRoutes will add some custom field,
// convert []IpForwardRow to []Route
func RowsToRoutes(rows []IpForwardRow) []*Route {
	routes := make([]*Route, 0)
	interfacesMap := make(map[int]net.Interface, 0)
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, interfaceNet := range interfaces {
			interfacesMap[interfaceNet.Index] = interfaceNet
		}
	}
	adaptersMap := make(map[int]*syscall.IpAdapterInfo, 0)
	adapters, err := GetAdapterList()
	if err == nil {
		var next *syscall.IpAdapterInfo
		next = adapters
		for next != nil {
			adaptersMap[int(next.Index)] = next
			next = next.Next
		}
	}

	for i, _ := range rows {
		row := &rows[i]
		route := &Route{}
		route.FromIpForwardRow(row)
		idx := int(route.IfIndex)
		interfaceNet, exists := interfacesMap[idx]
		if exists {
			route.IfName = interfaceNet.Name
			route.IfMacAddr = interfaceNet.HardwareAddr.String()
		}
		adapter, exists := adaptersMap[idx]
		if exists {
			route.IfDesc = trimName(adapter.Description[:])
			route.IfUUID = trimName(adapter.AdapterName[:])
		}
		routes = append(routes, route)
	}
	return routes
}
