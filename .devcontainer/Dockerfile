# [Choice] Go version: 1, 1.15, 1.14
ARG VARIANT=1
FROM mcr.microsoft.com/vscode/devcontainers/go:${VARIANT}

# install roamer
RUN wget https://github.com/thatoddmailbox/roamer/releases/download/v0.1.0/roamer_0.1.0_linux_amd64.tar.gz -O roamer.tar.gz
RUN tar zvxf roamer.tar.gz
RUN mv roamer /usr/local/bin/roamer
RUN rm roamer.tar.gz

RUN chmod 777 -R /go
RUN mkdir -p /go/src/github.com/myhomeworkspace

# install dependencies for the go vscode extension
RUN go get -u -v github.com/acroca/go-symbols
RUN go get -u -v github.com/cweill/gotests/...
RUN go get -u -v github.com/davidrjenni/reftools/cmd/fillstruct
RUN go get -u -v github.com/haya14busa/goplay/cmd/goplay
RUN go get -u -v github.com/mdempsky/gocode
RUN go get -u -v github.com/ramya-rao-a/go-outline
RUN go get -u -v github.com/rogpeppe/godef
RUN go get -u -v github.com/sqs/goreturns
RUN go get -u -v github.com/uudashr/gopkgs/cmd/gopkgs
RUN go get -u -v github.com/zmb3/gogetdoc
RUN go get -u -v golang.org/x/lint/golint
RUN go get -u -v golang.org/x/tools/cmd/gorename

