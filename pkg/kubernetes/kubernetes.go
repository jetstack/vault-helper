package kubernetes

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
)

type Backend interface {
	Ensure() error
	Path() string
}

var _ Backend = &PKI{}
var _ Backend = &Generic{}

type Kubernetes struct {
	clusterID string // clusterID is required parameter, lowercase only, [a-z0-9-]+

	etcdKubernetesPKI *PKI
	etcdOverlayPKI    *PKI
	kubernetesPKI     *PKI
	secretsGeneric    *Generic
}

func New(clusterID string) *Kubernetes {

	// TODO: validate clusterID

	k := &Kubernetes{
		clusterID: clusterID,
	}

	k.etcdKubernetesPKI = NewPKI(k, "etcd-k8s")
	k.etcdOverlayPKI = NewPKI(k, "etcd-overlay")
	k.kubernetesPKI = NewPKI(k, "k8s")

	k.secretsGeneric = NewGeneric(k)

	return k
}

func (k *Kubernetes) backends() []Backend {
	return []Backend{
		k.etcdKubernetesPKI,
		k.etcdOverlayPKI,
		k.kubernetesPKI,
		k.kubernetesPKI,
	}
}

func (k *Kubernetes) Ensure() error {
	var result error
	for _, backend := range k.backends() {
		if err := backend.Ensure(); err != nil {
			result = multierror.Append(result, fmt.Errorf("backend %s: %s", backend.Path(), err))
		}
	}
	return result
}

func (k *Kubernetes) Path() string {
	return k.clusterID
}

func NewPKI(k *Kubernetes, pkiName string) *PKI {
	return &PKI{
		pkiName:    pkiName,
		kubernetes: k,
	}
}

type PKI struct {
	pkiName    string
	kubernetes *Kubernetes
}

func NewGeneric(k *Kubernetes) *Generic {
	return &Generic{kubernetes: k}
}

func (p *PKI) Ensure() error {
	return errors.New("implement me")
}

func (p *PKI) Path() string {
	return filepath.Join(p.kubernetes.Path(), "pki", p.pkiName)
}

type Generic struct {
	kubernetes *Kubernetes
}

func (g *Generic) Ensure() error {
	return errors.New("implement me")
}
func (g *Generic) Path() string {
	return filepath.Join(g.kubernetes.Path(), "generic")
}
