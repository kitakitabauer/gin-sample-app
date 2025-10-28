# syntax=docker/dockerfile:1

FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o gin-sample-app .

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /app/gin-sample-app /app/gin-sample-app

EXPOSE 8080

ENV PORT=8080

CMD ["/app/gin-sample-app"]
