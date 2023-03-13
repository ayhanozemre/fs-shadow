package filenode

import (
	"crypto/sha256"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var sum = sha256.Sum256([]byte("test"))

func TestMetaData_CreatedDateWithNano(t *testing.T) {
	createAt, err := time.Parse(time.DateTime, "2010-01-01 00:00:00")
	assert.Nil(t, err)
	metadata := MetaData{
		IsDir:      false,
		Sum:        string(sum[:]),
		Size:       1,
		CreatedAt:  createAt.Unix(),
		Permission: "",
	}
	assert.Equal(t, int64(1262304000000000000), metadata.CreatedDate(NANO).UnixNano())
	assert.Equal(t, createAt, metadata.CreatedDate(NANO))
}

func TestMetaData_CreatedDateWithMilli(t *testing.T) {
	createAt, err := time.Parse(time.DateTime, "2010-01-01 00:00:00")
	assert.Nil(t, err)
	metadata := MetaData{
		IsDir:      false,
		Sum:        string(sum[:]),
		Size:       1,
		CreatedAt:  createAt.Unix(),
		Permission: "",
	}
	assert.Equal(t, int64(1262304000), metadata.CreatedDate(MILLI).Unix())
	assert.Equal(t, createAt, metadata.CreatedDate(MILLI))
}
