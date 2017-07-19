package kubernetes

import (
	"strings"
	"testing"
)

func TestIsValidClusterID(t *testing.T) {
	var err error

	err = isValidClusterID("valid-cluster")
	if err != nil {
		t.Error("unexpected an error: %s", err)
	}

	err = isValidClusterID("")
	if err == nil {
		t.Error("expected an error")
	} else if msg := "invalid clusterID"; !strings.Contains(err.Error(), msg) {
		t.Errorf("error '%s' should contain '%s'", err, msg)
	}

	err = isValidClusterID("invalid.cluster")
	if err == nil {
		t.Error("expected an error")
	} else if msg := "invalid clusterID"; !strings.Contains(err.Error(), msg) {
		t.Errorf("error '%s' should contain '%s'", err, msg)
	}

}
