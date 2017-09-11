package database

import (
	"fmt"
	"github.com/huacnlee/gobackup/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Monkey struct {
}

func (ctx Monkey) perform(model config.ModelConfig, dbConfig config.SubConfig) error {
	if model.Name != "TestMonkey" {
		return fmt.Errorf("Error")
	}
	if dbConfig.Name != "mysql1" {
		return fmt.Errorf("Error")
	}
	return nil
}

func TestBaseInterface(t *testing.T) {
	var ctx Base
	ctx = Monkey{}
	model := config.ModelConfig{
		Name: "TestMonkey",
	}
	dbConfig := config.SubConfig{
		Name: "mysql1",
	}
	err := ctx.perform(model, dbConfig)
	assert.Nil(t, err)
}
