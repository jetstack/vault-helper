// Copyright Jetstack Ltd. See LICENSE for details.
package cli

import (
	"testing"
)

func TestRead_Success(t *testing.T) {

	dir, err := TmpDir()
	if err != nil {
		t.Errorf("unexpected error getting tmp dir: %v", err)
		return
	}

	var args [][]string
	for _, role := range []string{"test-master", "test-all"} {
		args = append(args, []string{
			"read",
			"test/secrets/service-accounts",
			"--init-role=" + role,
		})
	}

	for _, arg := range args {
		RunTest(arg, true, dir, t)
	}
}

func TestRead_Fail(t *testing.T) {

	dir, err := TmpDir()
	if err != nil {
		t.Errorf("unexpected error getting tmp dir: %v", err)
		return
	}

	var args [][]string
	for _, role := range []string{"test-worker", "test-etcd"} {
		args = append(args, []string{
			"read",
			"test/secrets/service-accounts",
			"--init-role=" + role,
		})
	}

	for _, arg := range [][]string{
		[]string{"service-accounts", "test-worker"},
		[]string{"service-accounts", "test-etcd"},
		[]string{"init_token_all", "test-all"},
		[]string{"init_token_etcd", "test-etcd"},
		[]string{"init_token_master", "test-master"},
		[]string{"init_token_worker", "test-worker"},
		[]string{"init_token_all", "test-master"},
		[]string{"init_token_etcd", "test-master"},
		[]string{"init_token_worker", "test-master"},
	} {
		args = append(args, []string{"read", "test/secrets/" + arg[0], "--init-role=" + arg[1]})
	}

	for _, arg := range args {
		RunTest(arg, false, dir, t)
	}
}
