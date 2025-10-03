package hnadlers

import (
	"encoding/json"
	"net/http"
)

// StatusHandler is the HTTP handler for the /status endpoint.
// It responds with a JSON object indicating the application is live.
// StatusHandler is the HTTP handler for the /status endpoint.
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	//Check the HTTP Method, if not GET, return error
	if r.Method != http.MethodGet {
		// * status error 405, wrong method
		w.WriteHeader(http.StatusMethodNotAllowed)
		// Optionally write a JSON response for clarity
		w.Write([]byte(`{"error": "Method not allowed. Only GET is supported."}`))
		return
	}
	// ---------------------------------------------

	// Set the response header to indicate JSON content
	w.Header().Set("Content-Type", "application/json")

	//*  Set a 200 OK status code
	w.WriteHeader(http.StatusOK)

	// Create and encode the response
	response := map[string]string{
		"status":      "live",
		"application": "Go Simple Web Server",
		"message":     "Application is live and running!",
	}
	json.NewEncoder(w).Encode(response)
}
