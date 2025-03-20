FROM golang:1.23 as build

WORKDIR /app


COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server

FROM alpine:latest

WORKDIR /

COPY --from=build /app /app

ENTRYPOINT ["/app/server"]
