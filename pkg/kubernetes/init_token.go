package kubernetes

import (
	"fmt"
	//"github.com/Sirupsen/logrus"
	"github.com/hashicorp/go-multierror"
	"path/filepath"
	"strings"
)

type InitToken struct {
	Role          string
	Policies      []string
	kubernetes    *Kubernetes
	token         *string
	ExpectedToken string
}

func (i *InitToken) Ensure() error {
	var result error

	ensureInitToken := func() error {
		_, err := i.InitToken()
		return err
	}

	writeTokenRole_Police := []func() error{
		i.writeTokenRole,
		i.writeInitTokenPolicy,
	}

	token, err := i.secretsGeneric().InitTokenStore(i.Role)
	if err != nil {
		return err
	}

	// If token != user flag and the user token flag != ""
	if token != i.ExpectedToken && i.ExpectedToken != "" {
		// Write the init token role and policy using the user token flag
		for _, f := range writeTokenRole_Police {
			if err := f(); err != nil {
				result = multierror.Append(result, err)
			}
			if err := ensureInitToken(); err != nil {
				result = multierror.Append(result, err)
			}
		}

		err := i.secretsGeneric().SetInitTokenStore(i.Role, i.ExpectedToken)
		if err != nil {
			return fmt.Errorf("Failed to set '%s' init token: '%s'", i.Role, err)
		}

		tokenStr, err := i.secretsGeneric().InitTokenStore(i.Role)
		if err != nil {
			return fmt.Errorf("Failed to read '%s' init token: '%s'", i.Role, err)
		}
		i.token = &tokenStr

		// Token == user flag and the flag != "" - just need to ensure the init token
	} else if token == i.ExpectedToken && i.ExpectedToken != "" {
		if err := ensureInitToken(); err != nil {
			result = multierror.Append(result, err)
		}

		// No flag. Generate an init token and write to vault
	} else {
		for _, f := range writeTokenRole_Police {
			if err := f(); err != nil {
				result = multierror.Append(result, err)
			}
			if err := ensureInitToken(); err != nil {
				result = multierror.Append(result, err)
			}
		}

	}

	return result
}

// Write init token from user flag
func (i *InitToken) setInitToken(string) error {
	return nil
}

// Get init token name
func (i *InitToken) Name() string {
	return fmt.Sprintf("%s-%s", i.kubernetes.clusterID, i.Role)
}

// Get name path suffix for token role
func (i *InitToken) namePath() string {
	return fmt.Sprintf("%s/%s", i.kubernetes.clusterID, i.Role)
}

// Construct file path for ../create
func (i *InitToken) createPath() string {
	return filepath.Join("auth/token/create", i.Name())
}

// Construct file path for ../auth
func (i *InitToken) Path() string {
	return filepath.Join("auth/token/roles", i.Name())
}

// Write token role to vault
func (i *InitToken) writeTokenRole() error {
	policies := i.Policies
	policies = append(policies, "default")

	writeData := map[string]interface{}{
		"period":           fmt.Sprintf("%ds", int(i.kubernetes.MaxValidityComponents.Seconds())),
		"orphan":           true,
		"allowed_policies": strings.Join(policies, ","),
		"path_suffix":      i.namePath(),
	}

	_, err := i.kubernetes.vaultClient.Logical().Write(i.Path(), writeData)
	if err != nil {
		return fmt.Errorf("error writing token role %s: %v", i.Path(), err)
	}

	return nil
}

// Construct policy and send to kubernetes to be written to vault
func (i *InitToken) writeInitTokenPolicy() error {
	p := &Policy{
		Name: fmt.Sprintf("%s-creator", i.namePath()),
		Policies: []*policyPath{
			&policyPath{
				path:         i.createPath(),
				capabilities: []string{"create", "read", "update"},
			},
		},
	}
	return i.kubernetes.WritePolicy(p)
}

// Return init token if token exists
// Retrieve from generic if !exists
func (i *InitToken) InitToken() (string, error) {
	if i.token != nil {
		return *i.token, nil
	}

	// get init token from generic
	token, err := i.secretsGeneric().InitToken(i.Name(), i.Role, []string{fmt.Sprintf("%s-creator", i.namePath())})
	if err != nil {
		return "", err
	}

	i.token = &token
	return token, nil
}

func (i *InitToken) secretsGeneric() *Generic {
	return i.kubernetes.secretsGeneric
}
