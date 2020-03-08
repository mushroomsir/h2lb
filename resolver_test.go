package h2lb

import (
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
