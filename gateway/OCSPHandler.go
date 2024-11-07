package gateway

import (
	"io"
	"net/http"
)

// OCSPHandler fetches the OCSP response and serves it as a downloadable file to the client
func OCSPHandler(w http.ResponseWriter, r *http.Request) {
	// Perform a GET request to the OCSP status URL
	resp, err := http.Get("http://167.71.219.39:8080/ejbca/publicweb/status/ocsp")
	if err != nil {
		http.Error(w, "Failed to retrieve OCSP status", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Set headers for file download
	w.Header().Set("Content-Disposition", "attachment; filename=ocsp")
	w.Header().Set("Content-Type", "application/octet-stream")

	// Stream the response body to the client
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, "Failed to send OCSP response to client", http.StatusInternalServerError)
		return
	}
}
