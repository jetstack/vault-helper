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

func (i *InitToken) Ensure() (map[string]interface{}, error) {
	var result error

	ensureInitToken := func() (bool, error) {
		_, written, err := i.InitToken()
		return written, err
	}

	token, _ := i.GetInitToken()

	change := map[string]interface{}{
		"created_init": false,
		"written_init": false,
	}

	if token != i.kubernetes.FlagInitTokens[i.Role] && i.kubernetes.FlagInitTokens[i.Role] != "" {
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
		b, err := i.SetInitToken(fmt.Sprintf("%s", i.kubernetes.FlagInitTokens[i.Role]))
		if err != nil {
			return change, fmt.Errorf("Failed to set '%s' init token: '%s'", i.Role, err)
		}
		change["written_init"] = b
		change["created_init"] = false

	} else if token == i.kubernetes.FlagInitTokens[i.Role] && i.kubernetes.FlagInitTokens[i.Role] != "" {
		if written, err := ensureInitToken(); err != nil {
			result = multierror.Append(result, err)
		} else if written {
			change["written_init"] = true
			change["created_init"] = false
		}

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
				change["created_init"] = true
				change["written_init"] = true
			}
		}

	}

	return change, result
}

func (i *InitToken) SetInitToken(string) (bool, error) {
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

	token, _ := i.GetInitToken()
	logrus.Infof("Token written: %s", token)
	i.token = &token

	return true, nil
}

func (i *InitToken) GetInitToken() (string, error) {
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

func (i *InitToken) Name() string {
	return fmt.Sprintf("%s-%s", i.kubernetes.clusterID, i.Role)
}

func (i *InitToken) NamePath() string {
	return fmt.Sprintf("%s/%s", i.kubernetes.clusterID, i.Role)
}

func (i *InitToken) CreatePath() string {
	return filepath.Join("auth/token/create", i.Name())
}

func (i *InitToken) Path() string {
	return filepath.Join("auth/token/roles", i.Name())
}

func (i *InitToken) writeTokenRole() error {
	policies := i.Policies
	policies = append(policies, "default")

	writeData := map[string]interface{}{
		"period":           fmt.Sprintf("%ds", int(i.kubernetes.MaxValidityComponents.Seconds())),
		"orphan":           true,
		"allowed_policies": strings.Join(policies, ","),
		"path_suffix":      i.NamePath(),
	}

	_, err := i.kubernetes.vaultClient.Logical().Write(i.Path(), writeData)
	if err != nil {
		return fmt.Errorf("error writing token role %s: %s", i.Path(), err)
	}

	return nil
}

func (i *InitToken) writeInitTokenPolicy() error {
	p := &Policy{
		Name: fmt.Sprintf("%s-creator", i.NamePath()),
		Policies: []*policyPath{
			&policyPath{
				path:         i.CreatePath(),
				capabilities: []string{"create", "read", "update"},
			},
		},
	}
	return i.kubernetes.WritePolicy(p)
}

func (i *InitToken) InitToken() (string, bool, error) {
	if i.token != nil {
		return *i.token, false, nil
	}

	// get init token from generic
	token, written, err := i.kubernetes.secretsGeneric.InitToken(i.Name(), i.Role, []string{fmt.Sprintf("%s-creator", i.NamePath())})
	if err != nil {
		return "", false, err
	}

	i.token = &token
	return token, written, nil
}
