package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	m "github.com/nais/uptimed/monitor"
	"net/http"
)

var bindAddr string

func init() {
	flag.StringVar(&bindAddr, "bind-address", "127.0.0.1:8080", "ip:port where http requests are served")
	flag.Parse()
}

var monitors = make(map[string]*m.Monitor)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/start", startMonitor).Methods("POST")
	r.HandleFunc("/stop/{id}", stopMonitor).Methods("POST")

/*	go func() {
		for {
			for id, monitor := range monitors {
				fmt.Println("Verifying monitor: ", id)
				if monitor.StopTime.IsZero() {
					fmt.Println("Deleting monitor: ", id)
					delete(monitors, id)
				}

			}
			time.Sleep(100 * time.Second)
		}
	}()*/

	fmt.Println("running @", bindAddr)
	http.ListenAndServe(bindAddr, r)

}

func startMonitor(w http.ResponseWriter, r *http.Request) {
	monitor, err := m.New(r.URL.Query())
	if err != nil {

	}

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

	fmt.Fprintf(w, "%s\n", monitor.Result())
}

