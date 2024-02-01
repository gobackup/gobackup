## Install the MinIO Server and Client

You can use [etcd](https://etcd.io) for local development. It a distributed, reliable key-value store for the most critical data of a distributed system.
For deploying etcd you can follow the official [instruction](https://etcd.io/docs/v3.5/install/)

```bash
brew install etcd
```

Start simple etcd server:

```bash
etcd
```

And then etcd runs on port `2379`.

## Add a key-value pair

```bash
etcdctl put /gobackup/test "Hello World"
```

## Get the value

```bash
etcdctl get /gobackup/test
```

## Start GoBackup in local for etcd

```bash
GO_ENV=dev go run main.go -- perform --config ./tests/etcd.yml
```
