<p align="center">
<img src="https://user-images.githubusercontent.com/5518/205909959-12b92929-4ac5-4bb5-9111-6f9a3ed76cf6.png" width="160" />

<h1 align="center">GoBackup</h1>
<p align="center">CLI tool for backup your databases, files to FTP / SCP / S3 / GCS and other cloud storages.</p>
<p align="center">
   <a href="https://github.com/gobackup/gobackup/actions?query=workflow%3AGo"><img src="https://github.com/gobackup/gobackup/workflows/Go/badge.svg" alt="Build Status" /></a>
   <a href="https://github.com/gobackup/gobackup/releases"><img src="https://img.shields.io/github/v/release/gobackup/gobackup?label=Version&color=1" alt="GitHub release (latest by date)"></a>
   <a href="https://hub.docker.com/r/huacnlee/gobackup"><img src="https://img.shields.io/docker/v/huacnlee/gobackup?label=Docker&color=blue" alt="Docker Image Version (latest server)"></a>
   <a href="https://formulae.brew.sh/formula/gobackup"><img alt="homebrew version" src="https://img.shields.io/homebrew/v/gobackup?color=success&label=Brew"></a>
</p>

GoBackup is a fullstack backup tool design for application servers, to backup your databases, files to cloud storages (Local disk, FTP, SCP, S3, GCS, Aliyun OSS ...).

> Inspired by [backup/backup](https://github.com/backup/backup) and replace it for without Ruby dependency.

[![asciicast](https://asciinema.org/a/543564.svg)](https://asciinema.org/a/543564)

You can write a config file, run `gobackup perform` command by once to dump database as file, archive config files, and then package them into a single file.

It's allow you store the backup file to local, FTP, SCP, S3 or other cloud storages.

GoBackup 是一个类似 [backup/backup](https://github.com/backup/backup) 的一站式备份工具，为中小型服务器／个人服务器而设计，配合内置的计划任务，实现定时备份的目的。

使用 GoBackup 你可以通过一个简单的配置文件，一次（执行一个命令）将服务器上重要的（数据库、配置文件）东西导出、打包压缩，并备份到指定目的地（如：本地路径、FTP、云存储...）。

https://gobackup.github.io

## Features

- No dependencies.
- Multiple Databases source support.
- Multiple Storage type support.
- Archive paths or files into a tar.
- Split large backup file into multiple parts.
- Run as daemon to backup in schedully.
- Web UI to manage backups.

### Databases

- MySQL
- PostgreSQL
- Redis - `mode: sync/copy`
- MongoDB
- SQLite
- Microsoft SQL Server

### Storages

- Local
- FTP
- SFTP
- SCP - Upload via SSH copy
- [Amazon S3](https://aws.amazon.com/s3)
- [Aliyun OSS](https://www.aliyun.com/product/oss)
- [Google Cloud Storage](https://cloud.google.com/storage)
- [Azure Blob Storage](https://azure.microsoft.com/en-us/products/storage/blobs)
- [Backblaze B2 Cloud Storage](https://www.backblaze.com/b2)
- [Cloudflare R2](https://www.cloudflare.com/products/r2)
- [DigitalOcean Spaces](https://www.digitalocean.com/products/spaces)
- [QCloud COS](https://cloud.tencent.com/product/cos)
- [UCloud US3](https://docs.ucloud.cn/ufile/introduction/concept)
- [Qiniu Kodo](https://www.qiniu.com/products/kodo)
- [Baidu BOS](https://cloud.baidu.com/product/bos.html)
- [MinIO](https://min.io)
- [Huawei OBS](https://www.huaweicloud.com/intl/en-us/product/obs.html)
- [Volcengine TOS](https://www.volcengine.com/product/tos)
- [UpYun](https://upyun.com)
- [WebDAV](http://www.webdav.org)

## Notifier

> since: 1.5.0

Send notification when backup has success or failed.

- Mail (SMTP)
- Webhook
- Discord
- Slack
- Feishu
- DingTalk
- GitHub (Comment on Issue)
- Telegram
- AWS SES
- Postmark
- SendGrid

## Installation

```shell
curl -sSL https://gobackup.github.io/install | sh
```

after that, you will get `/usr/local/bin/gobackup` command.

### Install via Homebrew

```shell
brew install gobackup
```

```bash
$ gobackup -h
NAME:
   gobackup - Backup your databases, files to FTP / SCP / S3 / GCS and other cloud storages.

USAGE:
   gobackup [global options] command [command options] [arguments...]

VERSION:
   1.3.0

COMMANDS:
   perform
   start    Start as daemon
   run      Run GoBackup
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

## Configuration

GoBackup will seek config files in:

- ~/.gobackup/gobackup.yml
- /etc/gobackup/gobackup.yml

Example config: [gobackup_test.yml](https://github.com/huacnlee/gobackup/blob/master/gobackup_test.yml)

```yml
models:
  gitlab_app:
    databases:
      gitlab_db:
        type: postgresql
        database: gitlab_production
        username: gitlab
        password:
      gitlab_redis:
        type: redis
        mode: sync
        rdb_path: /var/db/redis/dump.rdb
        invoke_save: true
    storages:
      s3:
        type: s3
        bucket: my_app_backup
        region: us-east-1
        path: backups
        access_key_id: $S3_ACCESS_KEY_Id
        secret_access_key: $S3_SECRET_ACCESS_KEY
    compress_with:
      type: tgz
```

## Usage

### Perform backup

```bash
$ gobackup perform
```

### Backup schedule

GoBackup built in a daemon mode, you can use `gobackup start` to start it.

You can configure the `schedule` for each models, it will run backup task at the time you set.

#### For example

Configure your schedule in `gobackup.yml`

```yml
models:
  my_backup:
    schedule:
      # At 04:05 on Sunday.
      cron: "5 4 * * sun"
    storages:
      local:
        type: local
        path: /path/to/backups
    databases:
      mysql:
        type: mysql
        host: localhost
        port: 3306
        database: my_database
        username: root
        password: password
  other_backup:
    # At 04:05 on every day.
    schedule:
      every: "1day",
      at: "04:05"
    storages:
      local:
        type: local
        path: /path/to/backups
    databases:
      mysql:
        type: mysql
        host: localhost
        port: 3306
        database: my_database
        username: root
        password: password
```

### Start Daemon & Web UI

GoBackup bulit a HTTP Server for Web UI, you can start it by `gobackup start`.

It also will handle the backup schedule.

```bash
$ gobackup start

2023/03/15 23:00:30 [Config] Load config from default path.
Starting API server on port http://127.0.0.1:2703
```

> NOTE: If you wants start without daemon, use `gobackup run` instead.

Now visit http://127.0.0.1:2703 you can see the Web UI:

![gobackup-webui-main](https://user-images.githubusercontent.com/5518/225351245-90ff1eab-673a-44c7-bf37-d1964af24e12.png)
![gobackup-webui-files](https://user-images.githubusercontent.com/5518/225351184-32d9ada9-2faf-45a3-a7f3-10d41feffb8c.png)

### Signal handling

GoBackup will handle the following signals:

- `HUP` - Hot reload configuration.
- `QUIT` - Graceful shutdown.

```bash
$ ps aux | grep gobackup
jason            20443   0.0  0.1 409232800   8912   ??  Ss    7:47PM   0:00.02 gobackup run

# Reload configuration
$ kill -HUP 20443
# Exit daemon
$ kill -QUIT 20443
```

## Contributing

The [DEVELOPMENT](./DEVELOPMENT.md) document will help you to setup development environment, and guide you how to test them in local.

When you finish your work, please send a PR.

## License

MIT
