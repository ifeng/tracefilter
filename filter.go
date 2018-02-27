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
const defaultAPI = "http://www.ifeng.com"

var tracefilter *Filter

func init() {
	tracefilter = New(defaultAPI)
	ticker := time.NewTicker(time.Duration(defaultPullInterval) * time.Minute)
	for {
		select {
		case <-ticker.C:
			tracefilter.Update()
		}
	}
}

type Filter struct {
	Addr     string
	services []string
	Interval int
	mu       sync.RWMutex
}

type Resp struct {
	Status int      `json:"httpstatus"`
	Data   []string `json:"data"`
}

func New(addr string) *Filter {
	r := &Filter{
		Addr:     addr,
		Interval: defaultPullInterval,
	}
	return r
}

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

func Check(span *model.Span) bool {
	tracefilter.mu.RLock()
	services := tracefilter.services
	tracefilter.mu.RUnlock()
	for _, s := range services {
		if strings.HasPrefix(span.Process.ServiceName, s) {
			return true
		}
	}
	return false
}
