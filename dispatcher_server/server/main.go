package main

import (
	dispatcher "cloud-dispatcher-router/dispatcher_server"
	"flag"
	"github.com/gorilla/mux"
)

var baseHost string

func main() {
	flag.StringVar(&baseHost, "host", "cloud-dispatcher.uz", "Base Host")
	flag.Parse()
	d := dispatcher.New(baseHost)
	r := mux.NewRouter()
	r.HandleFunc("/create_tunnel", d.SocketHandler)
}
