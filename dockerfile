FROM ubuntu:20.04

RUN apt-get update -y \
  && apt-get install -y \
  && apt-get install curl -y \
  && curl -O -L "https://go.dev/dl/go1.17.7.linux-amd64.tar.gz" \
  && tar -C /usr/local -xzf go1.17.7.linux-amd64.tar.gz 


ENV GOROOT /usr/local/go
ENV PATH $GOROOT/bin:$PATH
ENV GOPATH /root/go
ENV APIPATH /root/go/src/api

WORKDIR $APIPATH
COPY go.mod .
COPY go.sum .
COPY main.go .

RUN \ 
  go mod tidy \
  && go build .

EXPOSE 8080
ENTRYPOINT ["./hdr-gen-backend"]