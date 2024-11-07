// https://www.sslshopper.com/ssl-converter.html
package main

import (
	"log"
	"net/http"
	"raweb/gateway"
	"raweb/handlers"
)

func main() {
	http.HandleFunc("/request-cert", handlers.ReqCertHandler)
	http.HandleFunc("/view-ca", handlers.ViewCAHandler)
	http.HandleFunc("/view-profile", handlers.ViewAllProfilesHandler)
	http.HandleFunc("/get-crl", handlers.GetLatestCRLHandler)
	http.HandleFunc("/api/request-cert", gateway.RESTReqCertHandler)
	http.HandleFunc("/api/edit-user", gateway.RESTeditUserHandler)
	http.HandleFunc("/api/get-crl", gateway.RESTGetCRLHandler)
	http.HandleFunc("/api/ocsp", gateway.OCSPHandler)
	// http.HandleFunc("/view-end-entity", handlers.ViewEndEntityHandler)

	log.Println("Server starting on :4444...")
	if err := http.ListenAndServe(":4444", nil); err != nil {
		log.Fatalf("Server failed: %s", err)
	}
}
