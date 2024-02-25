FROM golang:1.19-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o ./main.go
 
 
FROM alpine:latest AS runner
WORKDIR /app
COPY --from=builder /app/demo .
EXPOSE 1234
ENTRYPOINT ["./demo.exe", "--port", "1234"]