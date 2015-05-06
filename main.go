package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	// 第二天自动停止
	stopSelfOnNextDay()

	filenames := os.Args[1:]

	t := time.Now()
	month := t.Format("2006-01")
	day := t.Format("2006-01-02")
	var filepath string
	var filepaths []string
	for _, fn := range filenames {
		filepath = fmt.Sprintf("/data/logcollection/online/%s/%s/%s_%s.log", month, fn, fn, day)

		filepaths = append(filepaths, filepath)
	}
	m := NewMonitor(filepaths)
	m.Start()

	// 处理中断信号
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Signal: %v\r\n", <-ch)
	m.Stop()

	log.Println("ssss")

}

func stopSelfOnNextDay() {
	go func() {
		t := time.Now()
		nextday := t.AddDate(0, 0, 1)
		nextday, _ = time.Parse("2006-01-02", nextday.Format("2006-01-02"))
		<-time.After(nextday.Sub(t))
		syscall.Kill(os.Getpid(), syscall.SIGINT)
	}()
}

type Monitor struct {
	cmd       *exec.Cmd
	stopCh    chan bool
	errMsgCh  chan string
	filenames []string
	stopWG    *sync.WaitGroup
}

func NewMonitor(filenames []string) *Monitor {
	m := &Monitor{
		stopCh:    make(chan bool),
		errMsgCh:  make(chan string, 1),
		filenames: filenames,
		stopWG:    new(sync.WaitGroup),
	}

	args := append([]string{"-f"}, filenames...)
	m.cmd = exec.Command("tail", args...)
	go m.errorHandleLoop()
	return m
}

func (m *Monitor) errorHandleLoop() {
	m.stopWG.Add(1)
	defer m.stopWG.Done()
	for {
		select {
		case <-m.stopCh:
			log.Println("Stop Monitor:errorHandlerLoop gorutine")
			return
		case errMsg := <-m.errMsgCh:
			warn(errMsg)
		}
	}
}

func (m *Monitor) Start() {
	log.Println("Watching log files:", m.filenames)
	stdout, err := m.cmd.StdoutPipe()
	onErrorExit(err)

	err = m.cmd.Start()
	onErrorExit(err)

	go func() {

		m.stopWG.Add(1)
		defer m.stopWG.Done()
		reader := bufio.NewReader(stdout)
		for {
			b, _, err := reader.ReadLine()
			if err == io.EOF {
				break
			}
			line := string(b)
			if strings.Contains(line, "ERROR") {
				m.errMsgCh <- line
			}
		}
		log.Println("Stop Monitor:Start gorutine")
	}()
	err = m.cmd.Wait()
	onErrorExit(err)
}

func (m *Monitor) Stop() {
	m.stopCh <- true
	//	m.cmd.Process.Signal(syscall.SIGINT)

	m.stopWG.Wait()
	log.Println("Stop Monitor")
}

func warn(msg string) {
	url := "http://service.192.168.94.26.xip.io/service/call"
	data := `{"pack_type":"json", "data":"{\"service\": \"Api_Warn\", \"method\": \"notice\", \"params\": [\"common\", [\"日志告警: ` + msg + `\"]]}"}`
	log.Println("Catch an error")
	return
	buffer := bytes.NewBufferString(data)
	resp, err := http.Post(url, "application/json", buffer)
	checkError(err)
	_, err = ioutil.ReadAll(resp.Body)
	checkError(err)
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
