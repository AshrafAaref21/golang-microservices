package main

import (
	"log"
	"net/http"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/util"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleRidersWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade to WebSocket", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		http.Error(w, "Missing userID query parameter", http.StatusBadRequest)
		return
	}

	log.Printf("Rider %s connected via WebSocket", userID)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from rider %s: %v", userID, err)
			break
		}
		log.Printf("Received message from rider %s: %s", userID, string(message))
	}

}

func handleDriversWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade to WebSocket", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		http.Error(w, "Missing userID query parameter", http.StatusBadRequest)
		return
	}

	packageSlug := r.URL.Query().Get("packageSlug")
	if packageSlug == "" {
		http.Error(w, "Missing packageSlug query parameter", http.StatusBadRequest)
		return
	}

	log.Printf("Driver %s connected via WebSocket for package %s", userID, packageSlug)

	type Driver struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		ProfilePicture string `json:"profilePicture"`
		CarPlate       string `json:"carPlate"`
		PackageSlug    string `json:"packageSlug"`
	}

	msg := contracts.WSMessage{
		Type: "driver.cmd.register",
		Data: Driver{
			ID:             userID,
			Name:           "Driver Name",
			ProfilePicture: util.GetRandomAvatar(1),
			CarPlate:       "Car Plate",
			PackageSlug:    packageSlug,
		},
	}

	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Error sending initial data to driver %s: %v", userID, err)
		return
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from driver %s: %v", userID, err)
			break
		}
		log.Printf("Received message from driver %s: %s", userID, string(message))
	}
}
