FROM golang:1.24-alpine AS builder


RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o load-balancer ./cmd/app

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

RUN adduser -D -H -h /app appuser

WORKDIR /app

COPY config/docker.yml ./config/docker.yml

ENV CONFIG_PATH=/app/config/docker.yml

COPY --from=builder /app/load-balancer .


USER appuser

EXPOSE 8080 8081

CMD ["./load-balancer"]
