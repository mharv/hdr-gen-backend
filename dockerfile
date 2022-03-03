FROM ubuntu:20.04

RUN apt-get update -y \
  && apt-get install -y \
  && apt-get install curl -y \
  && apt-get install unzip -y \
  && curl -O -L "https://go.dev/dl/go1.17.7.linux-amd64.tar.gz" \
  && tar -C /usr/local -xzf go1.17.7.linux-amd64.tar.gz \
  && curl -O -L "https://github.com/LBNL-ETA/Radiance/releases/download/012cb178/Radiance_012cb178_Linux.zip" \
  && unzip Radiance_012cb178_Linux.zip \
  && tar -C . -xzf radiance-5.3.012cb17835-Linux.tar.gz \
  && mv radiance-5.3.012cb17835-Linux/usr/local/radiance /usr/local/ \
  && rm -rf radiance-5.3.012cb17835-Linux Radiance_012cb178_Linux.zip radiance-5.3.012cb17835-Linux.tar.gz \
  && export PATH=/usr/local/radiance/bin:$PATH


# radiance env variables
ENV RADIANCEPATH /usr/local/radiance
ENV PATH $RADIANCEPATH/bin:$PATH
ENV RAYPATH $RADIANCEPATH/lib


ENV GOROOT /usr/local/go
ENV PATH $GOROOT/bin:$PATH
ENV GOPATH /root/go
ENV APIPATH /root/go/src/api

WORKDIR $APIPATH
COPY go.mod .
COPY go.sum .
COPY main.go .
COPY scripts ./scripts

RUN chmod +x ./scripts/sleep.sh \ 
  && go mod tidy \
  && go build . \
  && echo $PATH

EXPOSE 8080
ENTRYPOINT ["./hdr-gen-backend"]