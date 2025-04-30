FROM golang:1.19-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY main.go ./

RUN go build -o weather-api

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/weather-api .

ENV PORT=8080
ENV METEOBLUE_API_KEY=""

EXPOSE 8080

CMD ["./weather-api"] 