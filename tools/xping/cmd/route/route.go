package main

import (
	"context"
	"fmt"
	netaddr "github.com/fooofei/go_pieces/pkg/net/addr"
	"github.com/fooofei/go_pieces/pkg/routetable"
	"github.com/fooofei/go_pieces/tools/ping/pkg/prober"
	"github.com/kbinani/win"
	"net"
)

// only support windows

type routeProbe struct {
	Dialer     *net.Dialer
	RouteTable *routetable.RouteTable
}

func getBestRoute(raddr string) (*routetable.Route, error) {
	var networkOrderAddr, err = netaddr.StringToNetwork(raddr)
	if err != nil {
		return nil, err
	}
	// IPv4 is enough.
	var bestRouteRow = &win.MIB_IPFORWARDROW{}
	// https://docs.microsoft.com/en-us/windows/win32/api/iphlpapi/nf-iphlpapi-getbestroute
	var r = win.GetBestRoute(win.DWORD(networkOrderAddr), 0, bestRouteRow)
	if r != 0 {
		return nil, fmt.Errorf("failed call win.GetBestRoute return value %v", r)
	}
	var bestRoute = &routetable.Route{}
	bestRoute.FromIpForwardRow(bestRouteRow)
	return bestRoute, nil
}

func getRouteTableList(routeTable *routetable.RouteTable) ([]*routetable.Route, error) {
	if routeRowMibList, err := routeTable.Routes(); err != nil {
		return nil, err
	} else {
		return routetable.RowsToRoutes(routeRowMibList), nil
	}
}

func (t *routeProbe) Probe(waitCtx context.Context, raddr string) (string, error) {
	var bestRoute, err = getBestRoute(raddr)
	if err != nil {
		return "", err
	}
	var routeTableList []*routetable.Route
	if routeTableList, err = getRouteTableList(t.RouteTable); err != nil {
		return "", err
	}
	for _, route := range routeTableList {
		if route.Equal(bestRoute) {
			return route.Digest(), nil
		}
	}
	return "", fmt.Errorf("cannot match route '%s' in route table", bestRoute.Digest())
}

func (t *routeProbe) Ready(ctx context.Context, raddr string) error {
	var err error
	t.Dialer = &net.Dialer{}
	t.RouteTable, err = routetable.NewRouteTable()
	return err
}

func (t *routeProbe) Name() string {
	return "RoutePing"
}

func (t *routeProbe) Example() string {
	return "routeping 1.1.1.1"
}

func (t *routeProbe) Close() error {
	return t.RouteTable.Close()
}

func main() {
	var p = new(routeProbe)
	prober.Do(p)
}
