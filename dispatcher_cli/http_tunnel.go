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
	"sync"
)

type HttpTunnel struct {
	Host    string `bson:"host"`
	Token   string `bson:"token"`
	Warning string `bson:"warning"`
	Error   string `bson:"error"`
}

var writeMutex sync.Mutex

func openHttpTunnel(port int, subdomain string, ctx context.Context) {
	query := fmt.Sprintf("port=%d&username=%s&version=%s", port, subdomain, version)
	url := urlPackage.URL{
		Scheme:   "wss",
		Host:     baseHost,
		Path:     "/_create_tunnel/",
		RawQuery: query,
	}
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
		for response := range responses {
			writeMutex.Lock()
			err := ws.WriteMessage(websocket.BinaryMessage, response)
			writeMutex.Unlock()

			if err != nil {
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
			go handleHTTPRequest(responses, tunnel.Token, port, request)
		}
	}
	fmt.Println("dispatcher tunnel closed")
}

func handleHttpRequests(ws *websocket.Conn, requests chan<- dispatcher.RequestMessage) {
	for {
		var requestMessage dispatcher.RequestMessage
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			return
		}

		err = bson.Unmarshal(msg, &requestMessage)
		if err != nil {
			log.Printf("Error decoding Message %s\n", msg)
			continue
		}

		requests <- requestMessage
	}
}

func handleHTTPRequest(responses chan<- []byte, token string, port int, r dispatcher.RequestMessage) {
	url := fmt.Sprintf("http://127.0.0.1:%d%s", port, r.URL)
	request, err := http.NewRequest(r.Method, url, bytes.NewReader(r.Body))
	if err != nil {
		log.Printf("Failed to Build Request: %s\n", err.Error())
		return
	}

	for key, val := range r.Header {
		request.Header.Add(key, val)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("Failed to Perform Request: %s\n", err.Error())
		return
	}
	defer response.Body.Close()

	responseMessage := dispatcher.ResponseMessage{
		Header:    make(map[string]string),
		Token:     token,
		Status:    response.StatusCode,
		RequestId: r.ID,
	}

	for name, values := range response.Header {
		responseMessage.Header[name] = values[0]
	}

	if response.Body != nil {
		responseMessage.Body, err = ioutil.ReadAll(response.Body)
		if err != nil {
			log.Printf("Error reading response body: %v", err)
			return
		}
	}

	message, err := bson.Marshal(responseMessage)
	if err != nil {
		log.Printf("Error Encoding Response Message: %s\n", err.Error())
		return
	}

	responses <- message

	fmt.Printf("%s %s %d\n", r.Method, r.URL, responseMessage.Status)
}
