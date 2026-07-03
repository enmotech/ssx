package entry

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	"github.com/vimiix/ssx/internal/utils"
)

func TestEntry_String(t *testing.T) {
	e := Entry{Host: "host", Port: "22", User: "user"}
	assert.Equal(t, "user@host:22", e.String())
}

func TestEntry_Tidy(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	// Tidy only sets KeyPath to defaultIdentityFile when utils.FileExists
	// reports it present. Stub the check so the test does not depend on the
	// runner having a real ~/.ssh/id_rsa.
	patches.ApplyFunc(utils.FileExists, func(filename string) bool {
		return filename == defaultIdentityFile
	})

	e := &Entry{}
	if err := e.Tidy(); err != nil {
		t.Fatalf("Received unexpected error:\n%+v", err)
	}

	assert.Equal(t, "root", e.User)
	assert.Equal(t, "22", e.Port)
	assert.Equal(t, defaultIdentityFile, e.KeyPath)
}
