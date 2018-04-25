// Copyright Jetstack Ltd. See LICENSE for details.
package kubernetes

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestGeneric_Ensure(t *testing.T) {
	generic := k.NewGeneric(logrus.NewEntry(logrus.New()))
	if err := generic.Ensure(); err != nil {
		t.Error("unexpected error: ", err)
		return
	}

	if err := generic.Ensure(); err != nil {
		t.Error("unexpected error: ", err)
	}
}
