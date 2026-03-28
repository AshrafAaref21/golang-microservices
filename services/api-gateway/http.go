package main

import (
	"encoding/json"
	"log"
	"net/http"
	"ride-sharing/services/api-gateway/grpc_clients"
	"ride-sharing/shared/contracts"
)

func handleTripPreview(w http.ResponseWriter, r *http.Request) {
	// if r.Method != http.MethodPost {
	// 	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	// 	return
	// }
	var reqBody previewTripRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validation
	if reqBody.UserID == "" {
		http.Error(w, "UserID is required", http.StatusBadRequest)
		return
	}
	log.Printf("Received trip preview request: %+v", reqBody)

	tripService, err := grpc_clients.NewTripServiceClient()
	if err != nil {
		log.Printf("Failed to connect to trip service: %v", err)
		http.Error(w, "Failed to connect to trip service", http.StatusInternalServerError)
		return
	}
	defer tripService.Close()

	tripData, err := tripService.Client.PreviewTrip(r.Context(), reqBody.ToProto())
	if err != nil {
		log.Printf("Error calling trip service: %v", err)
		http.Error(w, "Failed to preview trip", http.StatusInternalServerError)
		return
	}

	response := contracts.APIResponse{Data: tripData}

	if err := writeJSON(w, http.StatusOK, response); err != nil {
		log.Printf("Failed to write JSON response: %v", err)
	}

}
