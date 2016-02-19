FROM ubuntu:latest
MAINTAINER secretsquirrel@gtmtech.co.uk

# Install base utils
RUN apt-get update
RUN apt-get install -y git bzr build-essential
RUN mkdir -p /tmp/go/ /tmp/gocode/
ADD https://storage.googleapis.com/golang/go1.5.3.linux-amd64.tar.gz /tmp/go1.5.3.linux-amd64.tar.gz
RUN tar -C /tmp/ -xzvf /tmp/go1.5.3.linux-amd64.tar.gz && rm -f /tmp/go1.5.3.linux-amd64.tar.gz
ENV GOROOT=/tmp/go
ENV GOPATH=/tmp/gocode

# Install go dependencies
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/tmp/go/bin
RUN go get -d -u github.com/mitchellh/goamz/aws
RUN go get -d -u github.com/mitchellh/goamz/s3
RUN cd /tmp/gocode/src/github.com/mitchellh/goamz/aws && git checkout -q caaaea8b30ee15616494ee68abd5d8ebbbef05cf
RUN cd /tmp/gocode/src/github.com/mitchellh/goamz/s3 && git checkout -q caaaea8b30ee15616494ee68abd5d8ebbbef05cf

# Setup build environment
ENV CGO_ENABLED=0
ADD secret_squirrel.go /tmp/secret_squirrel.go
ADD secret_squirrel_s3.go /tmp/secret_squirrel_s3.go

# Build executable
RUN cd /tmp && go build -a -installsuffix nocgo --ldflags '-extldflags "--static"' /tmp/secret_squirrel.go
RUN cd /tmp && go build -a -installsuffix nocgo --ldflags '-extldflags "--static"' /tmp/secret_squirrel_s3.go


