// Copyright Jetstack Ltd. See LICENSE for details.
package main

import (
	"github.com/jetstack/vault-helper/cmd"
	"github.com/jetstack/vault-helper/pkg/kubernetes"
)

var (
	version string = "dev"
	commit  string = "unknown"
	date    string = ""
)

func main() {
	cmd.Version.Version = version
	cmd.Version.Commit = commit
	cmd.Version.BuildDate = date
	cmd.Execute()
	kubernetes.Version = version
}
