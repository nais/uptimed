package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestStartMonitor(t *testing.T) {

	t.Run("happy path", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/start?endpoint=http://test.no&interval=1&timeout=2", nil)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(startMonitor)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v. Body:\n%s",
				status, http.StatusOK, rr.Body.String())
		}
	})

	t.Run("invalid settings", func(t *testing.T) {
		for _, invalidSettings := range []string{
			"/start",
			"/start?endpoint=test.no",
			"/start?endpoint=http://test.no&timeout=a",
			"/start?endpoint=http://test.no&interval=a"} {
			req, _ := http.NewRequest("POST", invalidSettings, nil)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(startMonitor)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status == http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v. Body:\n%s",
					status, http.StatusBadRequest, rr.Body.String())
			}
		}
	})
}

func TestDefaultSettings(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		values := url.Values{
			"endpoint": []string{"http://test.no"},
			"interval": []string{"1"},
			"timeout":  []string{"2"},
		}
		endpoint, interval, timeout, e := getMonitorSettings(values)
		assert.NoError(t, e)
		assert.Equal(t, "http://test.no", endpoint.String())
		assert.Equal(t, 1, interval)
		assert.Equal(t, 2, timeout)
	})

	t.Run("defaults applied when not provided", func(t *testing.T) {
		values := url.Values{
			"endpoint": []string{"http://test.no"},
		}
		endpoint, interval, timeout, e := getMonitorSettings(values)
		assert.NoError(t, e)
		assert.Equal(t, "http://test.no", endpoint.String())
		assert.Equal(t, 2, interval)
		assert.Equal(t, 1800, timeout)
	})

	t.Run("invalid settings", func(t *testing.T) {
		for _, invalidSettings := range []url.Values{
			{"": []string{}},
			{"endpoint": []string{"test.no"}},
			{
				"interval": []string{"a"},
				"endpoint": []string{"http://test.no"},
			},
			{
				"timeout":  []string{"a"},
				"endpoint": []string{"http://test.no"},
			}} {
			_, _, _, e := getMonitorSettings(invalidSettings)
			assert.Error(t, e)
		}
	})
}

func TestTimeoutIntervalCorrelation(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		values := url.Values{
			"endpoint": []string{"http://test.no"},
			"interval": []string{"2"},
			"timeout":  []string{"2"},
		}
		_, _, _, e := getMonitorSettings(values)
		assert.EqualErrorf(t, e, "1 error occurred:\n\t* timeout must be longer than interval\n\n", "")
	})
}

func TestStopMonitor(t *testing.T) {
	t.Run("test monitor not found", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/stop/nonsense", nil)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(stopMonitor)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}
