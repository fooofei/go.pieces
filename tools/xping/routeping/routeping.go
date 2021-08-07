package main

import (
	"context"
	"fmt"
	"net"
	"time"

	netaddr "github.com/fooofei/pkg/net/addr"
	"github.com/fooofei/pkg/routetable"
	"github.com/fooofei/tools/ping/pkg/pinger"
	"github.com/kbinani/win"
)

type routePing struct {
	Dialer     *net.Dialer
	RouteTable *routetable.RouteTable
}

func (t *routePing) Ping(waitCtx context.Context, raddr string) (time.Duration, error) {
	networkDestAddr, err := netaddr.StringToNetwork(raddr)
	if err != nil {
		return 0, err
	}
	// IPv4 is enough.
	mibRouteRow := &win.MIB_IPFORWARDROW{}
	// https://docs.microsoft.com/en-us/windows/win32/api/iphlpapi/nf-iphlpapi-getbestroute
	r := win.GetBestRoute(win.DWORD(networkDestAddr), 0, mibRouteRow)
	if r != 0 {
		return 0, fmt.Errorf("failed call win.GetBestRoute return value %v", r)
	}
	mibRouteRows, err := t.RouteTable.Routes()
	if err != nil {
		return 0, err
	}
	routeRow := &routetable.Route{}
	routeRow.FromIpForwardRow(mibRouteRow)
	routeRows := routetable.RowsToRoutes(mibRouteRows)
	for _, row := range routeRows {
		if row.Equal(routeRow) {
			return 0, fmt.Errorf("%s", row.Digest())
		}
	}
	// return message by error
	return 0, fmt.Errorf("not found route detail %s", routeRow.Digest())
}

func (t *routePing) Ready(ctx context.Context, raddr string) error {
	var err error
	t.Dialer = &net.Dialer{}
	t.RouteTable, err = routetable.NewRouteTable()
	return err
}

func (t *routePing) Name() string {
	return "RoutePing"
}

func (t *routePing) Close() error {
	_ = t.RouteTable.Close()
	return nil
}

func main() {
	p := new(routePing)
	pinger.DoPing(p)
}
