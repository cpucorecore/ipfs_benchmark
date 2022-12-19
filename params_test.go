package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRoundRobinHost(t *testing.T) {
	host1 := "192.168.0.11"
	host2 := "192.168.0.12"
	host3 := "192.168.0.13"

	p.Hosts = []string{
		host1,
		host2,
		host3,
	}

	h := RoundRobinHost()
	require.Equal(t, host1, h)

	h = RoundRobinHost()
	require.Equal(t, host2, h)

	h = RoundRobinHost()
	require.Equal(t, host3, h)

	h = RoundRobinHost()
	require.Equal(t, host1, h)
}
