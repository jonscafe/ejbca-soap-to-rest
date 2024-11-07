// SOAP Client to Interact with EJBCA

package handlers

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"raweb/apiconfig"
	"time"
)

type SOAPEditUserEnvelope struct {
	XMLName      xml.Name         `xml:"soapenv:Envelope"`
	XmlnsSoapenv string           `xml:"xmlns:soapenv,attr"`
	XmlnsWs      string           `xml:"xmlns:ws,attr"`
	Header       struct{}         `xml:"soapenv:Header"`
	Body         SOAPEditUserBody `xml:"soapenv:Body"`
}

type SOAPEditUserBody struct {
	EditUser EditUserRequest `xml:"ws:editUser"`
}

type EditUserRequest struct {
	Arg0 EditUserArgs `xml:"arg0"`
}

type EditUserArgs struct {
	CaName                 string `xml:"caName"`                 // Certificate Authority Name
	CertificateProfileName string `xml:"certificateProfileName"` // Certificate Profile Name
	Email                  string `xml:"email"`                  // User's email
	EndEntityProfileName   string `xml:"endEntityProfileName"`   // End Entity Profile Name
	KeyRecoverable         bool   `xml:"keyRecoverable"`         // Whether the keys are recoverable
	Password               string `xml:"password"`               // User's password
	Status                 int    `xml:"status"`                 // User status, 10 for Active
	SubjectDN              string `xml:"subjectDN"`              // Distinguished Name (DN)
	TokenType              string `xml:"tokenType"`              // Token Type (e.g., USERGENERATED for PKCS#12)
	Username               string `xml:"username"`               // Username for the user
}

func EditUserHandler(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Extract form values
	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")
	status := 10 // Fixed value for Active status
	subjectDN := r.FormValue("subjectDN")
	// tokenType := r.FormValue("tokenType")

	// Prepare SOAP request body
	soapBody := &SOAPEditUserEnvelope{
		XmlnsSoapenv: "http://schemas.xmlsoap.org/soap/envelope/",
		XmlnsWs:      "http://ws.protocol.core.ejbca.org/",
		Body: SOAPEditUserBody{
			EditUser: EditUserRequest{
				Arg0: EditUserArgs{
					CaName:                 "RSA_subCA",
					CertificateProfileName: "RSA_enduser",
					Email:                  email,
					EndEntityProfileName:   "RSA_TTE-Doc",
					KeyRecoverable:         false, // Adjust based on requirements
					Password:               password,
					Status:                 status,
					SubjectDN:              subjectDN,
					TokenType:              "P12",
					Username:               username,
				},
			},
		},
	}

	// Marshal the SOAP request into XML
	soapRequest, err := xml.MarshalIndent(soapBody, "", "  ")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create XML: %v", err), http.StatusInternalServerError)
		return
	}

	// Log the SOAP request for debugging
	log.Printf("SOAP Request: %s", string(soapRequest))

	// Load the certificate for authentication
	cert, err := tls.LoadX509KeyPair("/etc/ssl/certs/cert.pem", "/etc/ssl/certs/cert.pem")
	if err != nil {
		log.Fatalf("Failed to load certificate: %v", err)
	}

	// Create HTTPS client with the certificate
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{cert},
				InsecureSkipVerify: true, // For testing purposes
			},
		},
		Timeout: 30 * time.Second, // Set a longer timeout
	}

	// Send SOAP request
	postUrl := apiconfig.GetSoapApiUrl()

	request, err := http.NewRequest("POST", postUrl, bytes.NewReader(soapRequest))
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	request.Header.Set("Content-Type", "text/xml")

	response, err := client.Do(request)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer response.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	body = []byte("OK")
	fmt.Fprintf(w, "%s", body)
}
