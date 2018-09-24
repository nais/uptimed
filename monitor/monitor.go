package monitor

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/rs/xid"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

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

func (m *Monitor) Result() string {
	return fmt.Sprintf("uptime=%.2f%%\n%d / %d\n%sstarted: %s\nstopped: %s",
		calculateUptimePercent(m.RequestCount, len(m.FailedRequests)),
		m.RequestCount-len(m.FailedRequests),
		m.RequestCount,
		m.PrintFailed(),
		m.StartTime.Format(time.RFC3339),
		m.StopTime.Format(time.RFC3339))
}

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

type FailedRequest struct {
	Timestamp time.Time
	Reason    string
}

func New(urlQuery url.Values) (Monitor, error) {

	endpoint, interval, timeout, err := getMonitorSettings(urlQuery)
	if err != nil {
		return Monitor{}, fmt.Errorf("unable to get monitor settings: %s\n", err)
	}

	return Monitor{
		endpoint: endpoint,
		Id:       xid.New().String(),
		stop:     make(chan struct{}),
		interval: interval,
		ticker:   time.NewTicker(time.Duration(interval) * time.Second).C,
		timeout:  time.NewTimer(time.Duration(timeout) * time.Second).C,
	}, nil
}

func (m *Monitor) Stop() {
	close(m.stop)
	time.Sleep((time.Duration(m.interval) * time.Second) + (250 * time.Millisecond)) // allow goroutine to return
}

func (m *Monitor) Run() {
	m.StartTime = time.Now()
	go func() {
		defer func() {
			m.StopTime = time.Now()
		}()

		fmt.Println("monitor started", m.Id)

		for range m.ticker {
			select {
			case <-m.stop:
				fmt.Println("monitor stopped", m.Id)
				return
			case <-m.timeout:
				fmt.Println("timed out", m.Id)
				return
			default:
				m.RequestCount++

				response, err := http.DefaultClient.Get(m.endpoint.String())
				if err != nil {
					m.FailedRequests = append(m.FailedRequests,
						FailedRequest{time.Now(), fmt.Sprintf("error performing http request: %s", err)})
					return
				}

				if response.StatusCode != 200 { //TODO: maybe make this configurable by query param
					body, err := ioutil.ReadAll(response.Body)
					if err != nil {
						m.FailedRequests = append(m.FailedRequests,
							FailedRequest{time.Now(), fmt.Sprintf("could not read response body: %s", err)})
						return
					}

					m.FailedRequests = append(m.FailedRequests,
						FailedRequest{time.Now(), fmt.Sprintf("http status code: %d\nresponse body: %s", response.StatusCode, string(body))})
					return
				}
			}
		}
	}()
}

func getMonitorSettings(input url.Values) (*url.URL, int, int, error) {
	var result = &multierror.Error{}

	endpointStr := input.Get("endpoint")
	if len(endpointStr) == 0 {
		multierror.Append(result, fmt.Errorf("no endpoint query parameter provided"))
	}

	endpoint, err := url.ParseRequestURI(endpointStr)

	if err != nil {
		multierror.Append(result, fmt.Errorf("invalid endpoint %s: %s", endpointStr, err))
	}

	interval, err := parseIntOrDefault(input.Get("interval"), 2)
	if err != nil {
		multierror.Append(result, err)
	}

	timeout, err := parseIntOrDefault(input.Get("timeout"), 1800)
	if err != nil {
		multierror.Append(result, err)
	}

	if interval >= timeout {
		multierror.Append(result, fmt.Errorf("timeout must be longer than interval"))
	}

	return endpoint, interval, timeout, result.ErrorOrNil()
}


func parseIntOrDefault(maybeInt string, defaultValue int) (int, error) {
	if len(maybeInt) == 0 {
		return defaultValue, nil
	}

	val, err := strconv.Atoi(maybeInt)

	if err != nil {
		return 0, fmt.Errorf("unable to parse string %s to int: %s", maybeInt, err)
	}

	return val, nil
}