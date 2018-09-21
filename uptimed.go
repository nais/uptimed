package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-multierror"
	m "github.com/nais/uptimed/monitor"
	"net/http"
	"net/url"
	"strconv"
)

var monitors = make(map[string]m.Monitor)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/start", startMonitor).Methods("POST")
	r.HandleFunc("/stop/{id}", stopMonitor).Methods("POST")

	serveAddress := "127.0.0.1:8080"
	fmt.Println("serving on", serveAddress)
	http.ListenAndServe(serveAddress, r)
}

func startMonitor(w http.ResponseWriter, r *http.Request) {
	endpoint, interval, timeout, err := getMonitorSettings(r.URL.Query())

	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "unable to get monitor settings: %s\n", err)
		return
	}

	monitor := m.New(endpoint, interval, timeout)
	monitor.Run()

	monitors[monitor.Id] = monitor

	fmt.Fprintf(w, "%s\n", monitor.Id)
}

func stopMonitor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	monitor, exist := monitors[id]

	if ! exist {
		w.WriteHeader(404)
		fmt.Fprintf(w, "monitor with id %s not found\n", id)
		return
	}

	if len(monitor.Result) == 0 {
		close(monitor.Stop)
	}
	fmt.Fprintf(w, "stopping %s, got result %d\n", id, <-monitor.Result)
}

func getMonitorSettings(input url.Values) (*url.URL, int, int, error) {
	var result = &multierror.Error{}

	endpointStr := input.Get("endpoint")
	if len(endpointStr) == 0 {
		multierror.Append(result, fmt.Errorf("no endpoint query parameter provided"))
	}

	endpoint, err := url.ParseRequestURI(endpointStr)
	fmt.Println(endpoint, err)
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
