FROM golang:alpine AS build


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

COPY --from=build /go/src/github.com/helto4real/go-daemon/example /example
WORKDIR /example
CMD ["/example/example"]