package h2lb

import (
	"context"
	"net"
	"sync"
	"time"
)

// Resolver ...
type Resolver struct {
	lock  sync.RWMutex
	cache map[string][]string
}

// NewResolver refresh specifies the refresh interval of DNS cache.
func NewResolver(refresh time.Duration) *Resolver {
	resolver := &Resolver{
		cache: make(map[string][]string, 0),
	}
	if refresh < time.Second {
		refresh = time.Minute
	}
	go resolver.autoRefresh(refresh)
	return resolver
}

// Get ...
func (a *Resolver) Get(host string) ([]string, error) {
	a.lock.RLock()
	ips, exists := a.cache[host]
	a.lock.RUnlock()
	if exists {
		return ips, nil
	}
	return a.Lookup(host)
}

// Lookup ...
func (a *Resolver) Lookup(host string) ([]string, error) {
	ips, err := net.DefaultResolver.LookupHost(context.Background(), host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, nil
	}
	a.lock.Lock()
	a.cache[host] = ips
	a.lock.Unlock()
	return ips, nil
}

func (a *Resolver) autoRefresh(interval time.Duration) {
	for {
		a.Refresh()
		time.Sleep(interval)
	}
}

// Refresh ...
func (a *Resolver) Refresh() {
	i := 0
	a.lock.RLock()
	addrs := make([]string, len(a.cache))
	for key := range a.cache {
		addrs[i] = key
		i++
	}
	a.lock.RUnlock()
	for _, addr := range addrs {
		a.Lookup(addr)
		time.Sleep(time.Second)
	}
}

// Dialer ...
type Dialer struct {
	Resolver *Resolver
	Dialer   *net.Dialer
}

// DialContext ...
func (a *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	ips, _ := a.Resolver.Get(host)
	for _, ip := range ips {
		conn, err := a.Dialer.DialContext(ctx, network, ip+":"+port)
		if err == nil {
			return conn, nil
		}
	}
	return a.Dialer.DialContext(ctx, network, address)
}
