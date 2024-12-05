# Stage 1: Build the binary
FROM golang:1.21 AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd

# Stage 2: Create a lightweight container
FROM scratch
WORKDIR /app
COPY --from=builder /app/app /app/app
COPY config.yml /app/config.yml
EXPOSE 8081
ENTRYPOINT ["/app/app"]



# docker build -t timeline-service .
# docker run -p 8081:8080 timeline-service