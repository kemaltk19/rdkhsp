FROM golang:1.23-alpine AS builder
ENV GOTOOLCHAIN=auto
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o radikal-hesap main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/radikal-hesap .
# NOT: .env imaja gömülmez; ortam değişkenleri runtime'da (compose/secret) verilir.
EXPOSE 9001
CMD ["./radikal-hesap"]
