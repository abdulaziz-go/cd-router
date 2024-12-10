package dispatcher

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
)

type Tunnel struct {
	port               int
	numOfRequestServed int
	host               string
	token              string
	conn               *websocket.Conn
	requestChan        chan RequestMessage
	responseChan       chan ResponseMessage
	requests           map[uuid.UUID]RequestMessage
}

func (d Dispatcher) GetTunnelByHost(host string) (*Tunnel, error) {
	t, ok := d.tunnels[host]
	if !ok {
		return t, errors.New("Tunnel doesn't exist")
	}
	return t, nil
}

func (d Dispatcher) OpenTunnel(host string, port int, conn *websocket.Conn) *Tunnel {
	token := generateToken()
	requests := make(map[uuid.UUID]RequestMessage)
	requestChan, responseChan := make(chan RequestMessage), make(chan ResponseMessage)
	tunnel := Tunnel{
		host:         host,
		port:         port,
		conn:         conn,
		token:        token,
		requests:     requests,
		requestChan:  requestChan,
		responseChan: responseChan,
	}
	fmt.Println("Opened Tunnel: ", tunnel)
	d.tunnels[host] = &tunnel
	return &tunnel
}

func (d Dispatcher) CloseTunnel(host string) {
	tunnel, ok := d.tunnels[host]
	if !ok {
		return
	}
	fmt.Printf("Closed tunnel %s , Number of Requests Served: %d", host, tunnel.numOfRequestServed)
	close(tunnel.requestChan)
	close(tunnel.responseChan)
	delete(d.tunnels, host)
}
