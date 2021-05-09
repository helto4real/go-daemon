FROM golang:1.14-alpine
# Avoid warnings by switching to noninteractive
ENV DEBIAN_FRONTEND=noninteractive

# Configure apt, install packages and tools
RUN apk update \
    #
    # Verify git, process tools, lsb-release (common in install instructions for CLIs) installed
    && apk add git procps g++ \
    #
    # Install Go tools
    && go get -u -v \
    github.com/mdempsky/gocode \
    github.com/uudashr/gopkgs/cmd/gopkgs \
    github.com/ramya-rao-a/go-outline \
    github.com/acroca/go-symbols \
    golang.org/x/tools/cmd/guru  \
    golang.org/x/tools/cmd/gorename  \
    github.com/stamblerre/gocode  \
    github.com/rogpeppe/godef  \
    github.com/sqs/goreturns  \
    golang.org/x/lint/golint  \
    golang.org/x/tools/gopls \
    github.com/derekparker/delve/cmd/dlv 2>&1
