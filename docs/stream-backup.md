# Stream backup (MySQL / MongoDB / PostgreSQL -> S3)

With `stream: true` the dump is piped straight from the dump tool to the storage:
no dump directory and no `.tar`/`.tar.gz` archive is written to disk.

```yml
models:
  mysql_stream:
    stream: true
    compress_with:
      type: gz # optional, compresses on the fly. Without it the dump is uploaded as is.
    databases:
      app:
        type: mysql # or mongodb, postgresql
        host: 127.0.0.1
        database: app
        username: root
        password: secret
    storages:
      s3:
        type: s3
        bucket: my-bucket
        path: backups
        keep: 10
        part_size: 64 # MiB, optional. Max object size = part_size * 10000
```

## Uploaded files and restore

| database     | dump tool                      | file (`compress_with: gz`)      | restore                                              |
| ------------ | ------------------------------ | ------------------------------- | ---------------------------------------------------- |
| `mysql`      | `mysqldump`                    | `<time>.sql` (`.sql.gz`)        | `gunzip -c f.sql.gz \| mysql`                        |
| `mongodb`    | `mongodump --archive`          | `<time>.archive` (`.archive.gz`)| `gunzip -c f.archive.gz \| mongorestore --archive`   |
| `postgresql` | `pg_dump` / `pg_dumpall`       | `<time>.sql` (`.sql.gz`)        | `gunzip -c f.sql.gz \| psql`                         |

The file is a plain gzip stream, not a tar archive: use `gunzip`, not `tar -xzf`.

## Restrictions

They are validated when the config is loaded:

- exactly one database, of type `mysql`, `mongodb` or `postgresql`;
- exactly one storage, any S3-compatible type (`s3`, `minio`, `r2`, `b2`, ...);
- `archive`, `encrypt_with` and `split_with` are not supported;
- `compress_with.type` can only be `gz` (the stream is a single file, so there is no tar);
- for `postgresql`, the database `compress` option (pg_dump custom format) can't be combined with `compress_with`.

If the dump tool fails, the multipart upload is aborted: no partial object is left in the bucket.
Models without `stream: true` behave exactly as before.
