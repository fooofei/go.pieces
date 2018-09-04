package main
/*
对比 https://blog.golang.org/pipelines
它写的那几个都不满意

for v:= range c{
    select{
        case <-done:
            return
        case c<-result:
    }
}

这样的写法 不还是阻塞在c上吗？如何c没有受信 那么 case <-done就不会得到执行
不符合预期

 */
import (
    "crypto/md5"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "sort"
    "sync"
    "time"
)

type result struct {
    path string
    digest [md5.Size]byte
    err error
}

// only for current directory files
func listFiles(done <- chan struct {}, dir string, wg * sync.WaitGroup) (<- chan * result){
    var results = make(chan * result)
    wg.Add(1)
    go func() {
        defer close(results)
        defer wg.Done()
        files,err := ioutil.ReadDir(dir)

        if err!=nil{
            results<- &result{err:err}
            return
        }
        for _,f := range files{

            // in case of timeout
            select {
                case <-done:
                    return
                case  results<-&result{path:filepath.Join(dir,f.Name())}:
            }
            //time.Sleep(time.Second * time.Duration(3))
        }

    }()

    return results
}

func walkFiles(done <- chan struct{}, dir string) (<- chan * result) {
    var results = make(chan * result)
    go func() {
        defer close(results)
        err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {

            if err != nil{
                results <- &result{err:err}
                return nil
            }
            //if info.Mode().IsDir() && path != dir{
            //    return filepath.SkipDir
            //}
            if !info.Mode().IsRegular(){
                return nil
            }

            select{
            case results<- &result{path:path}:
            case <-done:
                // 1. must return an error
                // 2 return SkipDir, the func will return nil
                return filepath.SkipDir
            }
            return nil
        })
        if err!=nil{
            panic(err)
        }
    }()
    return results
}

func digest(done <- chan struct{},inResult <-chan * result, wg * sync.WaitGroup)(chan * result){

    var outResult = make(chan * result)
    wg.Add(1)
    go func() {
        defer close(outResult)
        defer wg.Done()
        for{
            var v * result
            select {
            case <-done:
                return
            case v1,ok:=<-inResult:
                if !ok{
                    return
                }
                v=v1
            }
            if v==nil{
                return
            }
            if v.err == nil {
                b, err := ioutil.ReadFile(v.path)
                if err != nil {
                    v.err = err
                    break
                }
                if v.err == nil {
                    v.digest = md5.Sum(b)
                    //time.Sleep(time.Second * time.Duration(2))
                }
            }

            select {
                case <-done:
                    return
                case outResult<-v:
            }
        }
    }()

    return outResult
}


func MD5All(dir string) (map[string][md5.Size]byte, error){

    var done = make(chan struct{})
    var wg sync.WaitGroup
    var timeout = time.After(time.Second * time.Duration(2))

    var r = make(map[string][md5.Size]byte)

    //var paths = walkFiles(done,dir)
    var paths = listFiles(done,dir,&wg)
    var ms = digest(done,paths,&wg)

    for{
        breaked:=false
        select {
        case <-timeout:
            fmt.Println("timeout")
            breaked=true
        case v,ok:=<-ms:
            if !ok{
                breaked=true
                break
            }
            if v.err!=nil{
                return nil,v.err
            }
            r[v.path]=v.digest
        }
        if breaked{
            break
        }
    }

    // not care about timeout
    //for v := range ms{
    //    if v.err != nil{
    //        return nil,v.err
    //    }
    //    r[v.path]=v.digest
    //}
    close(done) // 错误地使用了 defer close(done) 结果在 wg.Wait()发生死锁
    wg.Wait()
    return r, nil
}


func TestHash(){

    var dir = "/Users/hujianfei/go/src/github.com/fooofei/go_pieces/.git"

    var result,err = MD5All(dir)
    if err != nil{
        panic(err)
    }

    var paths []string

    for path,_ := range result{
        paths=append(paths,path)
    }
    sort.Strings(paths)

    for _,path := range paths{
        shortName := filepath.Base(path)
        fmt.Printf("%x %v\n",result[path],shortName)
    }
}