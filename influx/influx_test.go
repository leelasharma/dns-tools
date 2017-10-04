package influx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoad(t *testing.T) {
	var conf Config

	err := conf.Load("testdata/good-config.json")
	assert.Equal(t, nil, err)

	err = conf.Load("testdata/bad-config.json")
	assert.NotEqual(t, nil, err)
}
