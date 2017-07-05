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
		path := ""
		components := []string{"etcd-k8s", "etcd-overlay", "k8s"}
		var description string
		var writeData map[string]interface{}

		for _, component := range components {
			path = basePath + "/" + component
			description = "Kubernetes " + clusterID + "/" + component + " CA"

			logrus.Infof("Mounting component %s ...", component)
			err = vaultClient.Sys().Mount(
				fmt.Sprintf("%s/pki/"+component+"/", clusterID),
				&vault.MountInput{
					Description: description,
					Type:        "pki",
				},
			)

			if err != nil {
				logrus.Fatal("Error mounting "+component+":", err)
			}
			logrus.Infof("Mounting component %s success", component)

			logrus.Infof("Tuning Mount %s ...", component)
			err = vaultClient.Sys().TuneMount(
				fmt.Sprintf("%s/pki/"+component+"/", clusterID),
				vault.MountConfigInput{
					MaxLeaseTTL: "175320h",
				},
			)

			if err != nil {
				logrus.Fatal("Error tuning "+component+":", err)
			}
			logrus.Infof("Tuning Mount %s success", component)

			sec, err := vaultClient.Logical().Read(path + "/cert/ca")

			if _, ok := sec.Data["certificate"]; ok {
				logrus.Infof("CA not found for %s ...", component)

				if err != nil {
					logrus.Fatal("Error reading "+component+" certificate:", err)
				}

				writeData = map[string]interface{}{
					"common_name": fmt.Sprintf("Kubernetes %s/"+component+" CA", clusterID),
					"ttl":         "175320h",
				}

				_, err = vaultClient.Logical().Write(path+"/root/generate/internal", writeData)

				if err != nil {
					logrus.Fatal("Error writting "+component+" data:", err)
				}
				logrus.Infof("CA created for %s success", component)
			}

			if component == "k8s" {

				writeData = map[string]interface{}{
					"use_csr_common_name": false,
					"enforce_hostnames":   false,
					"organization":        "system:masters",
					"allowed_domains":     "admin",
					"allow_bare_domains":  true,
					"allow_localhost":     false,
					"allow_subdomains":    false,
					"allow_ip_sans":       false,
					"server_flag":         false,
					"client_flag":         true,
					"max_ttl":             "8766h",
					"ttl":                 "8766h",
				}

				logrus.Infof("Writting data %s ...", component)
				_, err = vaultClient.Logical().Write(path+"/roles/admin", writeData)

				if err != nil {
					logrus.Fatal("Error writting k8s data [Admin]:", err)
				}
				logrus.Infof("Writting data %s success", component)

				roles := []string{"kube-scheduler", "kube-controller-manager", "kube-proxy"}

				for _, role := range roles {
					writeData = map[string]interface{}{
						"use_csr_common_name": false,
						"enforce_hostnames":   false,
						"allowed_domains":     role + ",system:" + role,
						"allow_bare_domains":  true,
						"allow_localhost":     false,
						"allow_subdomains":    false,
						"allow_ip_sans":       false,
						"server_flag":         false,
						"client_flag":         true,
						"max_ttl":             "8766h",
						"ttl":                 "8766h",
					}

					logrus.Infof("Writting role data %s-%s ...", component, role)
					_, err = vaultClient.Logical().Write(path+"/roles/"+role, writeData)

					if err != nil {
						logrus.Fatal("Error writting k8s role:"+role+" data", err)
					}
					logrus.Infof("Writting role data %s-%s success", component, role)

				}

				writeData = map[string]interface{}{
					"use_csr_common_name": false,
					"use_csr_sans":        false,
					"enforce_hostnames":   false,
					"organization":        "system:nodes",
					"allowed_domains":     "kubelet,system:node,system:node:*",
					"allow_bare_domains":  true,
					"allow_glob_domains":  true,
					"allow_localhost":     false,
					"allow_subdomains":    false,
					"server_flag":         true,
					"client_flag":         true,
					"max_ttl":             "8766h",
					"ttl":                 "8766h",
				}

				logrus.Infof("Writting role data %s-kubelet ...", component)
				_, err = vaultClient.Logical().Write(path+"/roles/kubelet", writeData)

				if err != nil {
					logrus.Fatal("Error writting k8s data [Kublet]:", err)
				}
				logrus.Infof("Writting role data %s-kubelet success", component)

				writeData = map[string]interface{}{
					"use_csr_common_name": false,
					"use_csr_sans":        false,
					"enforce_hostnames":   false,
					"allow_localhost":     true,
					"allow_any_name":      true,
					"allow_bare_domains":  true,
					"allow_ip_sans":       true,
					"server_flag":         true,
					"client_flag":         false,
					"max_ttl":             "8766h",
					"ttl":                 "8766h",
				}

				logrus.Infof("Writting role data %s-kube-apiserver ...", component)
				_, err = vaultClient.Logical().Write(path+"/roles/kube-apiserver", writeData)

				if err != nil {
					logrus.Fatal("Error writting k8s data [Kublet]:", err)
				}
				logrus.Infof("Writting role data %s-kube-apiserver success", component)

			} else {
				writeData = map[string]interface{}{
					"use_csr_common_name": false,
					"allow_any_name":      true,
					"max_ttl":             "720h",
					"ttl":                 "720h",
					"allow_ip_sans":       "true",
					"server_flag":         "true",
					"client_flag":         "true",
				}
				logrus.Infof("Writting role data %s-[Client] ...", component)

				_, err = vaultClient.Logical().Write(path+"/roles/client", writeData)

				if err != nil {
					logrus.Fatal("Error writting "+component+" data [Client]:", err)
				}
				logrus.Infof("Writting role data %s-[Client] success", component)

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
				logrus.Infof("Writting role data %s-[Server] ...", component)

				_, err = vaultClient.Logical().Write(path+"/roles/server", writeData)

				if err != nil {
					logrus.Fatal("Error writting "+component+" data [Server]:", err)
				}
				logrus.Infof("Writting role data %s-[Server] success", component)

			}

		}

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
