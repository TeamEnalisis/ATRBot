FROM golang:1.21-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /tradingbot .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /tradingbot /app/tradingbot
ENV PORT=8080
EXPOSE 8080
VOLUME ["/app/data"]
ENV CONFIG_PATH=/app/data/config.json
ENTRYPOINT ["/app/tradingbot"]
