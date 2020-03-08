package h2lb

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestResolver(t *testing.T) {
	require := require.New(t)

	key := "github.com"
	r := NewResolver(time.Minute)
	ips, err := r.Get(key)
	require.Nil(err)
	require.True(len(ips) > 0)

	require.Equal(ips, r.cache[key])
}

func TestDialContext(t *testing.T) {
	require := require.New(t)
	d := &Dialer{
		Resolver: NewResolver(time.Minute),
		Dialer:   &net.Dialer{},
	}
	conn, err := d.DialContext(context.Background(), "tcp", "github.com:80")
	require.Nil(err)
	require.NotNil(conn)
}
