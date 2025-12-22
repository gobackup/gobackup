FROM alpine:latest
ARG VERSION=latest
RUN apk add \
  curl \
  ca-certificates \
  openssl \
  postgresql18-client \
  mariadb-connector-c \
  mysql-client \
  mariadb-backup \
  redis \
  mongodb-tools \
  sqlite \
  # replace busybox utils
  tar \
  gzip \
  pigz \
  bzip2 \
  coreutils \
  # there is no pbzip2 yet
  lzip \
  xz-dev \
  lzop \
  xz \
  # pixz is in edge atm
  zstd \
  # microsoft sql dependencies \
  libstdc++ \
  gcompat \
  icu \
  # support change timezone
  tzdata \
  && \
  rm -rf /var/cache/apk/*

WORKDIR /tmp
RUN wget https://aka.ms/sqlpackage-linux && \
    unzip sqlpackage-linux -d /opt/sqlpackage && \
    rm sqlpackage-linux && \
    chmod +x /opt/sqlpackage/sqlpackage && \
    # https://learn.microsoft.com/en-us/sql/connect/odbc/linux-mac/installing-the-microsoft-odbc-driver-for-sql-server?view=sql-server-ver17&tabs=alpine18-install%2Calpine17-install%2Cdebian8-install%2Credhat7-13-install%2Crhel7-offline
    curl -O https://download.microsoft.com/download/fae28b9a-d880-42fd-9b98-d779f0fdd77f/msodbcsql18_18.5.1.1-1_amd64.apk && \
    curl -O https://download.microsoft.com/download/7/6/d/76de322a-d860-4894-9945-f0cc5d6a45f8/mssql-tools18_18.4.1.1-1_amd64.apk && \
    apk add --allow-untrusted msodbcsql18_18.5.1.1-1_amd64.apk && \
    apk add --allow-untrusted mssql-tools18_18.4.1.1-1_amd64.apk && \
    rm msodbcsql18_18.5.1.1-1_amd64.apk && \
    rm mssql-tools18_18.4.1.1-1_amd64.apk

ENV PATH="${PATH}:/opt/sqlpackage:/opt/mssql-tools18/bin"

# Install the influx CLI
ARG INFLUX_CLI_VERSION=2.7.5
RUN case "$(uname -m)" in \
      x86_64) arch=amd64 ;; \
      aarch64) arch=arm64 ;; \
      *) echo 'Unsupported architecture' && exit 1 ;; \
    esac && \
    curl -fLO "https://dl.influxdata.com/influxdb/releases/influxdb2-client-${INFLUX_CLI_VERSION}-linux-${arch}.tar.gz" \
         -fLO "https://dl.influxdata.com/influxdb/releases/influxdb2-client-${INFLUX_CLI_VERSION}-linux-${arch}.tar.gz.asc" && \
    tar xzf "influxdb2-client-${INFLUX_CLI_VERSION}-linux-${arch}.tar.gz" && \
    cp influx /usr/local/bin/influx && \
    rm -rf "influxdb2-client-${INFLUX_CLI_VERSION}-linux-${arch}" \
           "influxdb2-client-${INFLUX_CLI_VERSION}-linux-${arch}.tar.gz" \
           "influxdb2-client-${INFLUX_CLI_VERSION}-linux-${arch}.tar.gz.asc" \
           "influx" && \
    influx version

# Install the etcdctl
ARG ETCD_VER="v3.5.11"
RUN case "$(uname -m)" in \
      x86_64) arch=amd64 ;; \
      aarch64) arch=arm64 ;; \
      *) echo 'Unsupported architecture' && exit 1 ;; \
    esac && \
    curl -fLO https://github.com/etcd-io/etcd/releases/download/${ETCD_VER}/etcd-${ETCD_VER}-linux-${arch}.tar.gz && \
    tar xzf "etcd-${ETCD_VER}-linux-${arch}.tar.gz" && \
    cp etcd-${ETCD_VER}-linux-${arch}/etcdctl /usr/local/bin/etcdctl && \
    rm -rf "etcd-${ETCD_VER}-linux-${arch}/etcdctl" \
           "etcd-${ETCD_VER}-linux-${arch}.tar.gz" && \
    etcdctl version



ADD install /install
RUN /install ${VERSION} && rm /install

CMD ["/usr/local/bin/gobackup", "run"]
