package main

import (
	"gitoo.icu/Nexus/Nexus"
	"log"
	"net/http"
)

func main() {
	n := Nexus.New()
	http.HandleFunc("/ws", n.WebSocketService())
	addr := ":8080"
	log.Println("Server started on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
