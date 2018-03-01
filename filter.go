package tracefilter

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jaegertracing/jaeger/model"
	"github.com/lodastack/alarm-adapter/requests"
)

// unit: min
const defaultPullInterval = 2

/* API
{
	"httpstatus": 200,
	"data": ["monitor-router-", "test-", "picus-"]
}
*/
const defaultAPI = "http://trace.test.com/static/service.json"

var tracefilter *Filter

func init() {
	tracefilter = New(defaultAPI)
	tracefilter.Update()
	ticker := time.NewTicker(time.Duration(defaultPullInterval) * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				tracefilter.Update()
			}
		}
	}()
}

// Filter iFeng custom spanfilter
type Filter struct {
	Addr     string
	services []string
	mu       sync.RWMutex
}

// Resp API data struct
type Resp struct {
	Status int      `json:"httpstatus"`
	Data   []string `json:"data"`
}

// New Filter
func New(addr string) *Filter {
	r := &Filter{
		Addr: addr,
	}
	return r
}

// Update cache data
func (r *Filter) Update() error {
	var resp Resp
	response, err := requests.Get(r.Addr)
	if err != nil {
		return err
	}

	if response.Status == 200 {
		err = json.Unmarshal(response.Body, &resp)
		if err != nil {
			return err
		}
		r.mu.Lock()
		r.services = resp.Data
		r.mu.Unlock()
		return nil
	}
	return fmt.Errorf("get all ns failed: code %d", response.Status)
}

// Check span reported
func Check(span *model.Span) bool {
	tracefilter.mu.RLock()
	services := tracefilter.services
	tracefilter.mu.RUnlock()
	for _, s := range services {
		if strings.HasPrefix(span.Process.ServiceName, s) {
			return checkTime(span)
		}
	}
	return false
}

func checkTime(span *model.Span) bool {
	stime := span.StartTime.Unix()
	// From 2018-01-01 to 2118-01-01
	if stime > 1514739661 && stime < 4670413261 {
		return true
	}
	return false
}
