package gateway

import (
	"net/http"
	"net/http/httptest"
	"raweb/handlers"
)

// REST handler to fetch the CRL via the SOAP handler
func RESTGetCRLHandler(w http.ResponseWriter, r *http.Request) {
	// Allow only GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create a new GET request for the SOAP handler
	req, err := http.NewRequest(http.MethodGet, "/get-crl", nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Use an in-memory ResponseRecorder to capture the SOAP handler response
	rec := httptest.NewRecorder()

	// Call the SOAP handler to get the CRL
	handlers.GetLatestCRLHandler(rec, req)

	// Forward the SOAP handler's response back to the REST client
	w.Header().Set("Content-Type", rec.Header().Get("Content-Type"))
	w.WriteHeader(rec.Code)
	w.Write(rec.Body.Bytes())
}
