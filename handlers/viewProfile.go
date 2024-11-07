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

// Struct for the SOAP envelope to get profiles
type SOAPEnvelopeProfile struct {
	XMLName      xml.Name        `xml:"soapenv:Envelope"`
	XmlnsSoapenv string          `xml:"xmlns:soapenv,attr"`
	XmlnsWs      string          `xml:"xmlns:ws,attr"`
	Header       struct{}        `xml:"soapenv:Header"`
	Body         SOAPBodyProfile `xml:"soapenv:Body"`
}

type SOAPBodyProfile struct {
	GetProfile GetProfileRequest `xml:"ws:getProfile"`
}

type GetProfileRequest struct {
	Arg0 string `xml:"arg0"` // Profile Type (eep/cp)
	Arg1 string `xml:"arg1"` // Optional (leave empty if not needed)
}

func ViewAllProfilesHandler(w http.ResponseWriter, r *http.Request) {
	// Prepare to fetch both End Entity Profiles (EEP) and Certificate Profiles (CP)
	profileTypes := []string{"eep", "cp"}
	var allProfiles []byte

	for _, profileType := range profileTypes {
		// Prepare SOAP request body for each profile type
		soapBody := &SOAPEnvelopeProfile{
			XmlnsSoapenv: "http://schemas.xmlsoap.org/soap/envelope/",
			XmlnsWs:      "http://ws.protocol.core.ejbca.org/",
			Body: SOAPBodyProfile{
				GetProfile: GetProfileRequest{
					Arg0: profileType, // Correctly use "eep" for End Entity Profiles and "cp" for Certificate Profiles
					Arg1: "",          // Optional, leave empty if not needed
				},
			},
		}

		// Marshal the SOAP request into XML
		soapRequest, err := xml.MarshalIndent(soapBody, "", "  ")
		if err != nil {
			log.Printf("Failed to create XML for profile type %s: %v", profileType, err)
			continue
		}

		// Log SOAP Request Body for debugging
		log.Printf("SOAP Request for profile type %s: %s\n", profileType, string(soapRequest))

		// Load the certificate for authentication
		cert, err := tls.LoadX509KeyPair("/etc/ssl/certs/cert.pem", "/etc/ssl/certs/cert.pem")
		if err != nil {
			log.Printf("Failed to load certificate: %v", err)
			continue
		}

		// Create HTTPS client with the certificate
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					Certificates:       []tls.Certificate{cert},
					InsecureSkipVerify: true, // For testing purposes
				},
			},
		}

		// Send SOAP request
		request, err := http.NewRequest("POST", "https://192.168.103.219:8443/ejbca/ejbcaws/ejbcaws", bytes.NewReader(soapRequest))
		if err != nil {
			log.Printf("Failed to create request for profile type %s: %v", profileType, err)
			continue
		}

		request.Header.Set("Content-Type", "text/xml")

		response, err := client.Do(request)
		if err != nil {
			log.Printf("Failed to send request for profile type %s: %v", profileType, err)
			continue
		}
		defer response.Body.Close()

		// Read the response
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Printf("Failed to read response for profile type %s: %v", profileType, err)
			continue
		}

		// Log the response for debugging
		log.Printf("Response for profile type %s: %s", profileType, string(body))

		// Append the result to allProfiles
		allProfiles = append(allProfiles, body...)
	}

	// Check if profiles were fetched
	if len(allProfiles) == 0 {
		http.Error(w, "No profiles could be fetched", http.StatusInternalServerError)
		return
	}

	// Render the combined response in an HTML template
	tmpl, err := template.ParseFiles("static/view_profile.html")
	if err != nil {
		log.Fatalf("Failed to load template: %v", err)
	}

	data := struct {
		ProfileList string
	}{
		ProfileList: string(allProfiles), // Display all profile responses
	}

	// Execute the template with data
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Failed to execute template: %v", err)
		http.Error(w, "Failed to display profiles", http.StatusInternalServerError)
	}
}
