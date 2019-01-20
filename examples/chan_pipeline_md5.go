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
    "context"
    "crypto/md5"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "sort"
    "sync"
    "time"
)

type Entry struct {
    FilePath string
    Digest   [md5.Size]byte
    err      error
}

func listFiles(wg *sync.WaitGroup, waitCtx context.Context, dir string) (chan *Entry) {
    results := make(chan *Entry)

    wg.Add(1)
    go func() {
        defer close(results)
        defer wg.Done()
        files, err := ioutil.ReadDir(dir)
        if err != nil {
            select {
            case results <- &Entry{FilePath: dir, err: err}:
            case <-waitCtx.Done():
            }
            return
        }
    loop:
        for _, f := range files {
            select {
            case <-waitCtx.Done():
                break loop
            case results <- &Entry{FilePath: filepath.Join(dir, f.Name())}:
            }
        }

    }()

    return results
}

func walkFiles(wg *sync.WaitGroup, waitCtx context.Context, dir string) (chan *Entry) {
    results := make(chan *Entry)
    wg.Add(1)
    go func() {
        defer wg.Done()
        defer close(results)
        err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {

            // no block tell we are done, as early exit
            select {
            case <-waitCtx.Done():
                return filepath.SkipDir
            default:
            }

            if err != nil {
                select {
                case <-waitCtx.Done():
                case results <- &Entry{err: err}:
                }
                return nil
            }
            //if info.Mode().IsDir() && path != dir{
            //    return filepath.SkipDir
            //}
            if !info.Mode().IsRegular() {
                return nil
            }

            select {
            case results <- &Entry{FilePath: path}:
            case <-waitCtx.Done():
                return filepath.SkipDir
            }
            return nil
        })
        if err != nil {
            panic(err)
        }
    }()
    return results
}

func digest(wg *sync.WaitGroup, waitCtx context.Context, inr <-chan *Entry, ) (chan *Entry) {

    outr := make(chan *Entry)
    wg.Add(1)
    go func() {
        defer close(outr)
        defer wg.Done()
    loop:
        for {
            var e *Entry
            select {
            case <-waitCtx.Done():
                break loop
            case v1, ok := <-inr:
                if !ok {
                    break loop
                }
                e = v1
            }

            if e.err == nil {
                b, err := ioutil.ReadFile(e.FilePath)
                if err != nil {
                    e.err = err
                }
                if e.err == nil {
                    e.Digest = md5.Sum(b)
                }
            }

            select {
            case <-waitCtx.Done():
                break loop
            case outr <- e:
            }
        }
    }()

    return outr
}

func MD5All(dir string, waitCtx context.Context) (map[string][md5.Size]byte, error) {

    wg := new(sync.WaitGroup)
    r := make(map[string][md5.Size]byte)
    //var paths = walkFiles(done,dir)
    pathsCh := listFiles(wg, waitCtx, dir)
    dsCh := digest(wg, waitCtx, pathsCh)
loop:
    for {

        select {
        case <-waitCtx.Done():
            break loop
        case v, ok := <-dsCh:
            if !ok {
                break loop
            }
            if v.err != nil {
                return nil, v.err
            }
            r[v.FilePath] = v.Digest
        case <-time.After(time.Second * 3):
            return nil, fmt.Errorf("timeout after 3sec")
        }
    }
    wg.Wait()
    return r, nil
}

func TestHash(dir string) {

    waitCtx, cancel := context.WithCancel(context.Background())
    _ = cancel
    result, err := MD5All(dir, waitCtx)
    if err != nil {
        panic(err)
    }

    var paths []string

    for path, _ := range result {
        paths = append(paths, path)
    }
    sort.Strings(paths)

    for _, path := range paths {
        shortName := filepath.Base(path)
        fmt.Printf("%x %v\n", result[path], shortName)
    }
}

func main() {
    d, _ := os.Executable()
    TestHash(filepath.Dir(d))
}
