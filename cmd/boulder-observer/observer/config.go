package observer

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/letsencrypt/boulder/cmd"
	"github.com/miekg/dns"
)

var (
	errNewObsNoMons      = errors.New("observer config is invalid, 0 monitors configured")
	errNewObsEmpty       = errors.New("observer config is empty")
	errNewObsInvalid     = errors.New("observer config is invalid")
	errNewMonEmpty       = errors.New("monitor config is empty")
	errNewMonInvalid     = errors.New("monitor config is invalid")
	errNewMonDNSEmpty    = errors.New("monitor DNS config is empty")
	errNewMonDNSInvalid  = errors.New("monitor DNS config is invalid")
	errNewMonHTTPEmpty   = errors.New("monitor HTTP config is empty")
	errNewMonHTTPInvalid = errors.New("monitor HTTP config is invalid")

	validQprotos = []string{"udp", "tcp"}
	validQtypes  = map[string]uint16{"A": 1, "TXT": 16, "AAAA": 28, "CAA": 257}
)

// NewObs recieves config needed initialize a new `observer`
type NewObs struct {
	Syslog    cmd.SyslogConfig `yaml:"syslog"`
	DebugAddr string           `yaml:"debugAddr"`
	Timeout   int              `yaml:"timeout"`
	NewMons   []NewMon         `yaml:"monitors"`
}

func (n *NewObs) filterNewMons() (bool, error) {
	i := 0
	for _, m := range n.NewMons {
		// check if monitor is enabled
		if !m.Enabled {
			continue
		}
		// check if monitor is valid
		if ok, err := m.validate(); !ok {
			return false, err
		}
		n.NewMons[i] = m
		i++
	}
	n.NewMons = n.NewMons[:i]
	return true, nil
}

func (n *NewObs) validate() error {
	if n == nil {
		return errNewObsEmpty
	}
	if n.DebugAddr == "" {
		return errNewObsInvalid
	}
	if ok, err := n.filterNewMons(); !ok {
		return err
	}
	if len(n.NewMons) == 0 {
		return errNewObsNoMons
	}
	return nil
}

func (n *NewObs) asObserver() (*Observer, error) {
	err := n.validate()
	if err != nil {
		return nil, err
	}
	monitors, err := getMonitors(n.NewMons, n.Timeout)
	if err != nil {
		return nil, err
	}
	observer := &Observer{
		Timeout:  time.Duration(n.Timeout * 1000000000),
		Monitors: monitors,
	}
	return observer, nil
}

// NewMon recieves config needed initialize a new `Monitor`
type NewMon struct {
	Enabled bool    `yaml:"enabled"`
	Period  int     `yaml:"period"`
	Timeout int     `yaml:"timeout"`
	Handler string  `yaml:"handler"`
	NewDNS  NewDNS  `yaml:"DNS"`
	NewHTTP NewHTTP `yaml:"HTTP"`
}

func (n NewMon) validate() (bool, error) {
	switch strings.ToLower(n.Handler) {
	case "dns":
		if ok, err := n.NewDNS.validate(); !ok {
			return false, err
		}
	case "http":
		if ok, err := n.NewHTTP.validate(); !ok {
			return false, err
		}
	}
	return true, nil
}

func (n NewMon) asMonitor(timeout int) *Monitor {
	// if timeout was not specified, use default
	if n.Timeout == 0 {
		n.Timeout = timeout
	}
	return &Monitor{
		//logger:  logger,
		period:      time.Duration(n.Period * 1000000000),
		handler:     n.Handler,
		Timeout:     time.Duration(n.Timeout * 1000000000),
		handlerDNS:  n.NewDNS.asHandler(),
		handlerHTTP: n.NewHTTP.asHandler(),
	}
}

// NewDNS recieves config needed to call handleDNS
type NewDNS struct {
	QProto  string `yaml:"qproto"`
	QName   string `yaml:"qname"`
	QServer string `yaml:"qserver"`
	QType   string `yaml:"qtype"`
}

func (n NewDNS) normalize() (bool, error) {
	n.QProto = strings.ToLower(n.QProto)
	n.QName = strings.ToLower(n.QName)
	n.QServer = strings.ToLower(n.QServer)
	n.QType = strings.ToLower(n.QType)
	return true, nil
}

func (n NewDNS) validate() (bool, error) {
	// check if qproto is valid
	qprotoValid := func() bool {
		for _, i := range validQprotos {
			if n.QProto == i {
				return true
			}

		}
		return false
	}()
	if !qprotoValid {
		return false, fmt.Errorf("Invalid qproto: %w", errNewMonDNSInvalid)
	}
	// check if qname is a valid FQDN
	if !dns.IsFqdn(dns.Fqdn(n.QName)) {
		return false, fmt.Errorf("Invalid qname: %w", errNewMonDNSInvalid)
	}
	// check if qserver is a valid IP
	if net.ParseIP(n.QServer) == nil {
		// check if it's an FQDN
		if !dns.IsFqdn(dns.Fqdn(n.QServer)) {
			return false, fmt.Errorf("Invalid qserver: %w", errNewMonDNSInvalid)
		}
	}
	// check if qtype is valid
	qtypeValid := func() bool {
		for i := range validQtypes {
			if n.QType == i {
				return true
			}

		}
		return false
	}()
	if !qtypeValid {
		return false, fmt.Errorf("Invalid qtype: %w", errNewMonDNSInvalid)
	}

	return true, nil
}

func (n *NewDNS) asHandler() *HandlerDNS {
	// only one handler can be specified per monitor, if empty, just
	// return an empty handler
	if n == nil {
		return &HandlerDNS{}
	}
	n.normalize()
	n.validate()
	qtype := validQtypes[n.QType]
	return &HandlerDNS{n.QProto, n.QName, n.QServer, qtype}
}

// NewHTTP recieves config needed to call handleHTTP
type NewHTTP struct {
	URI string `yaml:"uri"`
	TLS bool   `yaml:"tls"`
}

func (n NewHTTP) normalize() (bool, error) {
	n.URI = strings.ToLower(n.URI)
	return true, nil
}

func (n NewHTTP) validate() (bool, error) {
	_, err := url.Parse(n.URI)
	if err != nil {
		return false, errNewMonInvalid
	}
	return true, nil
}

func (n *NewHTTP) asHandler() *HandlerHTTP {
	// only one handler can be specified per monitor, if empty, just
	// return an empty handler
	if n == nil {
		return &HandlerHTTP{}
	}
	n.normalize()
	n.validate()
	uri, _ := url.Parse(n.URI)
	return &HandlerHTTP{*uri, n.TLS}
}
