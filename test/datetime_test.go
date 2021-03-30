package go_pieces

import (
	"github.com/araddon/dateparse"
	"gotest.tools/v3/assert"
	"testing"
	"time"
)

const (
	DateTimeFmt = "2006/01/02 15:04:05"
)

func TestTimestampBaseTime(t *testing.T) {
	// One day of time
	nextTime := time.Date(0, 0, 0, 1, 0, 0, 0, time.Now().Location())
	// the timeOff begin with base
	baseTime := time.Date(0, 0, 0, 0, 0, 0, 0, time.Now().Location())
	t.Logf("baseTime=%v Date(0)", baseTime.Format(time.RFC3339))
	t.Logf("baseTime + 1h = nextTime = %v", nextTime.Format(time.RFC3339))

	dur := nextTime.Sub(baseTime)
	assert.Equal(t, dur, time.Hour*1)
}

func TestTimestampNextDayOff(t *testing.T) {
	now := time.Now()
	todayZero := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayZero2 := now.Add(time.Hour * 8).Truncate(time.Hour * 24).Add(-time.Hour * 8)
	t.Logf("todayZero= %v", todayZero.Format(time.RFC3339))
	t.Logf("todayZero2= %v", todayZero2.Format(time.RFC3339))
	assert.Equal(t, todayZero, todayZero2)

	nextDayZero := todayZero.Add(time.Hour * 24)
	t.Logf("nextDayZero= %v", nextDayZero.Format(time.RFC3339))
}

// truncate vs round

func TestTimestampRecoverBack(t *testing.T) {
	now := time.Now()
	t.Logf("now= %v", now.Format(time.RFC3339))
	unix := now.UnixNano()
	t.Logf("unix= %v", unix)
	b := time.Unix(0, unix)
	b.In(time.UTC)
	t.Logf("recover back Time= %v UTC= %v local= %v",
		b.Format(time.RFC3339), b.UTC().Format(time.RFC3339),
		b.Local().Format(time.RFC3339))
}

// 截断
func TestTimestampTruncate(t *testing.T) {
	//  datetime_test.go:57: now= 2019-08-06T18:28:15+08:00
	//    datetime_test.go:59: utcNow= 2019-08-06T10:28:15Z
	//    datetime_test.go:60: local.unix= 1565087295 utc.unix= 1565087295 unixNano= 1565087295075860000
	//    datetime_test.go:65: now.Truncate(24 h)= 2019-08-06T08:00:00+08:00
	//    datetime_test.go:66: utcNow.Truncate(24 h)= 2019-08-06T00:00:00Z
	//    datetime_test.go:68: now.Truncate().unix= 1565049600
	//    datetime_test.go:69: utcNow.Truncate().unix= 1565049600

	//dur := time.Second * 60
	dur := time.Hour * 24
	now := time.Now()
	t.Logf("now= %v", now.Format(time.RFC3339))
	utcNow := now.UTC()
	t.Logf("utcNow= %v ", utcNow.Format(time.RFC3339))
	t.Logf("local.unix= %v utc.unix= %v unixNano= %v",
		now.Unix(), utcNow.Unix(), utcNow.UnixNano())

	nowTr := now.Truncate(dur)
	utcNowTr := utcNow.Truncate(dur)
	t.Logf("now.Truncate(24 h)= %v", nowTr.Format(time.RFC3339))
	t.Logf("utcNow.Truncate(24 h)= %v", utcNowTr.Format(time.RFC3339))

	t.Logf("now.Truncate().unix= %v", nowTr.Unix())
	t.Logf("utcNow.Truncate().unix= %v", utcNowTr.Unix())

}

// 四舍五入
func TestTimestampRound(t *testing.T) {
	//    datetime_test.go:83: now= 2019-08-06T18:30:51+08:00 utcNow= 2019-08-06T10:30:51Z
	//    datetime_test.go:89: now.Round(24 h)= 2019-08-06T08:00:00+08:00 utcNow.Round(24 h)= 2019-08-06T00:00:00Z
	now := time.Now()
	utcNow := now.UTC()

	t.Logf("now= %v utcNow= %v", now.Format(time.RFC3339), utcNow.Format(time.RFC3339))

	dur := time.Hour * 24
	nowRnd := now.Round(dur)
	utcNowRnd := utcNow.Round(dur)

	t.Logf("now.Round(24 h)= %v utcNow.Round(24 h)= %v",
		nowRnd.Format(time.RFC3339), utcNowRnd.Format(time.RFC3339))

}

func TestTimestamp1(t *testing.T) {
	//    datetime_test.go:107: from unixNano= 1548639286891265000 time= 2019-01-28T09:34:46+08:00 utc= 2019-01-28T01:34:46Z
	//    datetime_test.go:113: Truncate(24 h)= 2019-01-28 08:00:00 +0800 CST Round(24 h)= 2019-01-28 08:00:00 +0800 CST
	//    datetime_test.go:120: from unixNano= 1565102637887862000 time= 2019-08-06T22:43:57+08:00 utc= 2019-08-06T14:43:57Z
	//    datetime_test.go:124: Truncate(24 h)= 2019-08-06 08:00:00 +0800 CST Round(24 h)= 2019-08-07 08:00:00 +0800 CST

	var fromUnixNano int64 = 1548639286891265000
	fromTime := time.Unix(0, fromUnixNano)
	fromTime.In(time.UTC)

	t.Logf("from unixNano= %v time= %v utc= %v",
		fromUnixNano, fromTime.Format(time.RFC3339),
		fromTime.UTC().Format(time.RFC3339))

	dur := time.Hour * 24

	t.Logf("Truncate(24 h)= %v Round(24 h)= %v",
		fromTime.Truncate(dur), fromTime.Round(dur))

	var fromUnixNano2 int64 = 1565102637887862000
	fromTime2 := time.Unix(0, fromUnixNano2)
	fromTime2.In(time.UTC)

	t.Logf("from unixNano= %v time= %v utc= %v",
		fromUnixNano2, fromTime2.Format(time.RFC3339),
		fromTime2.UTC().Format(time.RFC3339))

	t.Logf("Truncate(24 h)= %v Round(24 h)= %v",
		fromTime2.Truncate(dur), fromTime2.Round(dur))

}

func TestTimestampFromString1(t *testing.T) {
	s := "2020-12-14 10:44:04.650+0800"
	const timeFmt = "2006-01-02 15:04:05.999999-0700"
	timeValue, err := time.Parse(timeFmt, s)
	assert.NilError(t, err)
	assert.Equal(t, timeValue.UTC().Format(time.RFC3339), "2020-12-14T02:44:04Z")
	assert.Equal(t, timeValue.UnixNano(), int64(1607913844650000000))
}

func TestParseGMT(t *testing.T) {
	dt, err := dateparse.ParseAny("2021-03-29T08:15:05.005GMT")
	assert.NilError(t, err)
	assert.Equal(t, dt.Format(time.RFC3339), "2021-03-29T08:15:05Z")
}
