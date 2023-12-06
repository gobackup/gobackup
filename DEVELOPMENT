We are welcome to any contributions. Please read the following guides before you start.

This document describes how to develop GoBackup.

- gobackup_test.yml is a test configuration file for development.
- tests/ folder contains some test for special cases.

## Documentions

- [Release new version](./docs/release-new-version.md)

## Run tests

We have a `Makefile` provided some commands:

- `make test` run unit test.
- `make test:all` run all tests, this will execute test for `tests/` folder.

## Guide for adding a new storage

- Add a new storage in `storages/` folder, you can follow the `storages/gcs.go` as an example.
  - If the storage is a AWS S3 compatible storage, you can use `storages/s3.go` as a base.
- Add a new test file in `tests/` folder, you can follow the `tests/oss.yml` as an example.
  - Don't put your credentials in the test file, you can use environment variables to pass them.

## Storages development guides

- [MinIO](./docs/development-for-minio.md)

## Release a new version

1. Create a tag named `vX.Y.Z` and push tag, then GitHub Actions will release it.
2. Write a release note for descrbe this version to [GitHub Releases](https://github.com/gobackup/gobackup/releases).
3. Makesure to update the [Website](https://github.com/gobackup/gobackup.github.io) doc.
