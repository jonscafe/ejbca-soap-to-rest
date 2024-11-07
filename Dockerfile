# Use Go base image
FROM golang:1.20-alpine

# Install necessary packages
RUN apk add --no-cache ca-certificates openssl openjdk11

# Copy the certificate for API access
# use your cert.pem from your EJBCA Admin Certificate
COPY certs/cert.pem /etc/ssl/certs/cert.pem

# Set the working directory
WORKDIR /app

# Copy all source files to the container
COPY . .

# Download Go dependencies
RUN go mod download

# Build the Go application
RUN go build -o main .

# Expose port 4444 for the web application
EXPOSE 4444

# Start the Go application
CMD ["./main"]
