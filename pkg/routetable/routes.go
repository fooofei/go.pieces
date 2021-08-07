package routetable

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"syscall"
	"unsafe"

	netaddr "github.com/fooofei/go_pieces/pkg/net/addr"
	"github.com/kbinani/win"
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
	Type      int32
	Proto     int32
	Age       uint32
	NextHopAS uint32
	Metric1   uint32
	Metric2   uint32
	Metric3   uint32
	Metric4   uint32
	Metric5   uint32
}

func (r *Route) ToIpForwardRow() *win.MIB_IPFORWARDROW {
	var err error
	row := &win.MIB_IPFORWARDROW{
		DwForwardPolicy:    win.DWORD(r.Policy),
		DwForwardIfIndex:   win.IF_INDEX(r.IfIndex),
		ForwardType:        win.MIB_IPFORWARD_TYPE(r.Type),
		ForwardProto:       win.MIB_IPFORWARD_PROTO(r.Proto),
		DwForwardAge:       win.DWORD(r.Age),
		DwForwardNextHopAS: win.DWORD(r.NextHopAS),
		DwForwardMetric1:   win.DWORD(r.Metric1),
		DwForwardMetric2:   win.DWORD(r.Metric2),
		DwForwardMetric3:   win.DWORD(r.Metric3),
		DwForwardMetric4:   win.DWORD(r.Metric4),
		DwForwardMetric5:   win.DWORD(r.Metric5),
	}
	_ = err
	var value uint32
	value, err = netaddr.StringToNetwork(r.Dest)
	row.DwForwardDest = win.DWORD(value)
	value, err = netaddr.StringToNetwork(r.Mask)
	row.DwForwardMask = win.DWORD(value)
	value, err = netaddr.StringToNetwork(r.NextHop)
	row.DwForwardNextHop = win.DWORD(value)
	return row
}

func (r *Route) FromIpForwardRow(row *win.MIB_IPFORWARDROW) {
	// 11 fields
	var err error
	r.Policy = uint32(row.DwForwardPolicy)
	r.IfIndex = uint32(row.DwForwardIfIndex)
	r.Type = int32(row.ForwardType)
	r.Proto = int32(row.ForwardProto)
	r.Age = uint32(row.DwForwardAge)
	r.NextHopAS = uint32(row.DwForwardNextHopAS)
	r.Metric1 = uint32(row.DwForwardMetric1)
	r.Metric2 = uint32(row.DwForwardMetric2)
	r.Metric3 = uint32(row.DwForwardMetric3)
	r.Metric4 = uint32(row.DwForwardMetric4)
	r.Metric5 = uint32(row.DwForwardMetric5)
	_ = err
	r.Dest, err = netaddr.NetworkToString(uint32(row.DwForwardDest))
	r.Mask, err = netaddr.NetworkToString(uint32(row.DwForwardMask))
	r.NextHop, err = netaddr.NetworkToString(uint32(row.DwForwardNextHop))
}

func (r *Route) String() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func (r *Route) Digest() string {
	return fmt.Sprintf("%v/%v %v %v/%v/%v/%v",
		r.Dest, r.Mask, r.NextHop, r.IfIndex, r.IfName, r.IfDesc, r.IfMacAddr)
}

func (r *Route) Equal(dst *Route) bool {
	return r.IfIndex == dst.IfIndex && r.Dest == dst.Dest && r.Mask == dst.Mask
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
func RowsToRoutes(rows []win.MIB_IPFORWARDROW) []*Route {
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
