FROM golang:1.15.13-alpine3.14
WORKDIR /go/src/github.com/equestrianvault/lambda-search-go
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
WORKDIR /go
COPY --from=0 /go/src/github.com/equestrianvault/lambda-search-go/app .
EXPOSE 8080
CMD ["./app"]
