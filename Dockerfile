FROM golang:1.22-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY internal ./internal
COPY cmd ./cmd

RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.20

RUN adduser -D -H app
WORKDIR /app

COPY --from=build /server /app/server
COPY data /app/data

USER app

ENV COUPON_DATA_DIR=/app/data

EXPOSE 8080

ENTRYPOINT ["/app/server"]
