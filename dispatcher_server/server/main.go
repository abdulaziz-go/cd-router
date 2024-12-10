package main

import (
	dispatcher "cloud-dispatcher-router/dispatcher_server"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var baseHost string

func main() {
	flag.StringVar(&baseHost, "host", "camelot-register.uz", "Base Host")
	flag.Parse()
	d := dispatcher.New(baseHost)
	r := mux.NewRouter()
	r.HandleFunc("/create_tunnel", d.SocketHandler)
	r.PathPrefix("/").HandlerFunc(d.HttpHandler)
	fmt.Println("Server is running on Port 4200")
	log.Fatal(http.ListenAndServe(":4200", r))
}
