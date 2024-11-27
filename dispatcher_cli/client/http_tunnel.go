package main

import (
	"context"
	"fmt"
	"log"
	urlPackage "net/url"
	"os/user"

	"github.com/gorilla/websocket"
	"golang.org/x/net/websocket"
)

type HttpTunnel struct {
	Host    string `bson:"host"`
	Token   string `bson:"token"`
	Warning string `bson:"warning"`
	Error   string `bson:"error"`
}

func openHttpTunnel(port int, subdomain string, ctx context.Context) {
	if subdomain == "" {
		u, err := user.Current()
		if err != nil {
			log.Fatalf("Please specify -subdomain")
		}
		subdomain = u.Username
	}
	query := fmt.Sprintf("port=%d&username=%s&version=%s", port, subdomain, version)
	url := urlPackage.URL{Scheme: "wss", Host: baseHost, Path: "/create_tunnel", RawQuery: query} // agar baseHost ssl/tls support qilsa wss directly ip orqali bo'layotgan bo'lsa ws
	ws, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		log.Fatalf("Error connecting to %s: %s", baseHost, err.Error())
	}
	defer ws.Close()
}
