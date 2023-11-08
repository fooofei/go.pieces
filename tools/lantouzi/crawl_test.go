package main

import (
	"gotest.tools/v3/assert"
	"regexp"
	"testing"
	"time"
)

func TestRegex1(t *testing.T) {
	text := `
		2018年12月07日项目回款公告
		
	懒投资的小伙伴们：
	2018年12月06日16点30分至2018年12月07日16点30分回款情况如下，请您及时查收懒投资账户。
完成标的还款9453笔，共21,376,359.63元。结束到期智选服务406笔，共10,681,000.00元。		
	#news-signature {position: relative; height: 236px;padding-top: 20px; font-size: 14px; color: #666;}
	#news-signature > .sp-line {height: 1px;background-color: #ccc;}
	#news-signature > .qr-code {position: absolute;top: 45px;right: 0;width: 170px;height: 170px;border:5px solid #ccc;}
	#news-signature > .qr-label {position: absolute;top: 232px;right: 0;width: 180px;text-align: center; color: #ccc;}
	#news-signature > .qr-label > em {color: #f75f52;margin-left: 5px;font-weight: bold;}
	#news-signature > .desc {margin-top: 45px;margin-bottom: 35px;}
	#news-signature > .contact {margin-bottom: 10px;}
	
	懒投资微信版lantouzicom
	
	懒人专线：400-807-8000
	客服邮箱：kefu@lantouzi.com

	
`
	// https://lantouzi.com/post/1440
	re := regexp.MustCompile("(?s)至(?P<year>.*)年(?P<month>.*)月(?P<day>.*)日.*到期智选服务(?P<count>\\d+)笔，共(?P<money>.*?)元")
	result := re.FindStringSubmatch(text)
	assert.Equal(t, result != nil, true)
	assert.Equal(t, result[1], "2018")
	assert.Equal(t, result[2], "12")
	assert.Equal(t, result[3], "07")
}

func TestFloatlify(t *testing.T) {
	text := "10,681,000.00"
	num := floatlify(text)
	assert.Equal(t, num == 1.0681e+07, true)
}

func TestSubDay(t *testing.T) {
	endDay := time.Time{}
	assert.Equal(t, endDay.Year(), 1)
	assert.Equal(t, endDay.Month(), time.Month(1))
	assert.Equal(t, endDay.Day(), 1)
	endDay.In(time.Now().Location())

	endDay = endDay.AddDate(2019-1, 12-1, 8-1)
	assert.Equal(t, endDay.Year(), 2019)
	assert.Equal(t, endDay.Month(), time.Month(12))
	assert.Equal(t, endDay.Day(), 8)

	endDay = endDay.Add(-time.Hour * 24)
	assert.Equal(t, endDay.Year(), 2019)
	assert.Equal(t, endDay.Month(), time.Month(12))
	assert.Equal(t, endDay.Day(), 7)
}
