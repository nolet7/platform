FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod first
COPY go.mod ./

# Download dependencies (creates go.sum)
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o catalog ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/catalog .

EXPOSE 8080

CMD ["./catalog"]
