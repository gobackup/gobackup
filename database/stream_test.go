package database

import (
	"io"
	"testing"

	"github.com/longbridgeapp/assert"
)

func Test_Stream_ok(t *testing.T) {
	s, err := startStream("echo hello", "a.sql", ".sql")
	assert.NoError(t, err)
	out, err := io.ReadAll(s)
	assert.NoError(t, err)
	assert.Equal(t, "hello\n", string(out))
	assert.NoError(t, s.Close())
}

func Test_Stream_failedCommandIsNotEOF(t *testing.T) {
	s, err := startStream("sh -c exit_1_not_a_cmd", "a.sql", ".sql")
	assert.NoError(t, err)
	_, err = io.ReadAll(s)
	assert.NotNil(t, err)
}

func Test_Stream_missingCommand(t *testing.T) {
	_, err := startStream("no-such-dump-cmd", "a.sql", ".sql")
	assert.NotNil(t, err)
}

func Test_MySQL_build_unchanged(t *testing.T) {
	db := &MySQL{database: "mydb", username: "root", host: "h", port: "3306"}
	db.dumpPath = "/d"
	assert.Equal(t, "mysqldump --host h --port 3306 -u root mydb --result-file=/d/mydb.sql", db.build())
	assert.Equal(t, "mysqldump --host h --port 3306 -u root mydb", "mysqldump "+join(db.dumpArgs()))
}

func join(a []string) string {
	s := ""
	for i, v := range a {
		if i > 0 {
			s += " "
		}
		s += v
	}
	return s
}

func Test_MongoDB_build_unchanged_and_stream_command(t *testing.T) {
	db := &MongoDB{database: "app", host: "h", port: "27017", username: "u", password: "p", oplog: true}
	db.dumpPath = "/d"
	assert.Equal(t, "mongodump --db=app --username=u --password=p --host=h --port=27017 --oplog --out=/d", db.build())
	assert.Equal(t, "mongodump --db=app --username=u --password=p --host=h --port=27017 --oplog", db.dumpCommand())
}

func Test_PostgreSQL_build_unchanged_and_stream_command(t *testing.T) {
	db := &PostgreSQL{database: "app", host: "h", port: "5432", username: "u", tables: []string{"a", "b"}}
	db._dumpFilePath = "/d/app.sql"
	assert.Equal(t, "pg_dump --host=h --port=5432 --username=u --table=a --table=b app -f /d/app.sql", db.build())
	assert.Equal(t, "pg_dump --host=h --port=5432 --username=u --table=a --table=b app", db.dumpCommand())

	all := &PostgreSQL{allDatabases: true, host: "h", port: "5432", args: "--clean"}
	all._dumpFilePath = "/d/all_databases.sql"
	assert.Equal(t, "pg_dumpall --host=h --port=5432 --clean > /d/all_databases.sql", all.build())
	assert.Equal(t, "pg_dumpall --host=h --port=5432 --clean", all.dumpCommand())

	custom := &PostgreSQL{database: "app", compress: "gzip:6"}
	custom._dumpFilePath = "/d/app.sql.gz"
	assert.Equal(t, "pg_dump --compress=gzip:6 --format=custom app -f /d/app.sql.gz", custom.build())
}
