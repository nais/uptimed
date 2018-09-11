package monitor

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
	"time"
)

func TestMonitorTimeout(t *testing.T) {
	endpoint, _ := url.Parse("http://test.no")
	timeout := 1
	monitor := New(endpoint, 1, timeout)

	failedTimeout := make(chan bool)
	go func() {
		time.Sleep(time.Duration(timeout)*time.Second + (100 * time.Millisecond))
		failedTimeout <- true
	}()

	monitor.Run()

	select {
	case <-failedTimeout:
		t.Fatal("monitor should have timed out before this")
	case <-monitor.Timeout:
		return
	}
}

func TestMonitorStop(t *testing.T) {
	endpoint, _ := url.Parse("http://test.no")
	monitor := New(endpoint, 1, 3)

	monitor.Run()
	close(monitor.Stop)
	assert.Equal(t, 69, <-monitor.Result)
}

