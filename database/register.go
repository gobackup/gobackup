package database

// RegisterAll registers all built-in database types to the DefaultRegistry
func RegisterAll() {
	Register("mysql", NewMySQL)
	Register("mariadb", NewMariaDB)
	Register("redis", NewRedis)
	Register("postgresql", NewPostgreSQL)
	Register("mongodb", NewMongoDB)
	Register("sqlite", NewSQLite)
	Register("mssql", NewMSSQL)
	Register("influxdb2", NewInfluxDB2)
	Register("etcd", NewEtcd)
	Register("firebird", NewFirebird)
}

// NewMySQL creates a new MySQL database handler
func NewMySQL(base Base) Database {
	return &MySQL{Base: base}
}

// NewMariaDB creates a new MariaDB database handler
func NewMariaDB(base Base) Database {
	return &MariaDB{Base: base}
}

// NewRedis creates a new Redis database handler
func NewRedis(base Base) Database {
	return &Redis{Base: base}
}

// NewPostgreSQL creates a new PostgreSQL database handler
func NewPostgreSQL(base Base) Database {
	return &PostgreSQL{Base: base}
}

// NewMongoDB creates a new MongoDB database handler
func NewMongoDB(base Base) Database {
	return &MongoDB{Base: base}
}

// NewSQLite creates a new SQLite database handler
func NewSQLite(base Base) Database {
	return &SQLite{Base: base}
}

// NewMSSQL creates a new MSSQL database handler
func NewMSSQL(base Base) Database {
	return &MSSQL{Base: base}
}

// NewInfluxDB2 creates a new InfluxDB2 database handler
func NewInfluxDB2(base Base) Database {
	return &InfluxDB2{Base: base}
}

// NewEtcd creates a new Etcd database handler
func NewEtcd(base Base) Database {
	return &Etcd{Base: base}
}

// NewFirebird creates a new Firebird database handler
func NewFirebird(base Base) Database {
	return &Firebird{Base: base}
}
