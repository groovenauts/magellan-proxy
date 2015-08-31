FROM busybox:ubuntu-14.04
MAINTAINER magellan@groovenauts.jp

COPY magellan-proxy_linux_amd64 /usr/app/magellan-proxy

CMD ["/usr/app/magellan-proxy", "sleep", "3600"]
