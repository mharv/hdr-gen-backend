


# # LABEL for the custom image
# LABEL maintainer="mitchell.harvey@arup.com"
# LABEL version="0.1"
# LABEL description="This is a custom container to run a webserver that exposes radiance cli commands as a service"

# RUN apt-get update \ 
#      && apt-get install -y --no-install-recommends dialog \
#      && apt-get update \
#      && apt-get install -y --no-install-recommends openssh-server \
#      && echo "root:Docker!" | chpasswd 

# # Copy the sshd_config file to the /etc/ssh/ directory
# COPY sshd_config /etc/ssh/

# # Copy and configure the ssh_setup file
# RUN mkdir -p /tmp
# COPY ssh_setup.sh /tmp
# RUN chmod +x /tmp/ssh_setup.sh \
#     && (sleep 1;/tmp/ssh_setup.sh 2>&1 > /dev/null)

# # Open port 2222 for SSH access
# EXPOSE 80 2222

# ENTRYPOINT ["/bin/bash", "-c", "touch /storage/test.txt"]


# CMD ./bin/start.sh

# FROM golang:1.17-alpine AS build
# WORKDIR /go/src

# WORKDIR /go/src

# COPY main.go ./
# COPY go.mod ./
# COPY go.sum ./

# RUN go mod tidy
# ENV CGO_ENABLED=0
# RUN go build .

# FROM scratch AS runtime
# ENV GIN_MODE=release
# COPY --from=build /go/src/hdr-gen-backend ./
# # COPY --from=build /go/src/.env ./
# EXPOSE 8080
# ENTRYPOINT ["./hdr-gen-backend"]

# FROM ubuntu:18.04

# FROM ubuntu:18.04

# RUN apt-get update -y \
#   && apt-get install -y \
#   && apt-get install curl -y \
#   && curl -O -L "https://go.dev/dl/go1.17.7.linux-amd64.tar.gz" \
#   && tar -C /usr/local -xzf go1.17.7.linux-amd64.tar.gz 


# ENV GOROOT /usr/local/go
# ENV PATH $GOROOT/bin:$PATH
# ENV GOPATH /root/go
# ENV APIPATH /root/go/src/api

# WORKDIR $APIPATH
# COPY go.mod .
# COPY go.sum .
# COPY main.go .

# RUN \ 
#   go mod tidy \
#   && go build .

# EXPOSE 80
# ENTRYPOINT ["./hdr-gen-backend"]


FROM alpine:3.8

RUN apk add --no-cache \
  ca-certificates \
	git \
	gcc \
	musl-dev \
	openssl \
   curl 

RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

RUN curl -O -L "https://go.dev/dl/go1.17.7.linux-amd64.tar.gz" \
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

EXPOSE 80
ENTRYPOINT ["./hdr-gen-backend"]