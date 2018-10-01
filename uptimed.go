package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-multierror"
	"github.com/nais/uptimed/health"
	m "github.com/nais/uptimed/monitor"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var bindAddr string

func init() {
	flag.StringVar(&bindAddr, "bind-address", "127.0.0.1:8080", "ip:port where http requests are served")
	flag.Parse()
}

var monitors = make(map[string]*m.Monitor)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	r := mux.NewRouter()
	r.HandleFunc("/start", startMonitor).Methods("POST")
	r.HandleFunc("/stop/{id}", stopMonitor).Methods("POST")
	r.HandleFunc("/isAlive", health.IsAlive)

	log.Println("running @", bindAddr)
	server := &http.Server{Addr: bindAddr, Handler: r}

	go func() {
		log.Fatal(server.ListenAndServe())
	}()

	killSignal := <-interrupt
	switch killSignal {
	case os.Interrupt:
		log.Print("SIGINT received, shutting down gracefully")
	case syscall.SIGTERM:
		log.Print("SIGTERM received, shutting down gracefully")
	}

	if len(monitors) > 0 {
		log.Printf("Waiting for %d monitor(s) to finish", len(monitors))
		time.Sleep(30 * time.Second)
	}
	log.Print("Shutting down uptimed")
	server.Shutdown(context.Background())
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
	monitors[monitor.Id] = &monitor

	fmt.Fprintf(w, "%s\n", monitor.Id)
}

func stopMonitor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	monitor, found := monitors[id]

	if !found {
		w.WriteHeader(404)
		fmt.Fprintf(w, "monitor with id %s not found\n", id)
		return
	}

	monitor.Stop()
	delete(monitors, id)

	fmt.Fprintf(w, "%s\n", monitor.Result())
}

func getMonitorSettings(queryParams url.Values) (*url.URL, int, int, error) {
	var result = &multierror.Error{}

	endpointStr := queryParams.Get("endpoint")
	if len(endpointStr) == 0 {
		multierror.Append(result, fmt.Errorf("no endpoint query parameter provided"))
	}

	endpoint, err := url.ParseRequestURI(endpointStr)

	if err != nil {
		multierror.Append(result, fmt.Errorf("invalid endpoint %s: %s", endpointStr, err))
	}

	interval, err := parseIntOrDefault(queryParams.Get("interval"), 2)
	if err != nil {
		multierror.Append(result, err)
	}

	timeout, err := parseIntOrDefault(queryParams.Get("timeout"), 1800)
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
