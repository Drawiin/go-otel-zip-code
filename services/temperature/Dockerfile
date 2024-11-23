# Build stage
FROM golang:latest AS builder
LABEL authors="drawiin"
WORKDIR /app
COPY . .
# Set GOOS to linux to build a binary for linux
# Disable CGO for static linking
# Add ldflags to reduce binary size
# -o flag to specify the output binary name
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o runner ./cmd

# Stage to get CA certificates
FROM alpine:latest AS certs
RUN apk --no-cache add ca-certificates

# Final stage
FROM scratch
COPY --from=builder /app/runner .
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Use ENTRYPOINT to allow passing arguments to the binary
ENTRYPOINT ["./runner"]