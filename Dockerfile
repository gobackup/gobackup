FROM alpine:latest
ARG VERSION=latest
RUN apk add \
  curl \
  ca-certificates \
  openssl \
  postgresql-client \
  mysql-client \
  redis \
  mongodb-tools \
  sqlite \
  # replace busybox utils
  tar \
  gzip \
  pigz \
  bzip2 \
  # there is no pbzip2 yet
  lzip \
  xz-dev \
  lzop \
  xz \
  # pixz is in edge atm
  zstd \
  && \
  rm -rf /var/cache/apk/*

WORKDIR /tmp
RUN apk update && \
    apk add --no-cache \
    libstdc++ \
    gcompat \
    icu && \
    wget https://aka.ms/sqlpackage-linux && \
    unzip sqlpackage-linux -d /opt/sqlpackage && \
    rm sqlpackage-linux && \
    chmod +x /opt/sqlpackage/sqlpackage && \
    rm -rf /var/cache/apk/*

ENV PATH="${PATH}:/opt/sqlpackage"

ADD install /install
RUN /install ${VERSION} && rm /install

CMD ["/usr/local/bin/gobackup"]