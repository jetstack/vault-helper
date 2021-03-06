// Copyright Jetstack Ltd. See LICENSE for details.
package cli

import (
	"testing"
)

func TestRenewToken_Success(t *testing.T) {

	dir, err := TmpDir()
	if err != nil {
		t.Errorf("unexpected error getting tmp dir: %v", err)
		return
	}

	var args [][]string
	for _, role := range []string{"test-master", "test-worker", "test-etcd", "test-all"} {
		args = append(args, []string{
			"renew-token",
			"--init-role=" + role,
		})
	}

	for _, arg := range args {
		RunTest(arg, true, dir, t)
	}
}

func TestRenewToken_Fail(t *testing.T) {

	dir, err := TmpDir()
	if err != nil {
		t.Errorf("unexpected error getting tmp dir: %v", err)
		return
	}

	var args [][]string
	for _, role := range []string{"test-foo", "foo", ""} {
		args = append(args, []string{
			"renew-token",
			"--init-role=" + role,
		})
	}

	for _, arg := range args {
		RunTest(arg, false, dir, t)
	}
}
