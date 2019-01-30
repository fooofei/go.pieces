package main

import (
    "log"
    "time"
)

const (
    DateTimeFmt = "2006/01/02 15:04:05"
)

// often use for schedule
func timestampNextDayOff() {

    // One day of time
    timeOff := time.Date(0, 0, 0, 1, 0, 0, 0, time.Now().Location())
    log.Printf("we need timeOff= %v", timeOff.Format(DateTimeFmt))
    // the timeOff begin with base
    base := time.Date(0, 0, 0, 0, 0, 0, 0, time.Now().Location())
    log.Printf("the base= %v", base.Format(DateTimeFmt))
    dur := timeOff.Sub(base)
    log.Printf("the timeOff dur= %v", dur)

    now := time.Now()
    log.Printf("now= %v", now.Format(DateTimeFmt))
    dayZero := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
    log.Printf("dayZero= %v", dayZero.Format(DateTimeFmt))
    nextDayZero := dayZero.Add(time.Hour * 24)
    log.Printf("nextDayZero= %v", nextDayZero.Format(DateTimeFmt))

    nextTime := nextDayZero.Add(dur)
    log.Printf("we got the nextTime= %v", nextTime.Format(DateTimeFmt))
    //output:
    //2019/01/21 12:17:43 datetime.go:17: we need timeOff= -0001/11/30 01:00:00
    //2019/01/21 12:17:43 datetime.go:20: the base= -0001/11/30 00:00:00
    //2019/01/21 12:17:43 datetime.go:22: the timeOff dur= 1h0m0s
    //2019/01/21 12:17:43 datetime.go:25: now= 2019/01/21 12:17:43
    //2019/01/21 12:17:43 datetime.go:27: dayZero= 2019/01/21 00:00:00
    //2019/01/21 12:17:43 datetime.go:29: nextDayZero= 2019/01/22 00:00:00
    //2019/01/21 12:17:43 datetime.go:32: we got the nextTime= 2019/01/22 01:00:00
}

func timestampTruncate() {
    //dur := time.Second * 60
    dur := time.Hour * 24
    utcNow := time.Now().UTC()
    log.Printf("utcNow= %v unix= %v unixNano= %v", utcNow.Format(DateTimeFmt), utcNow.Unix(),
        utcNow.UnixNano())

    tr := utcNow.Truncate(dur)
    log.Printf("utcNow.Truncate= %v unix= %v", tr.Format(DateTimeFmt), tr.Unix())

    trUnix := tr.UnixNano()
    b := time.Unix(0, trUnix)
    b.In(time.UTC)
    log.Printf("recover back time= %v UTC= %v",
        b.Format(DateTimeFmt),b.UTC().Format(DateTimeFmt))
    // output:
    //2019/01/21 12:18:51 datetime.go:46: now= 2019/01/21 04:18:51
    //2019/01/21 12:18:51 datetime.go:48: now.Truncate= 2019/01/21 04:18:00
    //2019/01/21 12:18:51 datetime.go:53: recover back time= 2019/01/21 12:18:00 UTC= 2019/01/21 04:18:00
}

func timestampRound() {
    utcNow := time.Now().UTC()
    dur := time.Hour * 24
    rou := utcNow.Round(dur)

    log.Printf("utcNow= %v unix= %v", utcNow.Format(DateTimeFmt), utcNow.Unix())
    log.Printf("utcNow.Round= %v unix= %v", rou.Format(DateTimeFmt), rou.Unix())
    // output:
    //2019/01/21 12:16:58 datetime.go:54: now= 2019/01/21 04:16:58
    //2019/01/21 12:16:58 datetime.go:55: now.Round= 2019/01/21 04:17:00
}

func timestampWhy(){
    unixNano := int64(1548639286891265000)
    b := time.Unix(0,unixNano)
    b.In(time.UTC)

    log.Printf("b= %v unix= %v unixNano= %v", b.Format(DateTimeFmt), b.Unix(), b.UnixNano())
    dur := time.Hour * 24
    tr := b.Truncate(dur)
    tou := b.Round(dur) // 这个是四舍五入，不是全部是进位

    log.Printf("tr= %v unix= %v", tr.Format(DateTimeFmt), tr.Unix())
    log.Printf("tou= %v unix= %v", tou.Format(DateTimeFmt), tou.Unix())
    //2019/01/28 09:40:40 datetime.go:81: b= 2019/01/28 09:34:46 unix= 1548639286 unixNano= 1548639286891265000
    //2019/01/28 09:41:06 datetime.go:86: tr= 2019/01/28 08:00:00 unix= 1548633600
    //2019/01/28 09:41:06 datetime.go:87: tou= 2019/01/28 08:00:00 unix= 1548633600
}

func main() {
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    //timestampTruncate()
    //timestampRound()

    timestampWhy()

}
