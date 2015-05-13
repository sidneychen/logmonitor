package main

import (
	"fmt"
	"log"
	"os"
	"time"

	_ "syscall"

	"github.com/ActiveState/tail"
)

var (
	PHPLogs       = []string{"php_api", "php_guideapi", "php_serviceapi", "php_simulateapi", "php_simulateguideapi", "php_simulateserviceapi"}
	AccessLogs    = []string{"nginx_access_api", "nginx_access_guideapi", "nginx_access_serviceapi", "nginx_access_simulateapi", "nginx_access_simulateguideapi", "nginx_access_simulateserviceapi"}
	MysqlSlowLogs = []string{"mysql_slow_api"}

	TailConfig = tail.Config{Follow: true, Location: &tail.SeekInfo{Offset: 0, Whence: os.SEEK_END}}
)

func main() {

	t := time.Now()
	month := t.Format("2006-01")
	day := t.Format("2006-01-02")
	m := NewMonitor()

	for _, fn := range PHPLogs {
		filepath := fmt.Sprintf("/data/logcollection/online/%s/%s/%s_%s.log", month, fn, fn, day)
		t, err := tail.TailFile(filepath, TailConfig)
		if err != nil {
			log.Println(err)
			continue
		}
		p := PHPLog(*t)
		m.AddWatcher(&p)
	}
	for _, fn := range AccessLogs {
		filepath := fmt.Sprintf("/data/logcollection/online/%s/%s/%s_%s.log", month, fn, fn, day)
		t, err := tail.TailFile(filepath, TailConfig)
		if err != nil {
			log.Println(err)
			continue
		}
		p := AccessLog(*t)
		m.AddWatcher(&p)
	}
	for _, fn := range MysqlSlowLogs {
		filepath := fmt.Sprintf("/data/logcollection/online/%s/%s/%s_%s.log", month, fn, fn, day)
		t, err := tail.TailFile(filepath, TailConfig)
		if err != nil {
			log.Println(err)
			continue
		}
		p := MysqlSlowLog(*t)
		m.AddWatcher(&p)
	}

	m.Start()

	// handle interupt signal
	//	ch := make(chan os.Signal)
	//	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	//	log.Printf("Signal catched: %v\r\n", <-ch)

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
