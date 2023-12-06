package database

import (
	"testing"

	"github.com/gobackup/gobackup/config"
	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func TestInfluxDB2_init(t *testing.T) {
	createSubject := func(params map[string]interface{}) *InfluxDB2 {
		viper := viper.New()
		for key, value := range params {
			viper.Set(key, value)
		}
		base := newBase(
			config.ModelConfig{
				DumpPath: "/data/backups",
			},
			// Creating a new base object.
			config.SubConfig{
				Type:  "influxdb2",
				Name:  "influxdb-v2-oss",
				Viper: viper,
			},
		)

		return &InfluxDB2{
			Base: base,
		}
	}

	err1 := createSubject(map[string]interface{}{}).init()
	assert.EqualError(t, err1, "no host specified in influxdb2 configuration 'influxdb-v2-oss'")
	err2 := createSubject(map[string]interface{}{"host": "http://localhost:8086"}).init()
	assert.EqualError(t, err2, "no token specified in influxdb2 configuration 'influxdb-v2-oss'")
}

func TestInfluxDB2_influxCliArguments(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "http://localhost:8086")
	viper.Set("token", "my-token")
	viper.Set("org", "my-org")
	viper.Set("org_id", "my-org-id")
	viper.Set("bucket", "my-bucket")
	viper.Set("bucket_id", "my-bucket-id")
	viper.Set("http_debug", true)
	viper.Set("skip_verify", true)

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups",
		},
		// Creating a new base object.
		config.SubConfig{
			Type:  "influxdb2",
			Name:  "influxdb-v2-oss",
			Viper: viper,
		},
	)

	db := &InfluxDB2{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)
	assert.Equal(t, db.influxCliArguments(), []string{"backup",
		"--host=http://localhost:8086", "--token=my-token", "--bucket=my-bucket", "--bucket-id=my-bucket-id",
		"--org=my-org", "--org-id=my-org-id", "--skip-verify", "--http-debug", "/data/backups/influxdb2/influxdb-v2-oss"})
}
