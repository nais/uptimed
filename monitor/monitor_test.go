package monitor

import (
	"fmt"
	"github.com/h2non/gock"
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

func TestMonitorSuccess(t *testing.T) {
	endpoint, _ := url.Parse("http://test.no")
	monitor := New(endpoint, 1, 3)

	defer gock.Off()
	gock.New(fmt.Sprintf("%s", endpoint)).Reply(200)

	monitor.Run()
	time.Sleep(1*time.Second + 100*time.Millisecond)
	monitor.Stop()

	assert.Equal(t, gock.IsDone(), true)
	assert.Equal(t, 1, monitor.RequestCount)
	assert.Equal(t, 0, len(monitor.FailedRequests))

	fmt.Println(monitor.Result())
}

func TestMonitorFailed(t *testing.T) {
	endpoint, _ := url.Parse("http://test.no")
	monitor := New(endpoint, 1, 3)

	defer gock.Off()
	gock.New(fmt.Sprintf("%s", endpoint)).Reply(500)

	monitor.Run()
	time.Sleep(2 * time.Second)
	monitor.Stop()

	assert.Equal(t, gock.IsDone(), true)
	assert.Equal(t, 1, monitor.RequestCount)
	assert.Equal(t, 1, len(monitor.FailedRequests))

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
