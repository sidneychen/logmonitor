package main

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	_ "syscall"

	"github.com/ActiveState/tail"
)

var (
	Logs = map[reflect.Type][]string{
		reflect.TypeOf(PHPLog{}):       []string{"php_api", "php_guideapi", "php_serviceapi", "php_simulateapi", "php_simulateguideapi", "php_simulateserviceapi"},
		reflect.TypeOf(AccessLog{}):    []string{"nginx_access_api", "nginx_access_guideapi", "nginx_access_serviceapi", "nginx_access_simulateapi", "nginx_access_simulateguideapi", "nginx_access_simulateserviceapi"},
		reflect.TypeOf(MysqlSlowLog{}): []string{"mysql_slow_api"},
	}

	TailConfig = tail.Config{Follow: true, Location: &tail.SeekInfo{Offset: 0, Whence: os.SEEK_END}}
)

func main() {

	t := time.Now()
	month := t.Format("2006-01")
	day := t.Format("2006-01-02")
	m := NewMonitor()

	for logType, logs := range Logs {
		for _, fn := range logs {
			filepath := fmt.Sprintf("/data/logcollection/online/%s/%s/%s_%s.log", month, fn, fn, day)
			t, err := tail.TailFile(filepath, TailConfig)
			if err != nil {
				log.Println(err)
				continue
			}
			v := reflect.ValueOf(*t)
			w := v.Convert(logType).Interface().(LogWatcher)

			m.AddWatcher(w)
		}
	}

	m.Start()

	nextday := t.AddDate(0, 0, 1)
	nextday, _ = time.ParseInLocation("2006-01-02 (CST)", nextday.Format("2006-01-02 (CST)"), time.Local)
	after := nextday.Sub(t)
	log.Println("Monitor will stop after ", after)
	<-time.After(after)
	m.Stop()

}

func checkError(err error) {
	if err != nil {
		log.Println(err)
	}
}

func onErrorExit(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
