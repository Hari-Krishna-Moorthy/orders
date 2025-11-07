# Build stage
FROM golang:1.24.10 as builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/api ./cmd/api

# Runtime stage
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/bin/api /app/api
COPY configs/config.yaml /app/config.yaml
EXPOSE 8080
ENTRYPOINT ["/app/api"]
