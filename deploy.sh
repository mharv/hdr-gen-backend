#!/bin/bash
sudo az login 
sudo az acr login --name hdrgencontainers

sudo docker build -t hdr-gen .
sudo docker tag hdr-gen hdrgencontainers.azurecr.io/hdr-gen:v2
sudo docker push hdrgencontainers.azurecr.io/hdr-gen:v2


# hdrcontainers.azurecr.io/hdr-gen hdrcontainers.azurecr.io/hdrgen hdrcontainers.azurecr.io/openapi hdrgencontainers.azurecr.io/hdr-gen aesopsustainability.azurecr.io/hdr-gen aesopsustainability.azurecr.io/hdr-gen