FROM debian:bullseye
COPY sfptpd-exporter /usr/bin/sfptpd-exporter
ENTRYPOINT ["/usr/bin/sfptpd-exporter"]