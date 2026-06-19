ARG GO_VERSION=1.26.3

FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/task-tracker-api ./cmd/api

FROM alpine:3.22
WORKDIR /app

RUN addgroup -S app && adduser -S app -G app
COPY --from=build /out/task-tracker-api /app/task-tracker-api

USER app
EXPOSE 8485
ENTRYPOINT ["/app/task-tracker-api"]
