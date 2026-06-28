ARG GO_VERSION=1.26.3

FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/notification-api ./cmd/api

FROM alpine:3.22
WORKDIR /app

RUN addgroup -S app && adduser -S app -G app
COPY --from=build /out/notification-api /app/notification-api

USER app
EXPOSE 8487
EXPOSE 8587
ENTRYPOINT ["/app/notification-api"]
