package file

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExists(t *testing.T) {

	file, err := ioutil.TempFile("", "gosba")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	path, err := filepath.Abs(filepath.Dir(file.Name()))
	if err != nil {
		log.Fatal(err)
	}

	assert.True(t, Exists(path))
	assert.False(t, Exists(path+"false"))
}
