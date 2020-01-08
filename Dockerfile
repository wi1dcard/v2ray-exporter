FROM alpine

ARG ARCH
EXPOSE 9550

COPY dist/v2ray-exporter_linux_${ARCH} /usr/bin/v2ray-exporter
