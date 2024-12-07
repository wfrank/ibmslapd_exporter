ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest
LABEL maintainer="Frank Wang <gladandong@gmail.com>"

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/ibmslapd_exporter /bin/ibmslapd_exporter

USER        nobody
EXPOSE      9981
ENTRYPOINT  [ "/bin/ibmslapd_exporter" ]
