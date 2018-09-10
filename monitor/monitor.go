package monitor

import (
	"fmt"
	"github.com/rs/xid"
	"net/url"
	"time"
)

type Monitor struct {
	Id       string
	Endpoint *url.URL
	Stop     chan struct{}
	Result   chan int
	Timeout  <-chan time.Time
	Ticker   <-chan time.Time
}

func New(endpoint *url.URL, interval, timeout int) Monitor {
	return Monitor{
		Endpoint: endpoint,
		Id:       xid.New().String(),
		Stop:     make(chan struct{}),
		Result:   make(chan int),
		Ticker:   time.NewTicker(time.Duration(interval) * time.Second).C,
		Timeout:  time.NewTimer(time.Duration(timeout) * time.Second).C,
	}
}

func (m *Monitor) Run() {
	go func() {
		fmt.Println("monitoring...", m.Id)
		defer close(m.Result)

		for range m.Ticker {
			select {
			case <-m.Stop:
				fmt.Println("monitor stopped", m.Id)
				// calculate and write result to channel
				m.Result <- 69
				return
			case <-m.Timeout:
				fmt.Println("timed out", m.Id)
				m.Result <- -1
				return
			default:
				fmt.Println("running", m.Id)
				// perform http call here
				// and save result in func-local var
			}
		}
	}()
}
