//go:build linux || freebsd || openbsd || netbsd

package passwd

import (
	"os/user"
	"strings"
	"testing"
)

const _mockPwd = `test:x:0:0:test:/test:/bin/test
mock:x:1:65537:mock:/mock:/sbin/mock
`

func TestParse(t *testing.T) {
	got, err := Parse(strings.NewReader(_mockPwd))
	if err != nil {
		t.Fatalf("Parse: %s", err)
	}

	want := getWanted()
	for i := range got {
		if got[i].Details.Uid != want[i].Details.Uid {
			t.Errorf("got UID %s, want %s", got[i].Details.Uid, want[i].Details.Uid)
		}
		if got[i].Details.Gid != want[i].Details.Gid {
			t.Errorf("got GID %s, want %s", got[i].Details.Gid, want[i].Details.Gid)
		}
		if got[i].Details.Username != want[i].Details.Username {
			t.Errorf("got username %s, want %s", got[i].Details.Username, want[i].Details.Username)
		}
		if got[i].Details.Name != want[i].Details.Name {
			t.Errorf("got name %s, want %s", got[i].Details.Name, want[i].Details.Name)
		}
		if got[i].Details.HomeDir != want[i].Details.HomeDir {
			t.Errorf("got homedir %s, want %s", got[i].Details.HomeDir, want[i].Details.HomeDir)
		}
		if got[i].Shell != want[i].Shell {
			t.Errorf("got shell %s, want %s", got[i].Shell, want[i].Shell)
		}
	}
}

func getWanted() []User {
	rootGroup, _ := user.LookupGroupId("0")

	return []User{
		{
			Details: user.User{Uid: "0", Gid: "0", Username: "test", Name: "test", HomeDir: "/test"},
			Shell:   "/bin/test",
			Groups:  []*user.Group{rootGroup},
		},
		{
			Details: user.User{Uid: "1", Gid: "65537", Username: "mock", Name: "mock", HomeDir: "/mock"},
			Shell:   "/sbin/mock",
			Groups:  []*user.Group{nil},
		},
	}
}
