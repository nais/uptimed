package monitor

import (
	"fmt"
	"github.com/rs/xid"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

// Monitor contains the data fields for each monitor we set up
type Monitor struct {
	Id             string
	endpoint       *url.URL
	stop           chan struct{}
	timeout        <-chan time.Time
	ticker         <-chan time.Time
	interval       int
	RequestCount   int
	StartTime      time.Time
	StopTime       time.Time
	FailedRequests []FailedRequest
}

// FailedRequest contains time and error messages from monitoring
type FailedRequest struct {
	Timestamp time.Time
	Reason    string
}

// Result prints the results after monitoring
func (m *Monitor) Result() float64 {
	log.Printf("Monitor %s - requests: %d, failed requests: %d, started: %s, stopped: %s, %s",
		m.Id,
		m.RequestCount,
		len(m.FailedRequests),
		m.StartTime.Format(time.RFC3339),
		m.StopTime.Format(time.RFC3339),
		m.PrintFailed())
	return calculateUptimePercent(m.RequestCount, len(m.FailedRequests))

}

// PrintFailed prints the failed requests should there be any
func (m *Monitor) PrintFailed() string {
	failedStr := fmt.Sprintf("errorcount: %d\n", len(m.FailedRequests))
	for _, failed := range m.FailedRequests {
		failedStr += fmt.Sprintf("%s: %s\n", failed.Timestamp, failed.Reason)
	}

	return failedStr
}

func calculateUptimePercent(total, failed int) float64 {
	successful := float64(total) - float64(failed)
	return (successful / float64(total)) * 100
}

// New creates a new Monitor struct
func New(endpoint *url.URL, interval int, timeout int) Monitor {
	return Monitor{
		endpoint: endpoint,
		Id:       xid.New().String(),
		stop:     make(chan struct{}),
		interval: interval,
		ticker:   time.NewTicker(time.Duration(interval) * time.Second).C,
		timeout:  time.NewTimer(time.Duration(timeout) * time.Second).C,
	}
}

// Stop stops the monitoring and sleeps to allow goroutine to return
func (m *Monitor) Stop() {
	close(m.stop)
	log.Printf("Monitor %s stopped", m.Id)
	time.Sleep((time.Duration(m.interval) * time.Second) + (250 * time.Millisecond))
}

// Run controls the action of the monitoring of applications
func (m *Monitor) Run() {
	m.StartTime = time.Now()
	go func() {
		defer func() {
			m.StopTime = time.Now()
		}()

		log.Printf("Monitor %s started for endpoint %s", m.Id, m.endpoint)

		for range m.ticker {
			select {
			case <-m.stop:
				return
			case <-m.timeout:
				log.Println("Timed out", m.Id)
				return
			default:
				m.RequestCount++

				response, err := http.DefaultClient.Get(m.endpoint.String())
				if err != nil {
					m.FailedRequests = append(m.FailedRequests,
						FailedRequest{time.Now(), fmt.Sprintf("error performing http request: %s", err)})
				} else {
					if response.StatusCode != 200 { //TODO: maybe make this configurable by query param
						body, err := ioutil.ReadAll(response.Body)
						if err != nil {
							m.FailedRequests = append(m.FailedRequests,
								FailedRequest{time.Now(), fmt.Sprintf("could not read response body: %s", err)})
						}

						m.FailedRequests = append(m.FailedRequests,
							FailedRequest{time.Now(), fmt.Sprintf("http status code: %d\nresponse body: %s", response.StatusCode, string(body))})

					}
				}
				response.Body.Close()
			}
		}
	}()
}
