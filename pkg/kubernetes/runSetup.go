package kubernetes

import (
	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.jetstack.net/jetstack-experimental/vault-helper/pkg/testing/vault_dev"
	"time"
)

const MAX_VALIDITY_COMPONENTS = time.Hour * 720
const MAX_VALIDITY_ADMIN = time.Hour * 8766
const MAX_VALIDITY_CA = time.Hour * 175320

func Run(cmd *cobra.Command, args []string) {

	var clusterID string

	if len(args) > 0 {
		clusterID = args[0]
	} else {
		logrus.Fatalf("No cluster id was given")
	}

	logrus.Infof("setting up vault on prefix %s", clusterID)

	vault := vault_dev.New()
	if err := vault.Start(); err != nil {
		logrus.Fatalf("unable to initialise vault dev server for integration tests: ", err)
	}
	defer vault.Stop()

	k, err := New(vault.Client(), clusterID)
	if err != nil {
		logrus.Fatalf("Unable to create new kubernetes")
	}

	err = k.Ensure()
	if err != nil {
		logrus.Fatalf("Unable to ensure new kubernetes: ", err)
	}

	components := []string{"etcd-k8s", "etcd-overlay", "k8s"}
	var writeData map[string]interface{}

	for _, component := range components {

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
				"max_ttl":             MAX_VALIDITY_ADMIN,
				"ttl":                 MAX_VALIDITY_ADMIN,
			}

			componentRole := k.NewComponentRole(component, "admin", writeData)
			logrus.Infof("Writting data %s ...", component)
			err = componentRole.WriteComponentRole()
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
					"max_ttl":             MAX_VALIDITY_COMPONENTS,
					"ttl":                 MAX_VALIDITY_COMPONENTS,
				}

				componentRole = k.NewComponentRole(component, role, writeData)
				logrus.Infof("Writting role data %s-%s ...", component, role)
				err = componentRole.WriteComponentRole()

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
				"max_ttl":             MAX_VALIDITY_COMPONENTS,
				"ttl":                 MAX_VALIDITY_COMPONENTS,
			}

			componentRole = k.NewComponentRole(component, "kubelet", writeData)
			logrus.Infof("Writting role data %s-kubelet ...", component)
			err = componentRole.WriteComponentRole()

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
				"max_ttl":             MAX_VALIDITY_COMPONENTS,
				"ttl":                 MAX_VALIDITY_COMPONENTS,
			}

			componentRole = k.NewComponentRole(component, "kube-apiserver", writeData)
			logrus.Infof("Writting role data %s-kube-apiserver ...", component)
			err = componentRole.WriteComponentRole()

			if err != nil {
				logrus.Fatal("Error writting k8s data [Kublet]:", err)
			}
			logrus.Infof("Writting role data %s-kube-apiserver success", component)

		} else {
			writeData = map[string]interface{}{
				"use_csr_common_name": false,
				"allow_any_name":      true,
				"max_ttl":             MAX_VALIDITY_COMPONENTS,
				"ttl":                 MAX_VALIDITY_COMPONENTS,
				"allow_ip_sans":       "true",
				"server_flag":         "true",
				"client_flag":         "true",
			}

			componentRole := k.NewComponentRole(component, "client", writeData)
			logrus.Infof("Writting role data %s-[Client] ...", component)
			logrus.Info(MAX_VALIDITY_COMPONENTS)
			logrus.Info(MAX_VALIDITY_CA)
			err = componentRole.WriteComponentRole()

			if err != nil {
				logrus.Fatal("Error writting "+component+" data [Client]:", err)
			}
			logrus.Infof("Writting role data %s-[Client] success", component)

			writeData = map[string]interface{}{
				"use_csr_common_name": false,
				"use_csr_sans":        false,
				"allow_any_name":      true,
				"max_ttl":             MAX_VALIDITY_COMPONENTS,
				"ttl":                 MAX_VALIDITY_COMPONENTS,
				"allow_ip_sans":       "true",
				"server_flag":         "true",
				"client_flag":         "true",
			}

			componentRole = k.NewComponentRole(component, "server", writeData)
			logrus.Infof("Writting role data %s-[Server] ...", component)
			err = componentRole.WriteComponentRole()

			if err != nil {
				logrus.Fatal("Error writting "+component+" data [Server]:", err)
			}
			logrus.Infof("Writting role data %s-[Server] success", component)

		}

	}

	generic := k.NewGeneric()
	err = generic.Ensure()
	if err != nil {
		logrus.Fatalf("Unable to ensure new Genetic")
	}

	basePath := clusterID + "/pki"
	secrets_path := clusterID + "/secrets"

	for _, role := range []string{"master", "worker", "etcd"} {
		policy_name := clusterID + "/" + role

		if role == "master" || role == "worker" {
			for _, cert_role := range []string{"k8s/sign/kubelet", "k8s/sign/kube-proxy", "etcd-overlay/sign/client"} {
				rule := "\npath \"" + basePath + "/" + cert_role + "\" {\n    capabilities = [\"create\",\"read\",\"update\"]\n}\n"
				policy := k.NewPolicy(policy_name, rule, role)

				err = policy.WritePolicy()
				if err != nil {
					logrus.Fatalf("Error writting policy: ", err)
				}

			}

		}

		if role == "master" {
			for _, cert_role := range []string{"k8s/sign/kube-apiserver", "k8s/sign/kube-scheduler", "k8s/sign/kube-controller-manager", "k8s/sign/admin", "etcd-k8s/sign/client"} {
				rule := "path \"" + basePath + "/" + cert_role + "\" {\n    capabilities = [\"create\",\"read\",\"update\"]\n}\n"
				policy := k.NewPolicy(policy_name, rule, role)

				err = policy.WritePolicy()
				if err != nil {
					logrus.Fatalf("Error writting policy: ", err)
				}

			}
		}

		rule := "\npath \"" + secrets_path + "/service-accounts\" {\n    capabilities = [\"read\"]\n}\n"
		policy := k.NewPolicy(policy_name, rule, role)

		err = policy.WritePolicy()
		if err != nil {
			logrus.Fatalf("Error writting policy: ", err)
		}

		if role == "etcd" {
			for _, cert_role := range []string{"etcd-k8s/sign/server", "etcd-overlay/sign/server"} {

				rule := "path \"" + basePath + "/" + cert_role + "\" {\n    capabilities = [\"create\",\"read\",\"update\"]\n}\n"
				policy := k.NewPolicy(policy_name, rule, role)

				err = policy.WritePolicy()
				if err != nil {
					logrus.Fatalf("Error writting policy: ", err)
				}

			}
		}

		writeData = map[string]interface{}{
			"period":           "720h",
			"orphan":           true,
			"allowed_policies": "default," + policy_name,
			"path_suffix":      policy_name,
		}

		tokenRole := k.NewTokenRole(role, writeData)
		err = tokenRole.WriteTokenRole()
		if err != nil {
			logrus.Fatalf("Error writting token role: ", err)
		}

		initToken := k.NewInitToken(policy_name, role)
		err = initToken.CreateToken()
		if err != nil {
			logrus.Fatalf("Error creating init token", err)
		}
		err = initToken.WriteInitToken()
		if err != nil {
			logrus.Fatalf("Error creating init token", err)
		}

	}

	//kPKI := kubernetes_pki.New(prefix, vault)

	//kPKI.MaxValidityAdmin = time.Hour * 24 * 60

	//// TODO ensure that it is setup in that way
	//// kPKI.Ensure()
	//logrus.Debugf("kpki: %#+v", kPKI)

}
