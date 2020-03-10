package h2lb

import (
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/http2"
)

// Transport ...
type Transport struct {
	*http2.Transport
	HttpTransport *http.Transport
	Resolver      *Resolver

	lock              sync.Mutex
	pool              map[string]*http2.Transport // keys is host:port
	transportPoolOnce sync.Once
}

// RoundTrip ...
func (a *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	a.transportPoolOnce.Do(a.initTransport)

	if !(req.URL.Scheme == "https" || (req.URL.Scheme == "http" && a.Transport.AllowHTTP)) {
		return a.HttpTransport.RoundTrip(req)
	}
	tr, err := a.GetTransport(req)
	if err != nil {
		return nil, err
	}
	return tr.RoundTrip(req)
}

func (a *Transport) initTransport() {
	if a.Transport == nil {
		a.Transport = &http2.Transport{}
	}
	if a.HttpTransport == nil {
		a.HttpTransport = &http.Transport{}
	}
	if a.HttpTransport.DialContext == nil && a.Resolver != nil {
		d := &Dialer{
			Resolver: a.Resolver,
			Dialer:   &net.Dialer{},
		}
		a.HttpTransport.DialContext = d.DialContext
	}
	if a.pool == nil {
		a.pool = make(map[string]*http2.Transport)
	}
}

// GetTransport ...
func (a *Transport) GetTransport(req *http.Request) (*http2.Transport, error) {
	addrs, err := a.getAddrs(req.Host)
	if err != nil {
		return nil, err
	}
	x := 0
	if len(addrs) > 1 {
		rand.Seed(time.Now().UnixNano())
		x = rand.Intn(len(addrs))
	}
	var current *http2.Transport
	port := GetPort(req.URL.Scheme, req.URL.Host)
	for i, addr := range addrs {
		key := net.JoinHostPort(addr, port)

		a.lock.Lock()
		_, ok := a.pool[key]
		if !ok {
			a.pool[key] = a.clone()
		}
		if i == x {
			current = a.pool[key]
		}
		a.lock.Unlock()
	}
	return current, nil
}

func (a *Transport) getAddrs(host string) ([]string, error) {
	if a.Resolver != nil {
		return a.Resolver.Get(host)
	}
	return net.LookupHost(host)
}

func (a *Transport) clone() *http2.Transport {
	t := &http2.Transport{
		DialTLS:                    a.Transport.DialTLS,
		TLSClientConfig:            a.Transport.TLSClientConfig,
		DisableCompression:         a.Transport.DisableCompression,
		AllowHTTP:                  a.Transport.AllowHTTP,
		MaxHeaderListSize:          a.Transport.MaxHeaderListSize,
		StrictMaxConcurrentStreams: a.Transport.StrictMaxConcurrentStreams,
	}
	return t
}

// GetPort ...
func GetPort(scheme string, host string) string {
	_, port, err := net.SplitHostPort(host)
	if err != nil {
		port = "443"
		if scheme == "http" {
			port = "80"
		}
	}
	return port
}
