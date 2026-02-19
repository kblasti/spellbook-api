# Build stage
FROM golang:1.24-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker's build cache
COPY go.mod go.sum ./

# Download all Go module dependencies
RUN go mod download

# Copy the entire source code into the builder stage
COPY . .

# Build the server binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /usr/local/bin/server ./cmd/server

# Runtime stage
FROM alpine:latest

# Set the working directory for the runtime
WORKDIR /app

# Copy the built server executable from the builder stage to the runtime stage
COPY --from=builder /usr/local/bin/server /usr/local/bin/server

COPY data data

# Define the PORT environment variable.
ENV PORT=8080

# Expose the port.
EXPOSE ${PORT}

# The command to run when the container starts.
CMD ["/usr/local/bin/server"]
