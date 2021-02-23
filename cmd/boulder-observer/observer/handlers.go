package observer

import (
	"fmt"
	"time"

	"github.com/miekg/dns"
)

// dispatch calls the handler specified by the `Monitor`
func dispatch(m *Monitor, tick time.Time) error {
	switch m.handler {
	case "dns":
		return m.handlerDNS.run(tick, m.Timeout)
	case "http":
		return m.handlerHTTP.run(tick, m.Timeout)
	}
	return nil
}

func (h HandlerDNS) run(tick time.Time, timeout time.Duration) error {
	fmt.Println(h, timeout)
	m := new(dns.Msg)
	m.SetQuestion(h.Qname, h.Qtype)
	fmt.Println(m)
	c := new(dns.Client)
	c.Net = h.Qproto
	c.ReadTimeout = timeout
	c.Net = h.Qproto
	in, _, _ := c.Exchange(m, h.Qserver)
	fmt.Println(in)
	if a, ok := in.Answer[0].(*dns.A); ok {
		fmt.Println(a.A.String())
	}
	return nil
}

func (h HandlerHTTP) run(tick time.Time, timeout time.Duration) error {
	fmt.Printf("beep boop ♡ ⋆｡˚ pretend we did http things ♡ ⋆｡˚: %d\n", tick.Nanosecond())
	return nil
}
