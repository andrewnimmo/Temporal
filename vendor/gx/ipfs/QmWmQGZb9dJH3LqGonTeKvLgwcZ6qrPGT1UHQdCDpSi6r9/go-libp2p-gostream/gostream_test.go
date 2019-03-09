package gostream

import (
	"bufio"
	"context"
	"testing"
	"time"

	libp2p "gx/ipfs/QmRxk6AUaGaKCfzS1xSNRojiAPd7h2ih8GuCdjJBF3Y6GK/go-libp2p"
	multiaddr "gx/ipfs/QmTZBfrPJmjWsCvHEtX5FE6KimVJhsJg5sBbqEFYf4UZtL/go-multiaddr"
	host "gx/ipfs/QmYrWiWM4qtrnCeT3R14jY3ZZyirDNJgwK57q4qFYePgbd/go-libp2p-host"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	peerstore "gx/ipfs/QmaCTz9RkrU13bm9kMB54f7atgqM4qkjDZpRwRoJiWXEqs/go-libp2p-peerstore"
)

// newHost illustrates how to build a libp2p host with secio using
// a randomly generated key-pair
func newHost(t *testing.T, listen multiaddr.Multiaddr) host.Host {
	h, err := libp2p.New(
		context.Background(),
		libp2p.ListenAddrs(listen),
	)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestServerClient(t *testing.T) {
	m1, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/10000")
	m2, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/10001")
	srvHost := newHost(t, m1)
	clientHost := newHost(t, m2)
	defer srvHost.Close()
	defer clientHost.Close()

	srvHost.Peerstore().AddAddrs(clientHost.ID(), clientHost.Addrs(), peerstore.PermanentAddrTTL)
	clientHost.Peerstore().AddAddrs(srvHost.ID(), srvHost.Addrs(), peerstore.PermanentAddrTTL)

	var tag protocol.ID = "/testitytest"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(ctx context.Context) {
		listener, err := Listen(srvHost, tag)
		if err != nil {
			t.Fatal(err)
		}
		defer listener.Close()

		if listener.Addr().String() != srvHost.ID().Pretty() {
			t.Fatal("bad listener address")
		}

		servConn, err := listener.Accept()
		if err != nil {
			t.Fatal(err)
		}
		defer servConn.Close()

		reader := bufio.NewReader(servConn)
		for {
			msg, err := reader.ReadString('\n')
			if err != nil {
				t.Fatal(err)
			}
			if string(msg) != "is libp2p awesome?\n" {
				t.Fatalf("Bad incoming message: %s", msg)
			}

			_, err = servConn.Write([]byte("yes it is\n"))
			if err != nil {
				t.Fatal(err)
			}
			select {
			case <-ctx.Done():
				return
			}
		}
	}(ctx)

	clientConn, err := Dial(clientHost, srvHost.ID(), tag)
	if err != nil {
		t.Fatal(err)
	}

	if clientConn.LocalAddr().String() != clientHost.ID().Pretty() {
		t.Fatal("Bad LocalAddr")
	}

	if clientConn.RemoteAddr().String() != srvHost.ID().Pretty() {
		t.Fatal("Bad RemoteAddr")
	}

	if clientConn.LocalAddr().Network() != Network {
		t.Fatal("Bad Network()")
	}

	err = clientConn.SetDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}

	err = clientConn.SetReadDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}

	err = clientConn.SetWriteDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}

	_, err = clientConn.Write([]byte("is libp2p awesome?\n"))
	if err != nil {
		t.Fatal(err)
	}

	reader := bufio.NewReader(clientConn)
	resp, err := reader.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}

	if string(resp) != "yes it is\n" {
		t.Errorf("Bad response: %s", resp)
	}

	err = clientConn.Close()
	if err != nil {
		t.Fatal(err)
	}
}