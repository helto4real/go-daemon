FROM golang:alpine AS build

ENV ADD_ON_VERSION 0.0.10
#COPY . .
RUN apk update && \
    apk add git  && \
    go get github.com/helto4real/go-daemon/example

WORKDIR /go/src/github.com/helto4real/go-daemon/example

RUN go get -d -v ./... && \
    go install -v ./... && \
    cd /go/src/github.com/helto4real/go-daemon/example

RUN go build

FROM alpine AS runtime

RUN mkdir /daemon && \
    mkdir /daemon/config && \
    chmod 777 /daemon/config

COPY ./run.sh /daemon/run.sh
RUN  chmod 777 /daemon/run.sh

COPY --from=build /go/src/github.com/helto4real/go-daemon/example/example /daemon/go-daemon
COPY --from=build /go/src/github.com/helto4real/go-daemon/example/go-daemon.yaml /daemon/go-daemon.yaml

WORKDIR /daemon
CMD ["./run.sh"]