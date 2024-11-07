package gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"raweb/handlers"
)

// Struct to represent incoming REST request data for editing a user
type EditUserRequest struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	Email          string `json:"email"`
	SubjectDN      string `json:"subjectDN"`
	TokenType      string `json:"tokenType"`
	KeyRecoverable bool   `json:"keyRecoverable"`
	Status         int    `json:"status"`
}

// REST handler to edit a user via the SOAP handler
func RESTeditUserHandler(w http.ResponseWriter, r *http.Request) {
	// Allow only POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON request body
	var editReq EditUserRequest
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Unmarshal the JSON body into the EditUserRequest struct
	err = json.Unmarshal(body, &editReq)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Set up form values for the SOAP handler (mimicking form submission)
	formData := make(map[string][]string)
	formData["username"] = []string{editReq.Username}
	formData["password"] = []string{editReq.Password}
	formData["email"] = []string{editReq.Email}
	formData["subjectDN"] = []string{editReq.SubjectDN}
	// formData["tokenType"] = []string{editReq.TokenType}
	formData["keyRecoverable"] = []string{boolToString(editReq.KeyRecoverable)}
	formData["status"] = []string{intToString(editReq.Status)}

	// Create a new request for the SOAP handler
	req, err := http.NewRequest(http.MethodPost, "/edit-user", bytes.NewBufferString(formEncode(formData)))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Set content type to application/x-www-form-urlencoded to mimic form data submission
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Use an in-memory ResponseRecorder to capture the SOAP handler response
	rec := httptest.NewRecorder()

	// Call the SOAP handler
	handlers.EditUserHandler(rec, req) // Assuming you have an EditUserHandler in your SOAP handlers

	// Forward the SOAP handler's response back to the REST client
	w.Header().Set("Content-Type", rec.Header().Get("Content-Type"))
	w.WriteHeader(rec.Code)
	w.Write(rec.Body.Bytes())
}

// Helper function to encode form data as x-www-form-urlencoded
// This function is assumed to be declared elsewhere in the package
// func formEncode(data map[string][]string) string {
// 	var buf bytes.Buffer
// 	for key, values := range data {
// 		for _, value := range values {
// 			buf.WriteString(key + "=" + value + "&")
// 		}
// 	}
// 	return buf.String()[:buf.Len()-1] // Remove trailing '&'
// }

// Helper function to convert bool to string for form submission
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// Helper function to convert int to string
func intToString(i int) string {
	return fmt.Sprintf("%d", i)
}
