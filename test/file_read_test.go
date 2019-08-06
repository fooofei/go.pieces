package go_pieces

import (
	"bufio"
	"encoding/hex"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"gotest.tools/assert"
)

// include the \n as line end, and \r\n as line end
const content = `this
is
a` + "\r\nsample"
const expectTotalLen = len(content)
const expectLineCnt = 4

func TestReadFileByLine(t *testing.T) {
	rr := strings.NewReader(content)
	expect := []string{
		"this", "is", "a", "sample",
	}

	totalLen1 := 0
	lineCnt1 := 0

	scanr := bufio.NewScanner(rr)
	for i := 0; scanr.Scan(); i++ {
		lineTxt := scanr.Text()
		lineBytes := scanr.Bytes()
		assert.Equal(t, expect[i], lineTxt)
		lineCnt1++
		totalLen1 += len(lineBytes)
	}

	totalLen := 0
	for _, t := range expect {
		totalLen += len(t)
	}

	assert.Equal(t, totalLen1, totalLen)
	assert.Equal(t, lineCnt1, expectLineCnt)
}

func TestReadFileByRune(t *testing.T) {
	rr := strings.NewReader(content)

	var ch rune
	var sz int
	var err error
	var totalLen int

	expect := []rune{
		116, 104, 105, 115, 10, 105, 115, 10,
		97, 13, 10, 115, 97, 109, 112, 108, 101,
	}
	_ = sz

	for i := 0; ; i++ {
		ch, sz, err = rr.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatal(err)
		}
		assert.Equal(t, ch, expect[i])
		totalLen += 1
	}

	assert.Equal(t, totalLen, expectTotalLen)
}

func TestReadFileOnceForAll(t *testing.T) {
	rdr := strings.NewReader(content)
	// ioutil.ReadFile()
	c, err := ioutil.ReadAll(rdr)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, hex.EncodeToString(c), "746869730a69730a610d0a73616d706c65")
	assert.Equal(t, len(c), expectTotalLen)
}

func TestReadFileByBytes(t *testing.T) {
	rr := strings.NewReader(content)

	buf := make([]byte, 100)

	rsize, err := rr.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, rsize, expectTotalLen)
	assert.Equal(t, hex.EncodeToString(buf[:rsize]), "746869730a69730a610d0a73616d706c65")
}
