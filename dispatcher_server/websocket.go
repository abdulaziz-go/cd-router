package dispatcher

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/gosimple/slug"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strconv"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (d Dispatcher) SocketHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()
	query := r.URL.Query()
	usernames := query["username"]
	ports := query["port"]
	if len(usernames) != 1 || len(ports) != 1 {
		return
	}
	username := usernames[0]
	username = slug.Make(username)
	port, _ := strconv.Atoi(ports[0])
	host := fmt.Sprintf("%s.%s", username, d.baseHost)
	if _, err = d.GetTunnelByHost(host); err == nil {
		errMessage := fmt.Sprintf("Tunnel %s is busy, try different subdomain.", host)
		message := ErrorMessage{errMessage}
		messageContent, _ := bson.Marshal(message)
		ws.WriteMessage(websocket.BinaryMessage, messageContent)
		ws.Close()
		return
	}
	tunnel := d.OpenTunnel(host, port, ws)
	defer d.CloseTunnel(tunnel.host)
	message := TunnelMessage{tunnel.host, tunnel.token}
	messageContent, err := bson.Marshal(message)
	ws.WriteMessage(websocket.BinaryMessage, messageContent)
	go tunnel.DispatchRequests()
	go tunnel.DispatchResponses()

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			break
		}
		response := ResponseMessage{}
		err = bson.Unmarshal(message, &response)
		if err != nil {
			return
		}
		if response.Token != tunnel.token {
			fmt.Println("Auth failed: ", tunnel.host)
			continue
		}
		tunnel.responseChan <- response
	}
}
