/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package get

import (
	"fmt"

	"github.com/spf13/cobra"

	"sigs.k8s.io/kwok/pkg/kwokctl/cmd/get/artifacts"
	"sigs.k8s.io/kwok/pkg/kwokctl/cmd/get/clusters"
	"sigs.k8s.io/kwok/pkg/kwokctl/cmd/get/kubeconfig"
	"sigs.k8s.io/kwok/pkg/logger"
)

// NewCommand returns a new cobra.Command for get
func NewCommand(logger logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "get",
		Short: "Gets one of [artifacts, clusters, kubeconfig]",
		Long:  "Gets one of [artifacts, clusters, kubeconfig]",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("subcommand is required")
		},
	}
	// add subcommands
	cmd.AddCommand(clusters.NewCommand(logger))
	cmd.AddCommand(artifacts.NewCommand(logger))
	cmd.AddCommand(kubeconfig.NewCommand(logger))
	return cmd
}
