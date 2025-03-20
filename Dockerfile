FROM golang:1.23 as build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o app

FROM alpine:latest

WORKDIR /

COPY --from=build /app /app

RUN chmod +x /app

ENTRYPOINT ["/app"]
