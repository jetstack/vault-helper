package kubernetes

import (
	"testing"

	"github.com/golang/mock/gomock"
	vault "github.com/hashicorp/vault/api"
)

// tests a not yet existing init token, with random generated token
func TestInitToken_Ensure_NoExpectedToken_NotExisting(t *testing.T) {
	fv := NewFakeVault(t)
	defer fv.Finish()

	i := &InitToken{
		Role:          "etcd",
		Policies:      []string{"etcd"},
		kubernetes:    fv.Kubernetes(),
		ExpectedToken: "",
	}

	// expects a read and vault says secret is not existing
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

	fv.fakeLogical.EXPECT().Read(genericPath).Return(
		&vault.Secret{
			Data: map[string]interface{}{"init_token": "my-new-token"},
		},
		nil,
	)

	InitTokenEnsure_EXPECTs(fv)

	err := i.Ensure()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	token, err := i.InitToken()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	if exp, act := "my-new-token", token; exp != act {
		t.Errorf("unexpected token: act=%s exp=%s", act, exp)
	}

	return
}

// expected token not set, init token already exists
func TestInitToken_Ensure_NoExpectedToken_AlreadyExisting(t *testing.T) {
	fv := NewFakeVault(t)
	defer fv.Finish()

	i := &InitToken{
		Role:          "etcd",
		Policies:      []string{"etcd"},
		kubernetes:    fv.Kubernetes(),
		ExpectedToken: "",
	}

	// expect a read and vault says secret is existing
	genericPath := "test-cluster-inside/secrets/init_token_etcd"
	fv.fakeLogical.EXPECT().Read(genericPath).Times(2).Return(
		&vault.Secret{
			Data: map[string]interface{}{"init_token": "existing-token"},
		},
		nil,
	)

	InitTokenEnsure_EXPECTs(fv)

	err := i.Ensure()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	token, err := i.InitToken()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	if exp, act := "existing-token", token; exp != act {
		t.Errorf("unexpected token: act=%s exp=%s", act, exp)
	}

	return
}

// expected token set, init token already exists and it's matching
func TestInitToken_Ensure_ExpectedToken_Existing_Match(t *testing.T) {
	fv := NewFakeVault(t)
	defer fv.Finish()

	i := &InitToken{
		Role:          "etcd",
		Policies:      []string{"etcd"},
		kubernetes:    fv.Kubernetes(),
		ExpectedToken: "expected-token",
	}

	// expect a read and vault says secret is existing
	genericPath := "test-cluster-inside/secrets/init_token_etcd"
	fv.fakeLogical.EXPECT().Read(genericPath).Times(2).Return(
		&vault.Secret{
			Data: map[string]interface{}{"init_token": "expected-token"},
		},
		nil,
	)

	InitTokenEnsure_EXPECTs(fv)

	err := i.Ensure()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	token, err := i.InitToken()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	if exp, act := "expected-token", token; exp != act {
		t.Errorf("unexpected token: act=%s exp=%s", act, exp)
	}

	return
}

// expected token set, init token doesn't exist
func TestInitToken_Ensure_ExpectedToken_NotExisting(t *testing.T) {
	fv := NewFakeVault(t)
	defer fv.Finish()

	i := &InitToken{
		Role:          "etcd",
		Policies:      []string{"etcd"},
		kubernetes:    fv.Kubernetes(),
		ExpectedToken: "expected-token",
	}

	// expect a read and vault says secret is not existing
	genericPath := "test-cluster-inside/secrets/init_token_etcd"
	fv.fakeLogical.EXPECT().Read(genericPath).Times(2).Return(
		nil,
		nil,
	)

	// expect a write of the new token from user flag
	fv.fakeLogical.EXPECT().Write(genericPath, map[string]interface{}{"init_token": "expected-token"}).Return(
		nil,
		nil,
	)

	// expect to read out token from user
	fv.fakeLogical.EXPECT().Read(genericPath).Times(2).Return(
		&vault.Secret{
			Data: map[string]interface{}{"init_token": "expected-token"},
		},
		nil,
	)

	InitTokenEnsure_EXPECTs(fv)

	err := i.Ensure()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	token, err := i.InitToken()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	if exp, act := "expected-token", token; exp != act {
		t.Errorf("unexpected token: act=%s exp=%s", act, exp)
	}

	return
}

// General policy and write calls when init token ensuring
func InitTokenEnsure_EXPECTs(fv *fakeVault) {
	fv.fakeLogical.EXPECT().Write("auth/token/roles/test-cluster-inside-etcd", gomock.Any()).AnyTimes().Return(nil, nil)
	fv.fakeSys.EXPECT().PutPolicy(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
}
