package database

import (
	"testing"

	"github.com/longbridgeapp/assert"
)

func Test_PostgreSQL_prepare(t *testing.T) {
	db := &PostgreSQL{
		database: "foo",
		host:     "1.1.1.1",
		port:     "1234",
		username: "u1",
		password: "pass1",
		args:     "--foo",
	}

	err := db.prepare()
	assert.NoError(t, err)

	assert.Equal(t, db.dumpCommand, "pg_dump --host=1.1.1.1 --port=1234 --username=u1 --foo foo")
}

func Test_PostgreSQL_prepareForSocket(t *testing.T) {
	db := &PostgreSQL{
		database: "foo",
		socket:   "/var/run/postgresql/pg.5432",
		args:     "--foo",
	}

	err := db.prepare()
	assert.NoError(t, err)

	assert.Equal(t, db.dumpCommand, "pg_dump --host=/var/run/postgresql --port=5432 --foo foo")
}
