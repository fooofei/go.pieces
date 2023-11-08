package test

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"testing"
	"unicode/utf8"

	gbk "golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"golang.org/x/xerrors"

	"gotest.tools/v3/assert"
)


// GBK -> utf8
// http://mengqi.info/html/2015/201507071345-using-golang-to-convert-text-between-gbk-and-utf-8.html

// encodings
// https://www.haomeili.net/Code/DetailCodes?wd=%E4%BD%A0%E5%A5%BD

// performance of runes -> utf8
// https://stackoverflow.com/questions/29255746/how-encode-rune-into-byte-using-utf8


// equivalent to Java's PBEParametersGenerator.PKCS5PasswordToUTF8Bytes
// also you can
//  []byte(string(rs))
func runesToUTF8(rs []rune) []byte {
	size := 0
	for _, r := range rs {
		size += utf8.RuneLen(r)
	}

	bs := make([]byte, size)
	count := 0
	for _, r := range rs {
		count += utf8.EncodeRune(bs[count:], r)
	}
	return bs
}

func utf8ToRunes(inBytes []byte) ([]rune, error) {
	rs := make([]rune, 0)
	index := 0
	for len(inBytes) > 0 {
		u, n := utf8.DecodeRune(inBytes)
		if n <= 0 {
			return nil, xerrors.Errorf("invalid utf8 char in %v", index)
		}
		index += n
		rs = append(rs, u)
		inBytes = inBytes[n:]
	}
	return rs, nil
}

func gbkToUtf8(gbkBytes []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(gbkBytes), gbk.GBK.NewDecoder())
	return ioutil.ReadAll(reader)
}

func TestUtf8Rune(t *testing.T) {
	// 你好 utf8 encoding
	valueUtf8 := "E4BDA0E5A5BD"
	value, _ := hex.DecodeString(valueUtf8)
	rs, err := utf8ToRunes(value)
	assert.Equal(t, err == nil, true)
	value2 := runesToUTF8(rs)
	assert.DeepEqual(t, value, value2)
}

func TestAsciiToUtf8(t *testing.T) {
	// 你好 gbk encoding
	valueAscii, _ := hex.DecodeString("C4E3BAC3")
	valueUtf8, _ := hex.DecodeString("E4BDA0E5A5BD")
	// Rune=4F60597D utf16 BigEndian
	value, err := gbkToUtf8(valueAscii)
	assert.Equal(t, err == nil, true)
	assert.DeepEqual(t, value, valueUtf8)

}
