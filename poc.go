package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

var monitors = make(map[string]Monitor)

type Monitor struct {
	stop   chan struct{}
	result chan int
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/start", startMonitor).Methods("POST")
	r.HandleFunc("/stop/{id}", stopMonitor).Methods("POST")

	serveAddress := "127.0.0.1:8080"
	fmt.Println("serving on", serveAddress)
	http.ListenAndServe(serveAddress, r)
}


func startMonitor(w http.ResponseWriter, _ *http.Request) {
	id := strconv.FormatInt(time.Now().UnixNano(), 10) // TODO might want to switch this for some other UUID
	stop := make(chan struct{})
	result := make(chan int)
	go monitor(result, stop, id, 2*time.Second)

	monitors[id] = Monitor{stop: stop, result: result}

	fmt.Fprintf(w, "%s\n", id)
}

func monitor(result chan<- int, stop <-chan struct{}, id string, interval time.Duration) {
	fmt.Println("monitoring...", id)
	defer close(result)

	for {
		// non-blocking stop channel
		select {
		case <-stop:
			fmt.Println("monitor stopped", id)
			// calculate and write result to channel
			result <- 69
			return
		default:
			fmt.Println("running", id)
			// perform http call here
			// and save result in func-local var
		}

		time.Sleep(interval)
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
	result := <-monitor.result

	fmt.Fprintf(w, "stopping %s, got result %d\n", id, result)
}
