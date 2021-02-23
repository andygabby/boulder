package observer

import (
	"net/url"
	"time"
)

type Monitor struct {
	// logger  blog.Logger
	period      time.Duration
	Timeout     time.Duration
	handler     string
	handlerDNS  *HandlerDNS
	handlerHTTP *HandlerHTTP
}

type HandlerDNS struct {
	Qproto  string
	Qname   string
	Qserver string
	Qtype   uint16
}

type HandlerHTTP struct {
	uri url.URL
	tls bool
}

func (m Monitor) supervisor() *time.Ticker {
	ticker := time.NewTicker(m.period)
	go func() {
		for {
			select {
			case tick := <-ticker.C:
				dispatch(&m, tick)
			}
		}
	}()
	return ticker
}

func (m Monitor) start() *time.Ticker {
	return m.supervisor()
}

// getMonitors is a utility function for loading monitors from config
func getMonitors(configs []NewMon, timeout int) ([]*Monitor, error) {
	var monitors []*Monitor
	for _, c := range configs {
		monitors = append(monitors, c.asMonitor(timeout))
	}
	return monitors, nil
}
