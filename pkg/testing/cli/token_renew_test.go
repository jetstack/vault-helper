package cli

import (
	"testing"
)

func TestRenewToken_Success(t *testing.T) {

	args := [][]string{
		[]string{"renew-token", "--init-role=test-master"},
		[]string{"renew-token", "--init-role=test-worker"},
		[]string{"renew-token", "--init-role=test-etcd"},
		[]string{"renew-token", "--init-role=test-all"},
	}

	for _, arg := range args {
		RunTest(arg, 0, t)
	}
}

func TestRenewToken_Fail(t *testing.T) {

	args := [][]string{
		[]string{"renew-token", "--init-role=test-foo"},
		[]string{"renew-token", "--init-role=foo"},
		[]string{"renew-token", "--init-role="},
	}
	for _, arg := range args {
		RunTest(arg, 1, t)
	}
}
