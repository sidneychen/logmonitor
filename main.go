package main

import (
	"bufio"
	"bytes"
	"encoding/json"
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
	// stop self next day
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

	// handle interupt signal
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Signal catched: %v\r\n", <-ch)

	m.Stop()

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
	filenames []string
	wg        *sync.WaitGroup
}

func NewMonitor(filenames []string) *Monitor {
	m := &Monitor{
		filenames: filenames,
		wg:        new(sync.WaitGroup),
	}

	args := append([]string{"-f"}, filenames...)
	m.cmd = exec.Command("tail", args...)
	return m
}

func (m *Monitor) Start() {
	log.Println("Start Monitoring log files:", m.filenames)
	stdout, err := m.cmd.StdoutPipe()
	onErrorExit(err)

	err = m.cmd.Start()
	onErrorExit(err)

	go func() {
		m.wg.Add(1)
		defer func() { m.wg.Done() }()
		reader := bufio.NewReader(stdout)
		for {
			b, _, err := reader.ReadLine()
			if err == io.EOF {
				return
			}
			line := string(b)
			if strings.Contains(line, "ERROR") {
				warn(line)
			}
		}
	}()

	err = m.cmd.Wait()
	onErrorExit(err)
}

func (m *Monitor) Stop() {
	m.cmd.Process.Signal(syscall.SIGINT)
	m.wg.Wait()
}

func warn(msg string) {
	log.Println("Catch an error, error:", msg)
	url := "http://service.192.168.94.26.xip.io/service/call"
	param := make(map[string]interface{})
	param["pack_type"] = "json"
	data := make(map[string]interface{})
	data["service"] = "Api_Warn"
	data["method"] = "notice"
	params := []string{"common", "日志告警: " + msg}
	data["params"] = params
	var err error
	bData, err := json.Marshal(data)
	param["data"] = string(bData)
	checkError(err)

	jsonData, err := json.Marshal(param)
	checkError(err)

	log.Println("Warning params:", string(jsonData))
	buffer := bytes.NewBuffer(jsonData)
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
