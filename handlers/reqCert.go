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
	"strings"
	"time"
)

type SOAPEnvelope struct {
	XMLName      xml.Name `xml:"soapenv:Envelope"`
	XmlnsSoapenv string   `xml:"xmlns:soapenv,attr"`
	XmlnsWs      string   `xml:"xmlns:ws,attr"`
	Header       struct{} `xml:"soapenv:Header"`
	Body         SOAPBody `xml:"soapenv:Body"`
}

type SOAPBody struct {
	Pkcs12Request Pkcs12Request `xml:"ws:pkcs12Req"`
}

type Pkcs12Request struct {
	Arg0 string `xml:"arg0"` // Username
	Arg1 string `xml:"arg1"` // Password or authentication mechanism
	Arg2 string `xml:"arg2"` // Base64-encoded PKCS#10 request (without PEM headers)
	Arg3 string `xml:"arg3"` // Certificate profile, default to ENDUSER
	Arg4 string `xml:"arg4"` // End entity, default to tes1
}

// Fault structure for parsing SOAP faults
type SOAPFault struct {
	Faultcode   string `xml:"faultcode"`
	Faultstring string `xml:"faultstring"`
	Detail      struct {
		EjbcaException struct {
			ErrorCode string `xml:"errorCode>internalErrorCode"`
		} `xml:"EjbcaException"`
	} `xml:"detail"`
}

// Function to remove specified XML tags from the response
func cleanReqCertResponse(body []byte) string {
	response := string(body)

	// Define the tags to remove
	startTag := `<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><ns2:pkcs12ReqResponse xmlns:ns2="http://ws.protocol.core.ejbca.org/"><return><type>0</type><keystoreData>`
	endTag := `</keystoreData></return></ns2:pkcs12ReqResponse></soap:Body></soap:Envelope>`

	// Remove the specified tags
	response = strings.ReplaceAll(response, startTag, "")
	response = strings.ReplaceAll(response, endTag, "")

	return response
}

func ReqCertHandler(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Prepare SOAP request body with form data and defaults
	soapBody := &SOAPEnvelope{
		XmlnsSoapenv: "http://schemas.xmlsoap.org/soap/envelope/",
		XmlnsWs:      "http://ws.protocol.core.ejbca.org/",
		Body: SOAPBody{
			Pkcs12Request: Pkcs12Request{
				Arg0: username,
				Arg1: password,
				Arg2: "RSA_TTE-Doc", // EE profile
				Arg3: "RSA_enduser", // Default certificate profile
				Arg4: "RSA",         // algo
			},
		},
	}

	// Marshal the SOAP request into XML
	soapRequest, err := xml.MarshalIndent(soapBody, "", "  ")
	if err != nil {
		log.Fatalf("Failed to create XML: %v", err)
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

	// Parse SOAP Fault if any
	var soapFault SOAPFault
	if err := xml.Unmarshal(body, &soapFault); err == nil && soapFault.Faultcode != "" {
		log.Printf("SOAP Fault: %s, ErrorCode: %s", soapFault.Faultstring, soapFault.Detail.EjbcaException.ErrorCode)
		http.Error(w, fmt.Sprintf("SOAP Fault: %s, ErrorCode: %s", soapFault.Faultstring, soapFault.Detail.EjbcaException.ErrorCode), http.StatusInternalServerError)
		return
	}

	cleanReqCertResponse := cleanReqCertResponse(body)
	// Return the response to the client
	fmt.Fprintf(w, "%s", cleanReqCertResponse)

}

// notes: hasilnya PKCS12 base64 encoded2x, decode 2x terus convert ke binary
