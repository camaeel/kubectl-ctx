package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateKubeconfig(t *testing.T) {
	file := CreateKubeconfig(t, "dev", map[string]string{"dev": "", "prod": ""})
	assert.FileExists(t, file)
}
