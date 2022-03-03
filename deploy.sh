#!/bin/bash
az login 
az acr login --name hdrgencontainers

docker build -t hdr-gen .
docker tag hdr-gen hdrgencontainers.azurecr.io/hdr-gen:v2
docker push hdrgencontainers.azurecr.io/hdr-gen:v2


# hdrcontainers.azurecr.io/hdr-gen hdrcontainers.azurecr.io/hdrgen hdrcontainers.azurecr.io/openapi hdrgencontainers.azurecr.io/hdr-gen aesopsustainability.azurecr.io/hdr-gen aesopsustainability.azurecr.io/hdr-gen