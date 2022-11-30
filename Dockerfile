FROM alpine:latest

RUN apk update && apk add curl ca-certificates

WORKDIR /
ADD install /install 
RUN /install 
CMD ["/usr/local/bin/gobackup"]