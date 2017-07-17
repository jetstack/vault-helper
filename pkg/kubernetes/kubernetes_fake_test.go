package kubernetes

import (
	//"testing"

	"github.com/golang/mock/gomock"
	vault "github.com/hashicorp/vault/api"
	//mocks "gitlab.jetstack.net/jetstack-experimental/vault-helper/pkg/mocks"
	//"github.com/Sirupsen/logrus"
)

type fakeVault struct {
	fakeVault   *MockVault
	fakeSys     *MockVaultSys
	fakeLogical *MockVaultLogical
	fakeAuth    *MockVaultAuth
}

func NewFakeVault(ctrl *gomock.Controller) *fakeVault {
	v := &fakeVault{
		fakeVault: NewMockVault(ctrl),
		fakeSys:   NewMockVaultSys(ctrl),
	}

	v.fakeVault.EXPECT().Sys().AnyTimes().Return(v.fakeSys)
	v.fakeVault.EXPECT().Logical().AnyTimes().Return(v.fakeLogical)
	v.fakeVault.EXPECT().Auth().AnyTimes().Return(v.fakeAuth)

	return v

}

func DoubleEnsure(v *fakeVault) {

	mountInput1 := &vault.MountInput{
		Description: "Kubernetes test-cluster-inside/etcd-k8s CA",
		Type:        "pki",
		Config: vault.MountConfigInput{
			DefaultLeaseTTL: "175320",
			MaxLeaseTTL:     "175320",
		},
	}

	mountInput2 := &vault.MountInput{
		Description: "Kubernetes " + "test-cluster-inside" + "/" + "etcd-overlay" + " CA",
		Type:        "pki",
		Config: vault.MountConfigInput{
			DefaultLeaseTTL: "175320",
			MaxLeaseTTL:     "175320",
		},
	}

	mountInput3 := &vault.MountInput{
		Description: "Kubernetes " + "test-cluster-inside" + "/" + "k8s" + " CA",
		Type:        "pki",
		Config: vault.MountConfigInput{
			DefaultLeaseTTL: "175320",
			MaxLeaseTTL:     "175320",
		},
	}

	v.fakeSys.EXPECT().ListMounts().AnyTimes().Return(nil, nil)
	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/etcd-k8s", mountInput1).Times(2).Return(nil)
	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/etcd-overlay", mountInput2).Times(2).Return(nil)
	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/k8s", mountInput3).Times(2).Return(nil)
	//mounts := map[string]*vault.MountOutput{
	//	Type: "generic",
	//}
	//v.fakeSys.EXPECT().ListMounts().Times(1).Return(mounts, nil)

}
