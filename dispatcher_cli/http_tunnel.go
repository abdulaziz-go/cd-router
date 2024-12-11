package main

import (
	"bytes"
	dispatcher "cloud-dispatcher-router/dispatcher_server"
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
	"net/http"
	urlPackage "net/url"
)

type HttpTunnel struct {
	Host    string `bson:"host"`
	Token   string `bson:"token"`
	Warning string `bson:"warning"`
	Error   string `bson:"error"`
}

func openHttpTunnel(port int, subdomain string, ctx context.Context) {
	query := fmt.Sprintf("port=%d&username=%s&version=%s", port, subdomain, version)
	url := urlPackage.URL{Scheme: "wss", Host: baseHost, Path: "/_create_tunnel/", RawQuery: query} // agar baseHost ssl/tls support qilsa wss directly ip orqali bo'layotgan bo'lsa ws
	fmt.Println(url.String())
	ws, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		log.Fatalf("Error connecting to %s: %s", baseHost, err.Error())
	}
	defer ws.Close()

	var tunnel HttpTunnel
	_, message, err := ws.ReadMessage()
	if err != nil {
		log.Fatalf("Error Reading Message from Server: %s\n", err.Error())
	}
	err = bson.Unmarshal(message, &tunnel)
	if err != nil {
		log.Fatalf("Error while decoding tunnel Info: %s", err.Error())
	}

	if tunnel.Warning != "" {
		fmt.Printf("WARNING: %s", tunnel.Warning)
	}
	if tunnel.Error != "" {
		log.Fatalf(tunnel.Error)
	}
	fmt.Println("Tunnel status: Online")
	fmt.Printf("Forwarded: %s -> localhost:%d", tunnel.Host, port)
	requests := make(chan dispatcher.RequestMessage)
	responses := make(chan []byte)
	defer close(requests)
	defer close(responses)

	go handleHttpRequests(ws, requests)
	go func() {
		for respons := range responses {
			if err := ws.WriteMessage(websocket.BinaryMessage, respons); err != nil {
				log.Printf("Error Sending Message to Server: %s", err)
				return
			}
		}
	}()
out:
	for {
		select {
		case <-ctx.Done():
			break out
		case request := <-requests:
			go handleHTTPRequest(ws, tunnel.Token, port, request)
		}
	}
	fmt.Println("dispatcher tunnel closed")
}

func handleHttpRequests(ws *websocket.Conn, requests chan<- dispatcher.RequestMessage) {
	for {
		var requestMessage dispatcher.RequestMessage
		_, msg, err := ws.ReadMessage()
		if err != nil {
			return
		}
		err = bson.Unmarshal(msg, &requestMessage)
		if err != nil {
			log.Printf("Error decoding Message %s\n", msg)
		}
		requests <- requestMessage
	}
}
func handleHTTPRequest(ws *websocket.Conn, token string, port int, r dispatcher.RequestMessage) {
	url := fmt.Sprintf("http://127.0.0.1:%d%s", port, r.URL)
	request, err := http.NewRequest(r.Method, url, bytes.NewReader(r.Body))
	if err != nil {
		fmt.Printf("Failed to Build Request: %s\n", err.Error())
		return
	}

	for key, val := range r.Header {
		request.Header.Add(key, val)
	}

	var client http.Client
	response, err := client.Do(request)

	if err != nil {
		fmt.Printf("Failed to Perform Request: %s\n", err.Error())
		return
	}

	responseMessage := dispatcher.ResponseMessage{}

	responseMessage.Header = make(map[string]string)
	for name, values := range response.Header {
		responseMessage.Header[name] = values[0]
	}

	if response.Body != nil {
		responseMessage.Body, _ = ioutil.ReadAll(response.Body)
		response.Body.Close()
	}

	responseMessage.Status = response.StatusCode
	responseMessage.RequestId = r.ID
	responseMessage.Token = token

	message, err := bson.Marshal(responseMessage)
	if err != nil {
		fmt.Printf("Error Encoding Response Message: %s\n", err.Error())
		return
	}
	err = ws.WriteMessage(websocket.BinaryMessage, message)

	if err != nil {
		fmt.Printf("Error Sending Message to Server: %s", err)
		return
	}
	fmt.Println(r.Method, r.URL, responseMessage.Status)
}
