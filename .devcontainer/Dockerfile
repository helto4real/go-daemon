FROM golang:1.12
# Avoid warnings by switching to noninteractive
ENV DEBIAN_FRONTEND=noninteractive

# Configure apt, install packages and tools
RUN apt-get update \
    && apt-get -y install --no-install-recommends apt-utils 2>&1 \
    #
    # Verify git, process tools, lsb-release (common in install instructions for CLIs) installed
    && apt-get -y install git procps lsb-release \
    #
    # Install Go tools
    && go get -u -v \
        github.com/helto4real/go-hassclient \
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
        golang.org/x/tools/cmd/gopls \
        github.com/derekparker/delve/cmd/dlv 2>&1 \
    #
    # Clean up
    && apt-get autoremove -y \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*