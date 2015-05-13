package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/ActiveState/tail"
	simplejson "github.com/bitly/go-simplejson"
)

var (
	AccessLogTimeThreshold = 2 // seconds
	AccessLogPattern       = regexp.MustCompile(`^.*? "([^\s]+)" "([^\s]+)" "([^"]+)"\s+$`)
)

type LogWatcher interface {
	Watch()
}

type PHPLog tail.Tail

func (t *PHPLog) Watch() {
	for line := range t.Lines {
		if strings.Contains(line.Text, "ERROR") {
			Warn(line.Text)
		}
	}
}

type AccessLog tail.Tail

func (t *AccessLog) Watch() {
	for line := range t.Lines {
		elapse := GetElapseTime(line.Text)
		if elapse > 2 && !strings.Contains(line.Text, "transactions") {
			Warn(fmt.Sprintf("请求时间过长，%fs %s", elapse, line.Text))
		}
	}
}

func GetElapseTime(line string) float32 {
	//	line := `{"message":"121.206.142.74 - - [12/May/2015:11:38:59 +0800] GET /brands/189/salons/288/workers/1651/transactions?status=0&type=0&size=500&fields=basic,servicecombos,productioncombos,vipitems,price,payment,hairstyle_photo&access_token=0d1aea075d5c15395e3bcaf181cc10642e0616fcEDC00D5&version=202020123 HTTP/1.1 \"200\" 3740 \"-\" \"\\xE5\\xBD\\xA2\\xE8\\xB1\\xA1\\xE5\\xAE\\xB6.\\xE5\\x95\\x86\\xE5\\xAE\\xB6 2.2.1 rv:202020123 (iPad; iPhone OS 8.3; zh_CN)\" \"-\"\"5.97\"  \"1.673\"","@version":"1","@timestamp":"2015-05-12T11:39:00.268+08:00","type":"nginx_access_api","host":"222.77.191.150:10086","path":"/data/httplogs/api.meiyegj.com-access.log"}`
	js, err := simplejson.NewJson([]byte(line))
	if err != nil {
		log.Println(err)
		return 0.0
	}
	message, err := js.Get("message").String()
	match := AccessLogPattern.FindStringSubmatch(message)
	if len(match) == 0 {
		log.Println("Cannot match the access log line, ", message)
		return 0.0
	}
	elapse, err := strconv.ParseFloat(string(match[2]), 32)
	if err != nil {
		log.Println(err)
		return 0.0
	}
	return float32(elapse)
}

type MysqlSlowLog tail.Tail

type SlowLog struct {
	Content string
	Found   bool
}

func (t *MysqlSlowLog) Watch() {
	var s SlowLog

	for line := range t.Lines {
		js, err := simplejson.NewJson([]byte(line.Text))
		if err != nil {
			log.Println(err)
			continue
		}
		message, err := js.Get("message").String()
		if strings.HasPrefix(message, "# Time") {
			s.Found = false
			Warn("Mysql慢查询告警: " + s.Content)
			s.Content = ""
		} else if strings.HasPrefix(message, "# Time") {
			s.Found = true
		}

		if s.Found {
			s.Content += message
		}
	}
}
