package observer

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	errCountStat = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "observer_errors",
			Help: "count of errors encountered by all monitors",
		},
		[]string{},
	)
	obsCountStat = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "observations",
			Help: "count of observations performed by all monitors",
		},
		[]string{},
	)
	monCountStat = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "monitors",
			Help: "count of all monitors configured",
		},
		[]string{},
	)
)

// Observer is
type Observer struct {
	// log      blog.Logger
	Timeout  time.Duration
	Monitors []*Monitor
}

// Start acts as the supervisor for all monitor coroutines
func (o Observer) Start() {
	runningChan := make(chan bool)

	// start each monitor
	for _, monitor := range o.Monitors {
		go monitor.start()
	}

	// run forever
	<-runningChan
}

// New creates a observer instance from the provided configuration
func New(n NewObs) (*Observer, error) {
	// Setup logging and stats
	// prom, logger := cmd.StatsAndLogging(config.Syslog, config.DebugAddr)
	// prom.MustRegister(errCountStat)
	// prom.MustRegister(obsCountStat)
	// prom.MustRegister(monCountStat)
	// defer logger.AuditPanic()
	// logger.Info(cmd.VersionString())
	return n.asObserver()
}
