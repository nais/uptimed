package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/xid"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var monitors = make(map[string]Monitor)

type Monitor struct {
	id       string
	endpoint string
	stop     chan struct{}
	result   chan int
	timeout  <-chan time.Time
	ticker   <-chan time.Time
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/start", startMonitor).Methods("POST")
	r.HandleFunc("/stop/{id}", stopMonitor).Methods("POST")

	serveAddress := "127.0.0.1:8080"
	fmt.Println("serving on", serveAddress)
	http.ListenAndServe(serveAddress, r)
}

func startMonitor(w http.ResponseWriter, r *http.Request) {
	monitor, err := NewMonitor(r)

	if err != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, err)
		return
	}

	go run(&monitor)

	monitors[monitor.id] = monitor

	fmt.Fprintf(w, "%s\n", monitor.id)
}

func NewMonitor(req *http.Request) (Monitor, error) {
	queryParams := req.URL.Query()

	endpoint := queryParams.Get("endpoint")
	if len(endpoint) == 0 {
		return Monitor{}, fmt.Errorf("no endpoint query parameter provided")
	}

	if _, err := url.Parse(endpoint); err != nil {
		return Monitor{}, fmt.Errorf("invalid endpoint %s: %s", endpoint, err)
	}

	interval, err := parseIntOrDefault(queryParams.Get("interval"), 2)
	if err != nil {
		return Monitor{}, err
	}

	timeout, err := parseIntOrDefault(queryParams.Get("timeout"), 1800)
	if err != nil {
		return Monitor{}, err
	}

	return Monitor{
		endpoint: endpoint,
		id:       xid.New().String(),
		stop:     make(chan struct{}),
		result:   make(chan int),
		ticker:   time.NewTicker(time.Duration(interval) * time.Second).C,
		timeout:  time.NewTimer(time.Duration(timeout) * time.Second).C,
	}, nil
}

func run(m *Monitor) {
	fmt.Println("monitoring...", m.id)
	defer close(m.result)

	for range m.ticker {
		// non-blocking stop channel
		select {
		case <-m.stop:
			fmt.Println("monitor stopped", m.id)
			// calculate and write result to channel
			m.result <- 69
			return
		case <-m.timeout:
			fmt.Println("timed out", m.id)
			m.result <- -1
			return
		default:
			fmt.Println("running", m.id)
			// perform http call here
			// and save result in func-local var
		}
	}
}

func stopMonitor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	monitor := monitors[id]

	if monitor == (Monitor{}) {
		w.WriteHeader(404)
		fmt.Fprintf(w, "monitor with id %s not found\n", id)
		return
	}

	close(monitor.stop)
	fmt.Fprintf(w, "stopping %s, got result %d\n", id, <-monitor.result)
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
