#!/bin/bash

docker build -t hdr-gen .
docker run -p 8080:8080 hdr-gen
