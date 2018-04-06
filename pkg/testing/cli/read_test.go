package cli

import (
	"testing"
)

func TestRead_Success(t *testing.T) {

	args := [][]string{
		[]string{"read", "test/secrets/service-accounts", "--init-role=test-master"},
		[]string{"read", "test/secrets/init_token_all", "--init-role=test-all"},
		[]string{"read", "test/secrets/init_token_etcd", "--init-role=test-etcd"},
		[]string{"read", "test/secrets/init_token_master", "--init-role=test-master"},
		[]string{"read", "test/secrets/init_token_worker", "--init-role=test-worker"},

		[]string{"read", "test/secrets/init_token_all", "--init-role=test-master"},
		[]string{"read", "test/secrets/init_token_etcd", "--init-role=test-master"},
		[]string{"read", "test/secrets/init_token_worker", "--init-role=test-master"},
	}

	for _, arg := range args {
		RunTest(arg, 0, t)
	}
}

func TestRead_Fail(t *testing.T) {

	args := [][]string{
		[]string{"renew-token", "--init-role=test-foo"},
		[]string{"renew-token", "--init-role=foo"},
		[]string{"renew-token", "--init-role="},
	}
	for _, arg := range args {
		RunTest(arg, 1, t)
	}
}
