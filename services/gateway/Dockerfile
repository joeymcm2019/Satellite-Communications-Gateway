FROM golang:1.26-alpine AS build
WORKDIR /src
COPY . .
RUN go test ./... && CGO_ENABLED=0 go build -o /gateway ./cmd/gateway
FROM alpine:3.22
RUN addgroup -S app && adduser -S app -G app
COPY --from=build /gateway /usr/local/bin/gateway
USER app
EXPOSE 18080
CMD ["gateway"]
