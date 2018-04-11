package cli

import (
	"testing"
)

func TestRenewToken_Success(t *testing.T) {

	var args [][]string
	for _, role := range []string{"test-master", "test-worker", "test-etcd", "test-all"} {
		args = append(args, []string{
			"renew-token",
			"--init-role=" + role,
		})
	}

	for _, arg := range args {
		RunTest(arg, 0, t)
	}
}

func TestRenewToken_Fail(t *testing.T) {

	var args [][]string
	for _, role := range []string{"test-foo", "foo", ""} {
		args = append(args, []string{
			"renew-token",
			"--init-role=" + role,
		})
	}

	for _, arg := range args {
		RunTest(arg, 1, t)
	}
}
