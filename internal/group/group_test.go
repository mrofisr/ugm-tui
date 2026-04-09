//go:build linux || freebsd || openbsd || netbsd

package group

import (
	"os"
	"testing"
)

const _mockGroup = `test:x:0:
mock:x:65537:
`

func TestParse(t *testing.T) {
	f, _ := os.CreateTemp("", "group")
	if err := os.WriteFile(f.Name(), []byte(_mockGroup), 0o644); err != nil {
		t.Fatalf("write temp file: %s", err)
	}
	defer func() { _ = os.Remove(f.Name()) }()

	got, err := Parse(f.Name())
	if err != nil {
		t.Fatalf("Parse: %s", err)
	}

	want := []struct{ gid, name string }{
		{"0", "test"},
		{"65537", "mock"},
	}
	for i := range got {
		if got[i].Details.Gid != want[i].gid {
			t.Errorf("got GID %s, want %s", got[i].Details.Gid, want[i].gid)
		}
		if got[i].Details.Name != want[i].name {
			t.Errorf("got name %s, want %s", got[i].Details.Name, want[i].name)
		}
	}
}
