FROM alpine:latest

RUN apk update && apk add curl ca-certificates \
  postgresql-client mysql-client redis mongodb-tools pigz openssl && \
  rm -rf /var/cache/apk/*

WORKDIR /
ADD install /install 
RUN /install 
CMD ["/usr/local/bin/gobackup"]