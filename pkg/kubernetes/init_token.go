package kubernetes

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/go-multierror"
	"path/filepath"
	"strings"
)

type InitToken struct {
	Role       string
	Policies   []string
	kubernetes *Kubernetes
	token      *string
}

type Change struct {
	Written bool
	Created bool
}

func (i *InitToken) Ensure() (Change, error) {
	var result error

	ensureInitToken := func() (bool, error) {
		_, written, err := i.InitToken()
		return written, err
	}

	token, err := i.getInitToken()
	if err != nil {
		return Change{false, false}, err
	}

	// Used to pass if there has been change in init tokens
	var change Change

	// If token != user flag and the user token flag != ""
	if token != i.kubernetes.FlagInitTokens[i.Role] && i.kubernetes.FlagInitTokens[i.Role] != "" {
		// Write the init token role and policy using the user token flag
		for _, f := range []func() error{
			i.writeTokenRole,
			i.writeInitTokenPolicy,
		} {
			if err := f(); err != nil {
				result = multierror.Append(result, err)
			}
			if _, err := ensureInitToken(); err != nil {
				result = multierror.Append(result, err)
			}
		}
		b, err := i.setInitToken(fmt.Sprintf("%s", i.kubernetes.FlagInitTokens[i.Role]))
		if err != nil {
			return change, fmt.Errorf("Failed to set '%s' init token: '%s'", i.Role, err)
		}
		// User token written - not created
		change.Written = b
		change.Created = false

		// Token == user flag and the flag != "" - just need to ensure the init token
	} else if token == i.kubernetes.FlagInitTokens[i.Role] && i.kubernetes.FlagInitTokens[i.Role] != "" {
		if written, err := ensureInitToken(); err != nil {
			result = multierror.Append(result, err)
		} else if written {
			change.Written = true
			change.Created = false
		}

		// No flag. Generate an init token and write to vault
	} else {
		for _, f := range []func() error{
			i.writeTokenRole,
			i.writeInitTokenPolicy,
		} {
			if err := f(); err != nil {
				result = multierror.Append(result, err)
			}
			if written, err := ensureInitToken(); err != nil {
				result = multierror.Append(result, err)
			} else if written {
				change.Created = true
				change.Written = true
			}
		}

	}

	return change, result
}

// Write init token from user flag
func (i *InitToken) setInitToken(string) (bool, error) {
	path := filepath.Join(i.kubernetes.secretsGeneric.Path(), fmt.Sprintf("init_token_%s", i.Role))

	s, err := i.kubernetes.vaultClient.Logical().Read(path)
	if err != nil {
		return false, fmt.Errorf("Error reading init token path: %s", s)
	}

	s.Data["init_token"] = i.kubernetes.FlagInitTokens[i.Role]
	_, err = i.kubernetes.vaultClient.Logical().Write(path, s.Data)
	if err != nil {
		return false, fmt.Errorf("Error writting init token at path: %s", s)
	}

	token, _ := i.getInitToken()
	logrus.Infof("User token written for %s: %s", i.Role, token)
	i.token = &token

	return true, nil
}

// Read init token from vault at path
func (i *InitToken) getInitToken() (string, error) {
	path := filepath.Join(i.kubernetes.secretsGeneric.Path(), fmt.Sprintf("init_token_%s", i.Role))

	s, err := i.kubernetes.vaultClient.Logical().Read(path)
	if err != nil {
		return "", fmt.Errorf("Error reading init token.", err)
	}
	if s == nil {
		return "", nil
	}

	return fmt.Sprintf("%s", s.Data["init_token"]), nil
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
		return fmt.Errorf("error writing token role %s: %s", i.Path(), err)
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
func (i *InitToken) InitToken() (string, bool, error) {
	if i.token != nil {
		return *i.token, false, nil
	}

	// get init token from generic
	token, written, err := i.kubernetes.secretsGeneric.InitToken(i.Name(), i.Role, []string{fmt.Sprintf("%s-creator", i.namePath())})
	if err != nil {
		return "", false, err
	}

	i.token = &token
	return token, written, nil
}
