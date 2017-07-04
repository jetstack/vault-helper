package cmd

import (
	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"

	"fmt"
	"gitlab.jetstack.net/jetstack-experimental/vault-helper/pkg/kubernetes_pki"
	"math/rand"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// initCmd represents the init command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup kubernetes on a running vault server",
	Run: func(cmd *cobra.Command, args []string) {
		//////////////////////
		rand.Seed(time.Now().UnixNano())
		//////////////////////
		// TODO: this should be a cli parameter
		prefix := RandStringBytes(8)
		logrus.Infof("setting up vault on prefix %s", prefix)

		vaultClient, err := vault.NewClient(nil)
		if err != nil {
			logrus.Fatal("unable to create vault client: ", err)
		}

		clusterID := prefix
		basePath := clusterID + "/pki"
		path := basePath + "/" + "etcd-k8s"

		// TODO read env vars and populate

		err = vaultClient.Sys().Mount(
			fmt.Sprintf("%s/pki/etcd-k8s/", prefix),
			&vault.MountInput{
				Description: fmt.Sprintf("Kubernetes %s/etcd-k8s CA", prefix),
				Type:        "pki",
			},
		)

		if err != nil {
			logrus.Fatal("Error mounting etcd-k8s:", err)
		}

		err = vaultClient.Sys().TuneMount(
			fmt.Sprintf("%s/pki/etcd-k8s/", prefix),
			vault.MountConfigInput{
				MaxLeaseTTL: "175320h",
			},
		)

		if err != nil {
			logrus.Fatal("Error tuning etcd-k8s:", err)
		}

		writeData := map[string]interface{}{
			"common_name": fmt.Sprintf("Kubernetes %s/etcd-k8s CA", prefix),
			"ttl":         "175320h",
			"max_ttl":     "175320h",
		}

		_, err = vaultClient.Logical().Write(path+"/root/generate/internal", writeData)

		if err != nil {
			logrus.Fatal("Error writting etcd-k8s data:", err)
		}

		writeData = map[string]interface{}{
			"use_csr_common_name": false,
			"allow_any_name":      true,
			"max_ttl":             "720h",
			"ttl":                 "720h",
			"allow_ip_sans":       "true",
			"server_flag":         "true",
			"client_flag":         "true",
		}

		_, err = vaultClient.Logical().Write(path+"/roles/client", writeData)

		if err != nil {
			logrus.Fatal("Error writting etcd-k8s data [Client]:", err)
		}

		writeData = map[string]interface{}{
			"use_csr_common_name": false,
			"use_csr_sans":        false,
			"allow_any_name":      true,
			"max_ttl":             "720h",
			"ttl":                 "720h",
			"allow_ip_sans":       "true",
			"server_flag":         "true",
			"client_flag":         "true",
		}

		_, err = vaultClient.Logical().Write(path+"/roles/server", writeData)

		if err != nil {
			logrus.Fatal("Error writting etcd-k8s data [Server]:", err)
		}

		/////////////////////////////////////////////////////////////////

		err = vaultClient.Sys().Mount(
			fmt.Sprintf("%s/pki/etcd-overlay/", prefix),
			&vault.MountInput{
				Description: "Kubernetes %s/etcd-overlay CA",
				Type:        "pki",
			},
		)

		if err != nil {
			logrus.Fatal("Error mounting etcd-overlay:", err)
		}

		err = vaultClient.Sys().TuneMount(
			fmt.Sprintf("%s/pki/etcd-k8s/", prefix),
			vault.MountConfigInput{
				MaxLeaseTTL: "175320h",
			},
		)

		if err != nil {
			logrus.Fatal("Error tuning etcd-k8s:", err)
		}

		writeData = map[string]interface{}{
			"common_name": fmt.Sprintf("Kubernetes %s/etcd-overlay CA", prefix),
			"ttl":         "175320h",
		}

		_, err = vaultClient.Logical().Write(path+"/root/generate/internal", writeData)

		if err != nil {
			logrus.Fatal("Error writting etcd-overlay data:", err)
		}

		writeData = map[string]interface{}{
			"use_csr_common_name": false,
			"allow_any_name":      true,
			"max_ttl":             "720h",
			"ttl":                 "720h",
			"allow_ip_sans":       "true",
			"server_flag":         "true",
			"client_flag":         "true",
		}

		_, err = vaultClient.Logical().Write(path+"/roles/client", writeData)

		if err != nil {
			logrus.Fatal("Error writting etcd-overlay data [Client]:", err)
		}

		writeData = map[string]interface{}{
			"use_csr_common_name": false,
			"use_csr_sans":        false,
			"allow_any_name":      true,
			"max_ttl":             "720h",
			"ttl":                 "720h",
			"allow_ip_sans":       "true",
			"server_flag":         "true",
			"client_flag":         "true",
		}

		_, err = vaultClient.Logical().Write(path+"/roles/server", writeData)

		if err != nil {
			logrus.Fatal("Error writting etcd-overlay data [Server]:", err)
		}

		//////////////////////////////////////////////////////////////////////////////

		err = vaultClient.Sys().Mount(
			fmt.Sprintf("%s/pki/k8s/", prefix),
			&vault.MountInput{
				Description: fmt.Sprintf("Kubernetes %s/k8s CA", prefix),
				Type:        "pki",
			},
		)

		if err != nil {
			logrus.Fatal("Error mounting k8s:", err)
		}

		err = vaultClient.Sys().TuneMount(
			fmt.Sprintf("%s/pki/k8s/", prefix),
			vault.MountConfigInput{
				MaxLeaseTTL: "175320h",
			},
		)

		if err != nil {
			logrus.Fatal("Error tunning k8s:", err)
		}

		writeData = map[string]interface{}{
			"common_name": fmt.Sprintf("Kubernetes %s/k8s CA", prefix),
			"ttl":         "175320h",
		}

		_, err = vaultClient.Logical().Write(path+"/root/generate/internal", writeData)

		if err != nil {
			logrus.Fatal("Error writting k8s data:", err)
		}

		writeData = map[string]interface{}{
			"use_csr_common_name": false,
			"enforce_hostname":    false,
			"organization":        "system:masters",
			"allowed_domains":     "admin",
			"allow_bare_domains":  true,
			"allow_localhose":     false,
			"allow_subdomains":    false,
			"allow_ip_sans":       false,
			"server_flag":         false,
			"client_flag":         true,
			"max_ttl":             "8766h",
			"ttl":                 "8766h",
		}

		_, err = vaultClient.Logical().Write(path+"/roles/client", writeData)

		if err != nil {
			logrus.Fatal("Error writting k8s data [Client]:", err)
		}

		roles := []string{"kube-scheduler", "kube-controller-manager", "kube-proxy"}

		// nums := []int{2, 3, 4}
		// sum := 0
		// for _, num := range nums {
		//     sum += num
		// }

		///////////////////////////////////////////////////////////////////////////

		kPKI := kubernetes_pki.New(prefix, vaultClient)

		kPKI.MaxValidityAdmin = time.Hour * 24 * 60

		// TODO ensure that it is setup in that way
		// kPKI.Ensure()
		logrus.Debugf("kpki: %#+v", kPKI)

	},
}

func init() {
	RootCmd.AddCommand(setupCmd)
}
