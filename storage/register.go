package storage

// RegisterAll registers all built-in storage types to the DefaultRegistry
func RegisterAll() {
	Register("local", NewLocal)
	Register("webdav", NewWebDAV)
	Register("ftp", NewFTP)
	Register("scp", NewSCP)
	Register("sftp", NewSFTP)
	Register("gcs", NewGCS)
	Register("azure", NewAzure)

	// S3-compatible storages
	Register("s3", NewS3("s3"))
	Register("oss", NewS3("oss"))
	Register("minio", NewS3("minio"))
	Register("b2", NewS3("b2"))
	Register("us3", NewS3("us3"))
	Register("cos", NewS3("cos"))
	Register("kodo", NewS3("kodo"))
	Register("r2", NewS3("r2"))
	Register("spaces", NewS3("spaces"))
	Register("bos", NewS3("bos"))
	Register("obs", NewS3("obs"))
	Register("tos", NewS3("tos"))
	Register("upyun", NewS3("upyun"))
}

// NewLocal creates a new Local storage handler
func NewLocal(base Base) Storage {
	return &Local{Base: base}
}

// NewWebDAV creates a new WebDAV storage handler
func NewWebDAV(base Base) Storage {
	return &WebDAV{Base: base}
}

// NewFTP creates a new FTP storage handler
func NewFTP(base Base) Storage {
	return &FTP{Base: base}
}

// NewSCP creates a new SCP storage handler
func NewSCP(base Base) Storage {
	return &SCP{Base: base}
}

// NewSFTP creates a new SFTP storage handler
func NewSFTP(base Base) Storage {
	return &SFTP{Base: base}
}

// NewGCS creates a new GCS storage handler
func NewGCS(base Base) Storage {
	return &GCS{Base: base}
}

// NewAzure creates a new Azure storage handler
func NewAzure(base Base) Storage {
	return &Azure{Base: base}
}

// NewS3 returns a factory function for S3-compatible storage with the given service name
func NewS3(service string) Factory {
	return func(base Base) Storage {
		return &S3{Base: base, Service: service}
	}
}
