package go_pieces

import (
    "bufio"
    "fmt"
    "io"
    "math"
    "math/rand"
    "os"
    "path"
    "path/filepath"
    "runtime"
    "sort"
    "strings"
    "time"
)

// 定义一个 struct
type MyStruct struct {
    x int
    y int
}

// 给这个struct 增加一个成员函数
func (self *MyStruct) String() string {
    r := fmt.Sprintf("<MyStruct>(%v,%v)", self.x, self.y)
    return r
}

func ExampleMyStruct() {
    a := MyStruct{1, 1}
    var b MyStruct = MyStruct{1, 1}
    c := new(MyStruct)
    c.x = 1
    c.y = 1
    // a'type is not *MyStruct, but can call a.String()
    fmt.Printf("MyStruct a=%T a=%v\n", a, a.String())
    fmt.Printf("MyStruct b=%T b=%v\n", b, b.String())
    fmt.Printf("MyStruct c=%T c=%v\n", c, c.String())
    //output:MyStruct a=go_pieces.MyStruct a=<MyStruct>(1,1)
    //MyStruct b=go_pieces.MyStruct b=<MyStruct>(1,1)
    //MyStruct c=*go_pieces.MyStruct c=<MyStruct>(1,1)
}

func ExampleCmdArgs() {
    argsLen := len(os.Args)

    fmt.Printf("argsLen:%v\n", argsLen)

    if argsLen > 0 {
        arg0 := os.Args[0]
        exePath, _ := filepath.Abs(arg0)
        exeName_path := path.Base(exePath)
        exeName_filepath := filepath.Base(exePath)
        ext := filepath.Ext(exeName_filepath)
        //fmt.Printf("arg0=%v\n", arg0)
        //fmt.Printf("exePath=%v\n", exePath)
        // path.Base not work as filepath.Base
        exeName_path += ""
        // fmt.Printf("exeName_path=%v\n", exeName_path)
        fmt.Printf("exeName_filepath=%v\n", exeName_filepath)
        fmt.Printf("exeName_filepath_base=%v\n", filepath.Base(exeName_filepath))
        fmt.Printf("ext=%v\n", ext)
    }
    // windows output:
    //argsLen:2
    //exeName_filepath=temp.test.exe
    //exeName_filepath_base=temp.test.exe
    //ext=.exe
}

func ExampleSomeConstants() {
    rand.Seed(time.Now().Unix())
    n := rand.Intn(10)
    b := (n >= 0 && n < 10)
    fmt.Printf("randNum=[0,10) %v\n", b)
    fmt.Printf("Phi=%.3f\n", math.Phi)
    fmt.Printf("Pi=%.3f\n", math.Pi)
    fmt.Printf("GOOS=%v\n", runtime.GOOS)
    fmt.Printf("GOARCH=%v\n", runtime.GOARCH)
    // windows output:
    //randNum=[0,10) true
    //Phi=1.618
    //Pi=3.142
    //GOOS=windows
    //GOARCH=amd64
    // macOS output:
    //randNum=[0,10) true
    //Phi=1.618
    //Pi=3.142
    //GOOS=darwin
    //GOARCH=amd64
}

type FibMaker func() int

func (f FibMaker) Read(p []byte) (int, error) {
    next := f()
    if next > 1000 {
        return 0, io.EOF
    }
    s := fmt.Sprintf("%v\n", next)
    return strings.NewReader(s).Read(p)
}

func LinedPrinter(r io.Reader) {
    sn := bufio.NewScanner(r)

    for sn.Scan() {
        fmt.Println(sn.Text())
    }
}

func fib() func() int {
    a, b := 0, 1
    return func() int {
        a, b = b, a+b
        return a
    }
}

func ExampleFuncInterface() {
    var f FibMaker = fib()
    LinedPrinter(f)
    //output:
    //1
    //1
    //2
    //3
    //5
    //8
    //13
    //21
    //34
    //55
    //89
    //144
    //233
    //377
    //610
    //987
}



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