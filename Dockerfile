# Stage 1: Build the Go application
FROM golang:1.23-alpine3.20 AS builder

# Set GOARCH to amd64 and GOOS to linux
ENV GOARCH=amd64
ENV GOOS=linux
ENV CGO_ENABLED=0

WORKDIR /app

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o go-es -ldflags '-w -s' .

# Stage 2: Create a lightweight production image
FROM alpine:3.20

# Create a new user named "app" with the UID and GID of 1000
RUN adduser -D -u 1000 app

WORKDIR /app

# Expose port 80
EXPOSE 80

# Copy the .env file into the container
COPY --chown=app:app config.json ./

# Copy the compiled binary from the builder stage
COPY --from=builder --chown=app:app /app/go-es .

# Define the entry point and run as the "app" user
USER app
CMD ["./go-es", "-config", "config.json"]