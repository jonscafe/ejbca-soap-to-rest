package handlers

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

// Struct for the SOAP envelope to get available CAs
type SOAPEnvelopeCA struct {
	XMLName      xml.Name   `xml:"soapenv:Envelope"`
	XmlnsSoapenv string     `xml:"xmlns:soapenv,attr"`
	XmlnsWs      string     `xml:"xmlns:ws,attr"`
	Header       struct{}   `xml:"soapenv:Header"`
	Body         SOAPBodyCA `xml:"soapenv:Body"`
}

type SOAPBodyCA struct {
	GetAvailableCAs struct{} `xml:"ws:getAvailableCAs"`
}

func ViewCAHandler(w http.ResponseWriter, r *http.Request) {
	// Prepare SOAP request body
	soapBody := &SOAPEnvelopeCA{
		XmlnsSoapenv: "http://schemas.xmlsoap.org/soap/envelope/",
		XmlnsWs:      "http://ws.protocol.core.ejbca.org/",
	}

	// Marshal the SOAP request into XML
	soapRequest, err := xml.MarshalIndent(soapBody, "", "  ")
	if err != nil {
		log.Fatalf("Failed to create XML: %v", err)
	}

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
				InsecureSkipVerify: true, // for testing purposes
			},
		},
	}

	// Send SOAP request
	request, err := http.NewRequest("POST", "https://192.168.103.219:8443/ejbca/ejbcaws/ejbcaws", bytes.NewReader(soapRequest))
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	request.Header.Set("Content-Type", "text/xml")

	response, err := client.Do(request)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer response.Body.Close()

	// Read the response
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	// Render the response in an HTML template
	tmpl, err := template.ParseFiles("static/view_ca.html")
	if err != nil {
		log.Fatalf("Failed to load template: %v", err)
	}

	data := struct {
		CAList string
	}{
		CAList: string(body), // Directly putting the response body into the CAList field
	}

	// Execute the template with data
	tmpl.Execute(w, data)
}
