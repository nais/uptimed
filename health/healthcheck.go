package health

import "net/http"

// IsAlive verfifies application liveness
func IsAlive(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
