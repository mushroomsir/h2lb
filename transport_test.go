package h2lb

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransport(t *testing.T) {
	require := require.New(t)
	c := http.Client{}
	c.Transport = &Transport{}
	_, err := c.Get("https://www.ustc.edu.cn")
	require.Nil(err)
}
