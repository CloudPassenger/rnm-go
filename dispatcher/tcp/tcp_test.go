package tcp

import (
	"crypto/rand"
	"net"
	"testing"

	"github.com/CloudPassenger/rnm-go/config"
)

func BenchmarkDispatcher_Auth(b *testing.B) {
	const nServers = 100
	g := new(config.Group)
	for i := 0; i < nServers; i++ {
		var b [10]byte
		rand.Read(b[:])
		g.Servers = append(g.Servers, config.Server{
			Target:  "127.0.0.1:1080",
			PassKey: "MGVgZtEO6Rzrjelc-gwC7oSGsA9hYO3KcWnIu3wCYm4",
		})
	}
	g.BuildPrivateKeys()
	g.BuildUserContextPool(10)
	var data [50]byte
	var d = New(g)
	addr, _ := net.ResolveIPAddr("tcp", "127.0.0.1:50000")
	for i := 0; i < b.N; i++ {
		d.Auth(data[:], g.UserContextPool.GetOrInsert(addr, g.Servers))
	}
}
