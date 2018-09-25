package monitor

import (
	"fmt"
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
	case <-monitor.timeout:
		return
	}
}

func TestMonitorStop(t *testing.T) {
	endpoint, _ := url.Parse("http://test.no")
	monitor := New(endpoint, 1, 3)

	monitor.Run()
	monitor.Stop()

	//TODO: use gock to mock two http calls one failing and one successful and check for 1/2 (50%)
	fmt.Println(monitor.Result())
}

func TestNonexistantHost(t *testing.T) {
	endpoint, _ := url.Parse("http://test.nonexistant")
	monitor := New(endpoint, 1, 3)

	monitor.Run()
	time.Sleep(2 * time.Second)
	monitor.Stop()

	result := monitor.Result()
	assert.Contains(t, result, "errorcount: 1")
	assert.Contains(t, result, "uptime=0.00%")
}
