FROM ruby:2.2
MAINTAINER magellan@groovenauts.jp

COPY magellan-proxy_linux_amd64 /usr/app/magellan-proxy

CMD ["/usr/app/magellan-proxy", "ruby", "-run", "-e", "httpd", ".", "-p", "80"]
