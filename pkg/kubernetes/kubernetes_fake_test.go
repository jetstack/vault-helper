package kubernetes

import (
	//"testing"

	"github.com/golang/mock/gomock"
	vault "github.com/hashicorp/vault/api"
	//"gitlab.jetstack.net/jetstack-experimental/vault-helper/pkg/testing/vault_dev"
)

type fakeClient interface {
	Logical() *vault.Logical
}

type fakeServer interface {
	Logical() *vault.Logical
}

type fakeVault struct {
	//*vault_dev.VaultDev
	ctrl *gomock.Controller

	fakeClient fakeClient
	//fakeServer FakeServer
	//vaultRunning chan struct{}
}

//type fakeVault struct {
//	*VaultDev
//	ctrl *gomock.Controller
//
//	client       *vault.Client
//	server       *exec.Cmd
//	vaultRunning chan struct{}

//func newFakeKubernetes(t *testing.T) *Kubernetes {
//	vaultClient := &fakeVault{}
//
//	port := "20202"
//
//	args := []string{
//		"server",
//		"-dev",
//		"-dev-root-token-id=root-token",
//		fmt.Sprintf("-dev-listen-address=127.0.0.1:%d", port),
//	}
//
//	return nil
//
//}

//func (v *fakeVault) Start() error {
//	port := getUnusedPort()
//
//	args := []string{
//		"server",
//		"-dev",
//		"-dev-root-token-id=root-token",
//		fmt.Sprintf("-dev-listen-address=127.0.0.1:%d", port),
//	}
//
//	logrus.Infof("starting vault: %#+v", args)
//
//	v.server = exec.Command("vault", args...)
//
//	err := v.server.Start()
//	if err != nil {
//		return err
//	}
//
//	// this channel will close once vault is stopped
//	v.vaultRunning = make(chan struct{}, 0)
//
//	go func() {
//		err := v.server.Wait()
//		if err != nil {
//			logrus.Warn("vault stopped with error: ", err)
//
//		} else {
//			logrus.Info("vault stopped")
//		}
//		close(v.vaultRunning)
//	}()
//
//	v.client, err = vault.NewClient(&vault.Config{
//		Address: fmt.Sprintf("http://127.0.0.1:%d", port),
//	})
//	if err != nil {
//		return err
//	}
//	v.client.SetToken("root-token")
//
//	tries := 30
//	for {
//		select {
//		case _, open := <-v.vaultRunning:
//			if !open {
//				return fmt.Errorf("vault could not be started")
//			}
//		default:
//		}
//
//		_, err := v.client.Auth().Token().LookupSelf()
//		if err == nil {
//			break
//		}
//		if tries <= 1 {
//			return fmt.Errorf("vault dev server couldn't be started in time")
//		}
//		tries -= 1
//		time.Sleep(time.Second)
//	}
//
//	return nil
//}

//	if err := vault.Start(); err != nil {
//		t.Skip("unable to initialise vault dev server for integration tests: ", err)
//	}
//	defer vault.Stop()
//
//	k, err := New(vault.Client(), "test-cluster-inside")
//	if err != nil {
//		t.Error("unexpected error", err)
//	}
//
//	err = k.Ensure()
//	if err != nil {
//		t.Error("unexpected error: ", err)
//	}
//
//	err = k.Ensure()
//	if err != nil {
//		t.Error("unexpected error: ", err)
//	}

////go test -coverprofile=coverage.out
////  go tool cover -html=coverage.out
//func TestKubernetes_Run_Setup_Test(t *testing.T) {
//	args := []string{"test-cluster-run"}
//	Run(nil, args)
//}
//
//func TestInvalid_Cluster_ID(t *testing.T) {
//	vault := vault_dev.New()
//
//	if err := vault.Start(); err != nil {
//		t.Skip("unable to initialise vault dev server for integration tests: ", err)
//	}
//	defer vault.Stop()
//
//	_, err := New(vault.Client(), "INVALID CLUSTER ID $^^%*$^")
//	if err == nil {
//		t.Error("Should be invalid vluster ID")
//	}
//
//	_, err = New(vault.Client(), "5INVALID CLUSTER ID $^^%*$^")
//	if err == nil {
//		t.Error("Should be invalid vluster ID")
//	}
//
//}
//
//func TestKubernetes_Double_Ensure(t *testing.T) {
//	vault := vault_dev.New()
//
//	if err := vault.Start(); err != nil {
//		t.Skip("unable to initialise vault dev server for integration tests: ", err)
//	}
//	defer vault.Stop()
//
//	k, err := New(vault.Client(), "test-cluster-inside")
//	if err != nil {
//		t.Error("unexpected error", err)
//	}
//
//	err = k.Ensure()
//	if err != nil {
//		t.Error("unexpected error: ", err)
//	}
//
//	err = k.Ensure()
//	if err != nil {
//		t.Error("unexpected error: ", err)
//	}
//
//}
//func TestKubernetes_NewPolicy_Role(t *testing.T) {
//	vault := vault_dev.New()
//
//	if err := vault.Start(); err != nil {
//		t.Skip("unable to initialise vault dev server for integration tests: ", err)
//	}
//	defer vault.Stop()
//
//	k, err := New(vault.Client(), "test-cluster-inside")
//	if err != nil {
//		t.Error("unexpected error", err)
//	}
//
//	policyName := "test-cluster-inside/master"
//	policyRules := "path \"test-cluster-inside/pki/k8s/sign/kube-apiserver\" {\n        capabilities = [\"create\",\"read\",\"update\"]\n    }\n    "
//	role := "master"
//
//	masterPolicy := k.NewPolicy(policyName, policyRules, role)
//
//	err = masterPolicy.WritePolicy()
//	if err != nil {
//		t.Error("unexpected error", err)
//	}
//
//	err = masterPolicy.CreateTokenCreater()
//	if err != nil {
//		t.Error("unexpected error", err)
//	}
//
//}
//
//func TestKubernetes_NewToken_Role(t *testing.T) {
//
//	vault := vault_dev.New()
//
//	if err := vault.Start(); err != nil {
//		t.Skip("unable to initialise vault dev server for integration tests: ", err)
//	}
//	defer vault.Stop()
//
//	k, err := New(vault.Client(), "test-cluster-inside")
//	if err != nil {
//		t.Error("unexpected error", err)
//	}
//	writeData := map[string]interface{}{
//		"use_csr_common_name": false,
//		"enforce_hostnames":   false,
//		"organization":        "system:masters",
//		"allowed_domains":     "admin",
//		"allow_bare_domains":  true,
//		"allow_localhost":     false,
//		"allow_subdomains":    false,
//		"allow_ip_sans":       false,
//		"server_flag":         false,
//		"client_flag":         true,
//		"max_ttl":             "140h",
//		"ttl":                 "140h",
//	}
//
//	adminRole := k.NewTokenRole("admin", writeData)
//	err = adminRole.WriteTokenRole()
//
//	if err != nil {
//		t.Error("unexpected error", err)
//	}
//
//	writeData = map[string]interface{}{
//		"use_csr_common_name": false,
//		"enforce_hostnames":   false,
//		"allowed_domains":     "kube-scheduler,system:kube-scheduler",
//		"allow_bare_domains":  true,
//		"allow_localhost":     false,
//		"allow_subdomains":    false,
//		"allow_ip_sans":       false,
//		"server_flag":         false,
//		"client_flag":         true,
//		"max_ttl":             "140h",
//		"ttl":                 "140h",
//	}
//
//	kubeSchedulerRole := k.NewTokenRole("kube-scheduler", writeData)
//
//	err = kubeSchedulerRole.WriteTokenRole()
//
//	if err != nil {
//		t.Error("unexpected error", err)
//	}
//
//}
