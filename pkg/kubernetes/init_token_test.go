package kubernetes

import (
	"testing"

	"github.com/golang/mock/gomock"
	vault "github.com/hashicorp/vault/api"
)

// This tests a not yet existing init token
func TestInitToken_Ensure_NoExpectedToken_NotExisting(t *testing.T) {

	fv := NewFakeVault(t)

	i := &InitToken{
		Role:       "etcd",
		Policies:   []string{"etcd"},
		kubernetes: fv.Kubernetes(),
	}

	// expect a read and vault says secret is not existing
	genericPath := "test-cluster-inside/secrets/init_token_etcd"
	fv.fakeLogical.EXPECT().Read(genericPath).Return(
		nil,
		nil,
	)

	// expect a create new orphan
	fv.fakeToken.EXPECT().CreateOrphan(gomock.Any()).Return(&vault.Secret{
		Auth: &vault.SecretAuth{
			ClientToken: "my-new-token",
		},
	}, nil)

	// expect a write of the new token
	fv.fakeLogical.EXPECT().Write(genericPath, map[string]interface{}{"init_token": "my-new-token"}).Return(
		nil,
		nil,
	)

	token, err := i.InitToken()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	if exp, act := "my-new-token", token; exp != act {
		t.Errorf("unexpected token: act=%s exp=%s", act, exp)
	}

}

// Not expceted token set, init token already exists
func TestInitToken_Ensure_NoExpectedToken_AlreadyExisting(t *testing.T) {

	fv := NewFakeVault(t)

	i := &InitToken{
		Role:       "etcd",
		Policies:   []string{"etcd"},
		kubernetes: fv.Kubernetes(),
	}

	// expect a read and vault says secret is not existing
	genericPath := "test-cluster-inside/secrets/init_token_etcd"
	fv.fakeLogical.EXPECT().Read(genericPath).Return(
		&vault.Secret{
			Data: map[string]interface{}{"init_token": "existing-token"},
		},
		nil,
	)

	token, err := i.InitToken()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	if exp, act := "existing-token", token; exp != act {
		t.Errorf("unexpected token: act=%s exp=%s", act, exp)
	}

}

// TODO: Test token where an expected token is given and it's matching

// TODO: Test token where an expected token is given and none is existing

// TODO: Test token where an expected token is given and a conflicting one is existing
