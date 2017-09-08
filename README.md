GoBackup
--------

A simple backup tool like [backup/backup](https://github.com/backup/backup) RubyGem.

一个类似 [backup/backup](https://github.com/backup/backup) 的站式备份工具，为中小型服务器／个人环境设计，配合 Crontab 以实现定时备份的目的。

使用 GoBackup 你可以通过一个简单的配置文件，一次（执行一个命令）将服务有关的（数据库、配置文件）导出、打包压缩，并备份到指定目的地（本地路径、FTP、云存储...）。

## Features

- No deprecations.
- Multiple Databases source support.
- Multiple Storage type support.
- Archive paths or files into a tar.

## Current Support status

### Compressor

- Tgz - `.tar.gz`

### Databases

- MySQL
- PostgreSQL
- Redis - `mode: sync/copy`

### Archive

Use `tar` command to archive many file or path into a `.tar` file.

### Storages

- Local
- FTP
- SCP - Upload via SSH copy

## Install

```bash
# for Linux
$ curl -sSL https://git.io/v5oaP | bash
```

## Usage

```bash
$ gobackup perform
2017/09/08 06:47:36 ======== ruby_china ========
2017/09/08 06:47:36 WorkDir: /tmp/gobackup/1504853256396379166
2017/09/08 06:47:36 ------------- Databases --------------
2017/09/08 06:47:36 => database | Redis: mysql
2017/09/08 06:47:36 Dump mysql dump to /tmp/gobackup/1504853256396379166/mysql/ruby-china.sql
2017/09/08 06:47:36

2017/09/08 06:47:36 => database | Redis: redis
2017/09/08 06:47:36 Copying redis dump to /tmp/gobackup/1504853256396379166/redis
2017/09/08 06:47:36
2017/09/08 06:47:36 ----------- End databases ------------

2017/09/08 06:47:36 ------------- Compressor --------------
2017/09/08 06:47:36 => Compress with Tgz...
2017/09/08 06:47:39 -> /tmp/gobackup/2017-09-08T14:47:36+08:00.tar.gz
2017/09/08 06:47:39 ----------- End Compressor ------------

2017/09/08 06:47:39 => storage | FTP
2017/09/08 06:47:39 -> Uploading...
2017/09/08 06:47:39 -> upload /ruby_china/2017-09-08T14:47:36+08:00.tar.gz
2017/09/08 06:48:04 Cleanup temp dir...
2017/09/08 06:48:04 ======= End ruby_china =======
```

