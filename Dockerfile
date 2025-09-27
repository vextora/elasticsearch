# Stage 1: Build binary
FROM golang:1.24 AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Install Air
RUN curl -sSfL https://raw.githubusercontent.com/air-verse/air/master/install.sh | sh -s -- -b /usr/local/bin

WORKDIR /app

# Copy go.mod & go.sum lalu download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy semua source code
COPY . .
COPY .air.toml ./

RUN mkdir -p /app/tmp

# Expose port jika perlu
EXPOSE 8080

# Jalankan binary
CMD ["air"]