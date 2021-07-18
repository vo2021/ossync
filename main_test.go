package main


import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_get_metadata_filename(t *testing.T) {
	f := get_metadata_filename("tmp", "hello")
	actual := get_config_folder()+"/tmp-hello.json"
	assert.Equal(t, f, actual)

	f = get_metadata_filename("/tmp/", "hello")
	actual = get_config_folder()+"/tmp-hello.json"
	assert.Equal(t, f, actual)

	f = get_metadata_filename("~/tmp/", "hello")
	actual = get_config_folder()+"/tmp-hello.json"
	assert.Equal(t, f, actual)
}
