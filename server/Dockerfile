FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /fox/foxDenApp ./cmd/foxDenApp/main.go

FROM alpine:latest
WORKDIR /fox
COPY --from=builder /fox/foxDenApp .
EXPOSE 3001
CMD ["./foxDenApp"]