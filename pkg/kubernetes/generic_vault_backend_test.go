// Copyright Jetstack Ltd. See LICENSE for details.
package kubernetes

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestGenericVaultBackend_Ensure(t *testing.T) {
	backend := k.NewGenericVaultBackend(logrus.NewEntry(logrus.New()))
	if err := backend.Ensure(); err != nil {
		t.Error("unexpected error: ", err)
		return
	}

	if err := backend.Ensure(); err != nil {
		t.Error("unexpected error: ", err)
	}
}
