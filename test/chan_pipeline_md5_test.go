package go_pieces

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
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"
)

type PathDigest struct {
	FilePath string
	Digest   []byte
	Err      error
}

// md5 of path
// different with md5.Sum(b)
// md5.Sum read file once for all
// we use io.Copy, iter read
func md5SumPath(path string) ([]byte, error) {
	d := md5.New()
	fr, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(d, fr)
	_ = fr.Close()
	if err != nil {
		return nil, err
	}
	return d.Sum(make([]byte, 0)), nil
}

// listFiles of dir
// put result to empty chan
// iterator way
func listFiles(wg *sync.WaitGroup, waitCtx context.Context, dir string) chan *PathDigest {
	pathCh := make(chan *PathDigest)

	wg.Add(1)
	go func() {
		// make an anonymous func, make sure we can use return
		// and not skip wg.Done()
		func() {
			files, err := ioutil.ReadDir(dir)
			if err != nil {
				select {
				case pathCh <- &PathDigest{FilePath: dir, Err: err}:
				case <-waitCtx.Done():
				}
				return
			}
		loop:
			for _, f := range files {
				select {
				case <-waitCtx.Done():
					break loop
				case pathCh <- &PathDigest{FilePath: filepath.Join(dir, f.Name())}:
				}
			}
		}()
		close(pathCh)
		wg.Done()
	}()
	return pathCh
}

// walkFiles of dir
// push result to empty chan
// iterator way
func walkFiles(wg *sync.WaitGroup, waitCtx context.Context, dir string) chan *PathDigest {
	pathCh := make(chan *PathDigest)
	wg.Add(1)
	go func() {
		walkFn := func(path string, info os.FileInfo, err error) error {
			// see we are told done or not, as early exit
			select {
			case <-waitCtx.Done():
				return filepath.SkipDir
			default:
			}

			// check have err or not
			if err != nil {
				select {
				case <-waitCtx.Done():
				case pathCh <- &PathDigest{FilePath: dir, Err: err}:
				}
				return nil
			}
			// check is file or not
			if !info.Mode().IsRegular() {
				return nil
			}
			// push result to chan
			select {
			case pathCh <- &PathDigest{FilePath: path}:
				return nil
			case <-waitCtx.Done():
				return filepath.SkipDir
			}
		}
		err := filepath.Walk(dir, walkFn)
		if err != nil {
			select {
			case pathCh <- &PathDigest{FilePath: dir, Err: err}:
			case <-waitCtx.Done():
			}
		}
		close(pathCh)
		wg.Done()
	}()
	return pathCh
}

// digest a filepath
// read from pathCh
// write to digestCh
// pathCh digestCh share same item type
func digest(wg *sync.WaitGroup, waitCtx context.Context, pathCh <-chan *PathDigest) chan *PathDigest {

	digestCh := make(chan *PathDigest)
	wg.Add(1)
	go func() {
		var more bool
	loop:
		for {
			var e *PathDigest

			select {
			case <-waitCtx.Done():
				break loop
			case e, more = <-pathCh:
				if !more {
					break loop
				}
			}
			if e.Err != nil {
				continue
			}
			e.Digest, e.Err = md5SumPath(e.FilePath)
			select {
			case <-waitCtx.Done():
				break loop
			case digestCh <- e:
			}
		}
		close(digestCh)
		wg.Done()
	}()
	return digestCh
}

// enum files of dir, md5 digest of each file
// return map[filepath]=digest
func md5All(dir string, waitCtx context.Context) (map[string][]byte, error) {
	var err error
	const limit = 3

	wg := new(sync.WaitGroup)
	r := make(map[string][]byte)
	subWaitCtx, cancel := context.WithCancel(waitCtx)
	//var paths = walkFiles(done,dir)
	pathCh := listFiles(wg, subWaitCtx, dir)
	digestCh := digest(wg, subWaitCtx, pathCh)
loop:
	for {
		select {
		case <-waitCtx.Done():
			break loop
		case v, more := <-digestCh:
			if !more {
				break loop
			}
			if v.Err != nil {
				err = v.Err
			} else {
				r[v.FilePath] = v.Digest
				if len(r) >= limit {
					cancel()
					break loop
				}
			}
		}
	}
	wg.Wait()
	if len(r) == 0 && err != nil {
		return nil, err
	}
	return r, nil
}

func TestHash(t *testing.T) {

	dir := "/usr/sbin/"
	waitCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	result, err := md5All(dir, waitCtx)
	if err != nil {
		t.Fatal(err)
	}
	cancel()

	var paths []string

	for path, _ := range result {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		shortName := filepath.Base(path)
		t.Logf("md5 %v = %x", shortName, result[path])
	}
}
