## EJBCA REST API Gateway for SOAP API
- usage: ```docker-compose up --build```
- feature: editUser SOAP API, getCertificate SOAP API

## APIs
- editUser
```
"username": "testuser3",
"password": "dummyPassword123",
"email": "testuser3@example.com",
"subjectDN": "CN=tes1",
"keyRecoverable": false,
"status": 10
```
    Response: OK
- requestCertificate
```
"username":"tes1",
"password":"password1"
```
    Response: base64 encoded of Certificate