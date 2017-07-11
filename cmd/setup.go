package cmd

import (
	"github.com/Sirupsen/logrus"
	//vault "github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"

	//"fmt"
	"gitlab.jetstack.net/jetstack-experimental/vault-helper/pkg/kubernetes"

	"gitlab.jetstack.net/jetstack-experimental/vault-helper/pkg/testing/vault_dev"
)

// initCmd represents the init command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup kubernetes on a running vault server",
	Run: func(cmd *cobra.Command, args []string) {

		// TODO: this should be a cli parameter
		clusterID := "vault-setup-test"
		logrus.Infof("setting up vault on prefix %s", clusterID)

		vault := vault_dev.New()
		if err := vault.Start(); err != nil {
			logrus.Fatalf("unable to initialise vault dev server for integration tests: ", err)
		}
		defer vault.Stop()

		k, err := kubernetes.New(vault.Client(), clusterID)
		if err != nil {
			logrus.Fatalf("Unable to create new kubernetes")
		}

		err = k.Ensure()
		if err != nil {
			logrus.Fatalf("Unable to ensure new kubernetes")
		}

		components := []string{"etcd-k8s", "etcd-overlay", "k8s"}
		var writeData map[string]interface{}

		for _, component := range components {

			//path = basePath + "/" + component
			//description = "Kubernetes " + clusterID + "/" + component + " CA"

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
					"max_ttl":             "1440h",
					"ttl":                 "1440h",
				}

				tokenRole := kubernetes.NewTokenRole("admin", writeData, k)
				logrus.Infof("Writting data %s ...", component)
				err = tokenRole.WriteTokenRole()
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

					tokenRole = kubernetes.NewTokenRole(role, writeData, k)
					logrus.Infof("Writting role data %s-%s ...", component, role)
					err = tokenRole.WriteTokenRole()

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

				tokenRole = kubernetes.NewTokenRole("kubelet", writeData, k)
				logrus.Infof("Writting role data %s-kubelet ...", component)
				err = tokenRole.WriteTokenRole()

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

				tokenRole = kubernetes.NewTokenRole("kube-apiserver", writeData, k)
				logrus.Infof("Writting role data %s-kube-apiserver ...", component)
				err = tokenRole.WriteTokenRole()

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

				tokenRole := kubernetes.NewTokenRole("client", writeData, k)
				logrus.Infof("Writting role data %s-[Client] ...", component)
				err = tokenRole.WriteTokenRole()

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

				tokenRole = kubernetes.NewTokenRole("server", writeData, k)
				logrus.Infof("Writting role data %s-[Server] ...", component)
				err = tokenRole.WriteTokenRole()

				if err != nil {
					logrus.Fatal("Error writting "+component+" data [Server]:", err)
				}
				logrus.Infof("Writting role data %s-[Server] success", component)

			}

		}

		generic := kubernetes.NewGeneric(k)
		err = generic.Ensure()
		if err != nil {
			logrus.Fatalf("Unable to ensure new Genetic")
		}

		//kPKI := kubernetes_pki.New(prefix, vault)

		//kPKI.MaxValidityAdmin = time.Hour * 24 * 60

		//// TODO ensure that it is setup in that way
		//// kPKI.Ensure()
		//logrus.Debugf("kpki: %#+v", kPKI)

	},
}

func init() {
	RootCmd.AddCommand(setupCmd)
}
