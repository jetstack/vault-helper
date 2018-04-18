// Copyright Jetstack Ltd. See LICENSE for details.
package cli

import (
	"fmt"
	"testing"
)

func TestCert_Success(t *testing.T) {

	dir, err := TmpDir()
	if err != nil {
		t.Errorf("unexpected error getting tmp dir: %v", err)
		return
	}

	path := fmt.Sprintf("%s/test", dir)

	var args [][]string
	for _, role := range []string{"master", "worker", "etcd", "all"} {
		for _, sign := range []string{"kube-apiserver", "kube-controller-manager", "kube-proxy", "kube-scheduler"} {
			args = append(args, []string{
				"cert",
				"test/pki/k8s/sign/" + sign,
				"0-99-37-81.eu-west-2.compute.internal",
				path,
				"--san-hosts=ip-10-99-37-81.eu-west-2.compute.internal",
				"--ip-sans=10.99.37.81",
				"--key-type=rsa",
				"--key-bit-size=2048",
				"--owner=1000",
				"--group=1000",
				"--init-role=test-" + role,
			})
		}
	}
	for _, role := range []string{"master", "worker", "etcd", "all"} {
		for _, sign := range []string{"admin", "kubelet"} {
			args = append(args, []string{
				"cert",
				"test/pki/k8s/sign/" + sign,
				"system:node:0-99-37-81.eu-west-2.compute.internal",
				path,
				"--san-hosts=ip-10-99-37-81.eu-west-2.compute.internal",
				"--ip-sans=10.99.37.81",
				"--key-type=rsa",
				"--key-bit-size=2048",
				"--owner=1000",
				"--group=1000",
				"--init-role=test-" + role,
			})
		}
	}

	for _, arg := range args {
		RunTest(arg, true, t)
	}
}

func TestCert_Fail(t *testing.T) {

	dir, err := TmpDir()
	if err != nil {
		t.Errorf("unexpected error getting tmp dir: %v", err)
		return
	}

	path := fmt.Sprintf("%s/test", dir)

	var args [][]string
	for _, role := range []string{"master", "worker", "etcd", "all"} {
		for _, sign := range []string{"admin", "kubelet"} {
			args = append(args, []string{
				"cert",
				"test/pki/k8s/sign/" + sign,
				"0-99-37-81.eu-west-2.compute.internal",
				path,
				"--san-hosts=ip-10-99-37-81.eu-west-2.compute.internal",
				"--ip-sans=10.99.37.81",
				"--key-type=rsa",
				"--key-bit-size=2048",
				"--owner=1000",
				"--group=1000",
				"--init-role=test-" + role,
			})
		}
	}

	for _, arg := range args {
		RunTest(arg, false, t)
	}
}
