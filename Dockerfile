FROM scratch

ARG ARCH
EXPOSE 9550

COPY dist/v2ray-exporter_linux_${ARCH} /usr/bin/v2ray-exporter
ENTRYPOINT [ "/usr/bin/v2ray-exporter" ]
