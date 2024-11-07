package gateway

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"raweb/handlers"
)

// Struct to represent incoming REST request data
type CertRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// REST handler to request a certificate via the SOAP handler
func RESTReqCertHandler(w http.ResponseWriter, r *http.Request) {
	// Allow only POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON request body
	var certReq CertRequest
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Unmarshal the JSON body into the CertRequest struct
	err = json.Unmarshal(body, &certReq)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Set up form values for the SOAP handler (mimicking form submission)
	formData := make(map[string][]string)
	formData["username"] = []string{certReq.Username}
	formData["password"] = []string{certReq.Password}

	// Create a new request for the SOAP handler
	req, err := http.NewRequest(http.MethodPost, "/request-cert", bytes.NewBufferString(formEncode(formData)))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Set content type to application/x-www-form-urlencoded to mimic form data submission
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Use an in-memory ResponseRecorder to capture the SOAP handler response
	rec := httptest.NewRecorder()

	// Call the SOAP handler
	handlers.ReqCertHandler(rec, req)

	// Forward the SOAP handler's response back to the REST client
	w.Header().Set("Content-Type", rec.Header().Get("Content-Type"))
	w.WriteHeader(rec.Code)
	w.Write(rec.Body.Bytes())
}

// Helper function to encode form data as x-www-form-urlencoded
func formEncode(data map[string][]string) string {
	var buf bytes.Buffer
	for key, values := range data {
		for _, value := range values {
			buf.WriteString(key + "=" + value + "&")
		}
	}
	return buf.String()[:buf.Len()-1] // Remove trailing '&'
}
