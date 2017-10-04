package rrdb

import (
	"io/ioutil"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFromDirectoryFail(t *testing.T) {
	dentries, err := ioutil.ReadDir(path.Join("testdata", "fail"))
	assert.Equal(t, nil, err)
	if err != nil {
		return
	}
	for _, dentry := range dentries {
		if !dentry.IsDir() {
			continue
		}
		db, err := NewFromDirectory(path.Join("testdata", "fail", dentry.Name()))
		assert.Equal(t, (*RRDB)(nil), db, dentry.Name())
		assert.NotEqual(t, nil, err, dentry.Name())
	}
}

func TestNewFromDirectoryPass(t *testing.T) {
	dentries, err := ioutil.ReadDir(path.Join("testdata", "pass"))
	assert.Equal(t, nil, err)
	if err != nil {
		return
	}
	for _, dentry := range dentries {
		if !dentry.IsDir() {
			continue
		}
		db, err := NewFromDirectory(path.Join("testdata", "pass", dentry.Name()))
		assert.NotEqual(t, (*RRDB)(nil), db, dentry.Name())
		assert.Equal(t, nil, err, dentry.Name())
	}
}
