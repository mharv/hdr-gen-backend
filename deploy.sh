#!/bin/bash
sudo az login 
sudo az acr login --name hdrcontainers

sudo docker build -t hdr-gen .
sudo docker tag openapi hdrcontainers.azurecr.io/hdr-gen:v1
sudo docker push hdrcontainers.azurecr.io/hdr-gen:v1


