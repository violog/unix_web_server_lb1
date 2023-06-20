# go get lib from private repo and stop using image to not share priv key
FROM golang:1.18.2-alpine as builder

# install everything 'go build' needs
RUN apk add --no-cache make gcc musl-dev linux-headers git openssh

# set GOPATH and WORKDIR so go build (prior to go1.13) can build correctly
ENV GOPATH=/go
WORKDIR /go/src/github.com/violog/unix_web_server_lb1

# copy all the files to the image
ADD . .

# build the application
RUN go build -mod vendor -o main ./cmd/main.go

# use a minimal alpine image
FROM alpine:3.8

# set working directory
WORKDIR /root

# copy the binary from builder
COPY --from=builder /go/src/github.com/violog/unix_web_server_lb1 .
EXPOSE 80

ENTRYPOINT ["sh", "-c", "./main"]
