FROM alpine:latest

LABEL MAINTAINER="yujizhou@sinosoft.com.cn"

WORKDIR /app/cbit-paas-gateway

COPY ./cbit-paas-gateway ./
COPY ./resource/* ./resource/
COPY ./config.yaml ./config.yaml

EXPOSE 30000
ENTRYPOINT ./cbit-paas-gateway
