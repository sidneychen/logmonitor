package main

import (
	"log"
)

type Monitor struct {
	logs []LogWatcher

	msgCh  chan string
	stopCh chan bool
}

func NewMonitor() *Monitor {
	m := &Monitor{
		msgCh:  make(chan string),
		stopCh: make(chan bool),
	}

	return m
}

func (m *Monitor) AddWatcher(w LogWatcher) {
	m.logs = append(m.logs, w)
}

func (m *Monitor) Start() {
	log.Println("Start Monitoring log files:", m.logs)
	for _, logfile := range m.logs {
		go logfile.Watch()
	}
}

func (m *Monitor) Stop() {
	close(m.stopCh)
}
