package monitor

import (
	"fmt"
	"github.com/rs/xid"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Monitor struct {
	Id             string
	endpoint       *url.URL
	Stop           chan struct{}
	Result         chan int
	timeout        <-chan time.Time
	ticker         <-chan time.Time
	FailedRequests []failedRequest
}

type failedRequest struct {
	timestamp time.Time
	statusCode int
	responseBody string
}

func New(endpoint *url.URL, interval, timeout int) Monitor {
	return Monitor{
		endpoint:       endpoint,
		Id:             xid.New().String(),
		Stop:           make(chan struct{}),
		Result:         make(chan int),
		ticker:         time.NewTicker(time.Duration(interval) * time.Second).C,
		timeout:        time.NewTimer(time.Duration(timeout) * time.Second).C,
		FailedRequests: []failedRequest{},
	}
}

func (m *Monitor) Run() {
	go func() {
		fmt.Println("monitoring...", m.Id)
		defer close(m.Result)

		for range m.ticker {
			select {
			case <-m.Stop:
				fmt.Println("monitor stopped", m.Id)

				if len(m.FailedRequests) > 0 {
					m.Result <- -1
					return
				}

				m.Result <- 1
				return
			case <-m.timeout:
				fmt.Println("timed out", m.Id)
				m.Result <- -1
				return
			default:

				req, err := http.NewRequest("GET", m.endpoint.String(), nil)
				if err != nil {
					fmt.Printf("request creation failed: %s", err)
					m.Result <- -1
					return
				}

				response, err := http.DefaultClient.Do(req)
				if err != nil {
					fmt.Printf("could not create http client: %s", err)
					m.Result <- -1
					return
				}

				fmt.Println("Status: ", response.StatusCode)
				if response.StatusCode != 200 {
					body, err := ioutil.ReadAll(response.Body)
					if err != nil {
						fmt.Errorf("could not read response body: %s", err)
						m.Result <- -1
						return
					}
					s := append(m.FailedRequests, failedRequest{time.Now(), response.StatusCode, string(body)})
					fmt.Println(s)
					return
				}
			}
		}
	}()
}
