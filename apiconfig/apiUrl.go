package apiconfig

// Define the API URL as a constant
const SoapApiUrl = "https://167.71.219.39:8443/ejbca/ejbcaws/ejbcaws"

// Alternatively, you could define a function to return the URL
func GetSoapApiUrl() string {
	return SoapApiUrl
}
