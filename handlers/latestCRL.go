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

// General SOAPEnvelope that can contain any SOAP body
type SOAPGetLatestCRLEnvelope struct {
	XMLName      xml.Name    `xml:"soapenv:Envelope"`
	XmlnsSoapenv string      `xml:"xmlns:soapenv,attr"`
	XmlnsWs      string      `xml:"xmlns:ws,attr"`
	Header       struct{}    `xml:"soapenv:Header"`
	Body         interface{} `xml:"soapenv:Body"`
}

// Body for the GetLatestCRL request
type SOAPGetLatestCRLBody struct {
	GetLatestCRLReq GetLatestCRLReq `xml:"ws:getLatestCRL"`
}

type GetLatestCRLReq struct {
	Arg0 string `xml:"arg0"` // CA name (optional)
	Arg1 bool   `xml:"arg1"` // Delta CRL (optional)
}

// Fault structure for parsing SOAP faults
type SOAPGetLatestCRLFault struct {
	Faultcode   string `xml:"faultcode"`
	Faultstring string `xml:"faultstring"`
	Detail      struct {
		EjbcaException struct {
			ErrorCode string `xml:"errorCode>internalErrorCode"`
		} `xml:"EjbcaException"`
	} `xml:"detail"`
}

// Function to remove specified XML tags from the response
func cleanResponse(body []byte, startTag, endTag string) string {
	response := string(body)
	// Remove the specified tags
	response = strings.ReplaceAll(response, startTag, "")
	response = strings.ReplaceAll(response, endTag, "")
	return response
}

func GetLatestCRLHandler(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get the optional form parameters
	caName := "SUBCA_FIX2"
	deltaCRL := r.FormValue("deltaCRL") == "true"

	// Prepare SOAP request body
	soapCRLBody := SOAPGetLatestCRLBody{
		GetLatestCRLReq: GetLatestCRLReq{
			Arg0: caName,
			Arg1: deltaCRL,
		},
	}

	// Create the general SOAP envelope with the specific body
	soapEnvelope := &SOAPGetLatestCRLEnvelope{
		XmlnsSoapenv: "http://schemas.xmlsoap.org/soap/envelope/",
		XmlnsWs:      "http://ws.protocol.core.ejbca.org/",
		Body:         soapCRLBody,
	}

	// Marshal the SOAP request into XML
	soapRequest, err := xml.MarshalIndent(soapEnvelope, "", "  ")
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

	// Clean the response
	startTag := `<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><ns2:getLatestCRLResponse xmlns:ns2="http://ws.protocol.core.ejbca.org/"><return>`
	endTag := `</return></ns2:getLatestCRLResponse></soap:Body></soap:Envelope>`
	cleanedResponse := cleanResponse(body, startTag, endTag)

	// Return the response to the client
	fmt.Fprintf(w, "%s", cleanedResponse)
}
