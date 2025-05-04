# # Step 1: Build the Go application
# FROM golang:1.23 AS builder

# WORKDIR /app

# # Copy all files into the container
# COPY . .

# # Build the Go application for Linux architecture
# RUN GOOS=linux GOARCH=amd64 go build -o restapi .

# Step 2: Create the final image with a minimal base (Alpine)
FROM alpine:latest

# Install libc6-compat for Go binary compatibility
RUN apk add --no-cache libc6-compat

# Create the directory where the binary will be copied
WORKDIR /api

# Copy the Go binary from the builder image
# COPY --from=builder /app/restapi .
COPY ./restapi .

# Check if the file exists and its permissions (this is just for debugging)
RUN ls -l /api

# Ensure the Go binary has execute permissions
RUN chmod +x /api/restapi

# Command to run the application
CMD ["/api/restapi"]
