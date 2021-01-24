// 操作 Windows 的 路由表

package routetable

import (
	"syscall"
	"unsafe"

	"github.com/kbinani/win"
)

type SliceHeader struct {
	Addr uintptr
	Len  int
	Cap  int
}

type DynamicMemory struct {
	mem []byte // 保存引用,防止被回收
}

func NewDynamicMemory(bytes uint32) *DynamicMemory {
	return &DynamicMemory{
		mem: make([]byte, bytes, bytes),
	}
}

func (dm *DynamicMemory) Len() uint32 {
	return uint32(len(dm.mem))
}

func (dm *DynamicMemory) Address() uintptr {
	return (*SliceHeader)(unsafe.Pointer(&dm.mem)).Addr
}

type RouteTable struct {
	dllHandle                *syscall.DLL
	getIpForwardTableProc    *syscall.Proc
	createIpForwardEntryProc *syscall.Proc
	deleteIpForwardEntryProc *syscall.Proc
	setIpForwardEntryProc    *syscall.Proc
}

func NewRouteTable() (*RouteTable, error) {
	dll, err := syscall.LoadDLL("iphlpapi.dll")
	if err != nil {
		return nil, err
	}

	getIpForwardTable, err := dll.FindProc("GetIpForwardTable")
	if err != nil {
		return nil, err
	}
	createIpForwardEntry, err := dll.FindProc("CreateIpForwardEntry")
	if err != nil {
		return nil, err
	}
	deleteIpForwardEntry, err := dll.FindProc("DeleteIpForwardEntry")
	if err != nil {
		return nil, err
	}
	setIpForwardEntry, err := dll.FindProc("SetIpForwardEntry")
	if err != nil {
		return nil, err
	}

	return &RouteTable{
		dllHandle:                dll,
		getIpForwardTableProc:    getIpForwardTable,
		createIpForwardEntryProc: createIpForwardEntry,
		deleteIpForwardEntryProc: deleteIpForwardEntry,
		setIpForwardEntryProc:    setIpForwardEntry,
	}, nil
}

func (table *RouteTable) Close() error {
	return table.dllHandle.Release()
}

//https://msdn.microsoft.com/en-us/library/windows/desktop/aa366852(v=vs.85).aspx
//typedef struct _MIB_IPFORWARDTABLE {
//  DWORD            dwNumEntries;
//  MIB_IPFORWARDROW table[ANY_SIZE];
//} MIB_IPFORWARDTABLE, *PMIB_IPFORWARDTABLE;

func (table *RouteTable) Routes() ([]win.MIB_IPFORWARDROW, error) {
	// 加4,是为了越过DWORD
	mem := NewDynamicMemory(
		uint32(
			4 + unsafe.Sizeof(win.MIB_IPFORWARDROW{}),
		),
	)
	tableSize := uint32(0)
	// 获取路由表数量
	_, r2, err := table.getIpForwardTableProc.Call(
		mem.Address(),
		uintptr(unsafe.Pointer(&tableSize)),
		0,
	)
	// msdn https://msdn.microsoft.com/en-us/library/windows/desktop/aa365953(v=vs.85).aspx
	if r2 != 0 {
		return nil, err
	}

	// 获取全部路由表
	mem = NewDynamicMemory(tableSize)
	_, r2, err = table.getIpForwardTableProc.Call(
		mem.Address(),
		uintptr(unsafe.Pointer(&tableSize)),
		0,
	)
	if r2 != 0 {
		return nil, err
	}

	num := *(*uint32)(unsafe.Pointer(mem.Address()))

	rows := []win.MIB_IPFORWARDROW{}
	shRows := (*SliceHeader)(unsafe.Pointer(&rows))
	shRows.Addr = mem.Address() + 4
	shRows.Len = int(num)
	shRows.Cap = int(num)
	return rows, nil
}

// 添加路由,需要管理员权限,才能添加成功
// IPHLPAPI_DLL_LINKAGE DWORD CreateIpForwardEntry(
//  PMIB_IPFORWARDROW pRoute
//);
func (table *RouteTable) Add(row win.PMIB_IPFORWARDROW) error {
	// https://msdn.microsoft.com/en-us/library/windows/desktop/aa365860(v=vs.85).aspx
	// The function returns NO_ERROR (zero) if the function is successful.
	r1, r2, err := table.createIpForwardEntryProc.Call(uintptr(unsafe.Pointer(row)))
	if r2 != 0 {
		return err
	}
	if r1 != 0 {
		return syscall.Errno(r1)
	}
	return nil
}

// IPHLPAPI_DLL_LINKAGE DWORD SetIpForwardEntry(
//  PMIB_IPFORWARDROW pRoute
//);
func (table *RouteTable) Set(row win.PMIB_IPFORWARDROW) error {
	r1, r2, err := table.setIpForwardEntryProc.Call(uintptr(unsafe.Pointer(row)))
	if r2 != 0 {
		return err
	}
	if r1 != 0 {
		return syscall.Errno(r1)
	}
	return nil
}

func (table *RouteTable) Remove(row win.PMIB_IPFORWARDROW) error {
	// The function returns NO_ERROR (zero) if the function is successful.
	r1, r2, err := table.deleteIpForwardEntryProc.Call(uintptr(unsafe.Pointer(row)))
	if r2 != 0 {
		return err
	}
	if r1 != 0 {
		return syscall.Errno(r1)
	}
	return nil
}
