package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

// include the \n as line end, and \r\n as line end
const content = `this
is
a` + "\r\nsample"


func ExampleReadFileByLine(){
	rdr := strings.NewReader(content)
	var totalLen int
	var lineCnt int
	scanner := bufio.NewScanner(rdr)
	for i:=0; scanner.Scan(); i+=1{
		bs := scanner.Bytes()
		fmt.Printf("[%v] Read %v, %v\n",i, hex.EncodeToString(bs), scanner.Text())
		lineCnt += 1
		totalLen += len(bs)
	}
	fmt.Printf("totalLen=%v, lineCnt=%v\n", totalLen, lineCnt)
	// output:
	//[0] Read 74686973, this
	//[1] Read 6973, is
	//[2] Read 61, a
	//[3] Read 73616d706c65, sample
	//totalLen=13, lineCnt=4

}

func ExampleReadFileByRune(){
	rdr := strings.NewReader(content)

	var ch rune
	var sz int
	var err error
	var totalLen int
	for {
		ch,sz,err = rdr.ReadRune()
		if err != nil{
			if err == io.EOF{
				break
			}
			panic(err)
		}
		fmt.Printf("[%v]Read ch=%02x size=%v\n",totalLen, (ch), sz)
		totalLen += 1
	}
	fmt.Printf("totalLen=%v\n", totalLen)
	// output:
	//[0]Read ch=74 size=1
	//[1]Read ch=68 size=1
	//[2]Read ch=69 size=1
	//[3]Read ch=73 size=1
	//[4]Read ch=0a size=1
	//[5]Read ch=69 size=1
	//[6]Read ch=73 size=1
	//[7]Read ch=0a size=1
	//[8]Read ch=61 size=1
	//[9]Read ch=0d size=1
	//[10]Read ch=0a size=1
	//[11]Read ch=73 size=1
	//[12]Read ch=61 size=1
	//[13]Read ch=6d size=1
	//[14]Read ch=70 size=1
	//[15]Read ch=6c size=1
	//[16]Read ch=65 size=1
	//totalLen=17
}

func ExampleReadFileOnceAll() {
	var rdr * strings.Reader
	rdr = strings.NewReader(content)
	// ioutil.ReadFile()
	if c, err:= ioutil.ReadAll(rdr); err == nil {
		fmt.Printf("Read Hex %v\n", hex.EncodeToString(c))
		fmt.Printf("Read Bytes Array %v\n", c)
		fmt.Printf("totalLen=%v\n", len(c))
	}else{
		fmt.Fprintf(os.Stdout, "%v", err)
	}
	// output:
	//Read Hex 746869730a69730a610d0a73616d706c65
	//Read Bytes Array [116 104 105 115 10 105 115 10 97 13 10 115 97 109 112 108 101]
	//totalLen=17

}

func ExampleGetFuncNameFromAnotherFile(){
	p := reflect.ValueOf(ExampleGetFuncNameFromAnotherFile).Pointer()
	pName := runtime.FuncForPC(p).Name()
	// pName is fullpath
	pName = filepath.Base(pName)
	fmt.Printf("FuncName=%v\n", pName)
	fmt.Printf("FuncName2=%v\n",GetFuncName(ExampleGetFuncNameFromAnotherFile))
	//output:
	// FuncName=go_pieces.ExampleGetFuncNameFromAnotherFile
	//FuncName2=ExampleGetFuncNameFromAnotherFile

}


func ExampleGetCurDirFromOtherFile(){
	var a = GetCurrentDir()
	var b = filepath.Base(a)
	fmt.Printf("curdir_name=%v\n",b)
	//output:curdir_name=go_pieces
}