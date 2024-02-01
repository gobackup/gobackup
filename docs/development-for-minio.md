## Install the MinIO Server and Client

Use can use [MinIO](https://min.io) for local development. It is a self-hosted S3-compatible object storage server.

```bash
brew install minio/stable/minio
brew install minio/stable/mc
```

Start MinIO server:

```bash
minio server /tmp/minio
```

And then visit http://localhost:9000 to see the MinIO browser.

The Admin user:

- username: `minioadmin`
- password: `minioadmin`

## Initialize a MinIO bucket

Now we need to create a bucket for testing, we will use the following credentials:

- Bucket: `gobackup-test`
- AccessKeyId: `test-user`
- SecretAccessKey: `test-user-secret`

### Configure MinIO Client

Config MinIO Client with a default alias: `minio`

```bash
mc config host add minio http://localhost:9000 minioadmin minioadmin
```

Create a Bucket

```bash
mc mb minio/gobackup-test
```

Add Test AccessKeyId and SecretAccessKey.

With

- access_key_id: `test-user`
- secret_access_key: `test-user-secret`

```bash
mc admin user add minio test-user test-user-secret
mc admin policy attach minio readwrite --user test-user
```

## Start GoBackup in local for MinIO

```bash
GO_ENV=dev go run main.go -- perform --config ./tests/minio.yml
```
