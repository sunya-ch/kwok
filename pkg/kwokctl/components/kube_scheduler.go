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

package components

import (
	"sigs.k8s.io/kwok/pkg/apis/internalversion"
	"sigs.k8s.io/kwok/pkg/utils/format"
)

type BuildKubeSchedulerComponentConfig struct {
	Binary           string
	Image            string
	Workdir          string
	Address          string
	Port             uint32
	SecurePort       bool
	CaCertPath       string
	AdminCertPath    string
	AdminKeyPath     string
	KubeconfigPath   string
	KubeFeatureGates string
}

func BuildKubeSchedulerComponent(conf BuildKubeSchedulerComponentConfig) (component internalversion.Component, err error) {
	if conf.Address == "" {
		conf.Address = publicAddress
	}

	kubeSchedulerArgs := []string{}

	if conf.KubeFeatureGates != "" {
		kubeSchedulerArgs = append(kubeSchedulerArgs,
			"--feature-gates="+conf.KubeFeatureGates,
		)
	}

	inContainer := conf.Image != ""
	var volumes []internalversion.Volume

	if inContainer {
		volumes = append(volumes,
			internalversion.Volume{
				HostPath:  conf.KubeconfigPath,
				MountPath: "/root/.kube/config",
				ReadOnly:  true,
			},
			internalversion.Volume{
				HostPath:  conf.AdminCertPath,
				MountPath: "/etc/kubernetes/pki/admin.crt",
				ReadOnly:  true,
			},
			internalversion.Volume{
				HostPath:  conf.AdminKeyPath,
				MountPath: "/etc/kubernetes/pki/admin.key",
				ReadOnly:  true,
			},
		)
		kubeSchedulerArgs = append(kubeSchedulerArgs,
			"--kubeconfig=/root/.kube/config",
		)
	} else {
		kubeSchedulerArgs = append(kubeSchedulerArgs,
			"--kubeconfig="+conf.KubeconfigPath,
		)
	}

	if conf.SecurePort {
		kubeSchedulerArgs = append(kubeSchedulerArgs,
			"--authorization-always-allow-paths=/healthz,/readyz,/livez,/metrics",
		)

		if inContainer {
			kubeSchedulerArgs = append(kubeSchedulerArgs,
				"--bind-address="+publicAddress,
				"--secure-port=10259",
			)
		} else {
			kubeSchedulerArgs = append(kubeSchedulerArgs,
				"--bind-address="+conf.Address,
				"--secure-port="+format.String(conf.Port),
			)
		}

		// TODO: Support disable insecure port
		//	kubeSchedulerArgs = append(kubeSchedulerArgs,
		//		"--port=0",
		//	)
	} else {
		if inContainer {
			kubeSchedulerArgs = append(kubeSchedulerArgs,
				"--address="+publicAddress,
				"--port=10251",
			)
		} else {
			kubeSchedulerArgs = append(kubeSchedulerArgs,
				"--address="+conf.Address,
				"--port="+format.String(conf.Port),
			)
		}

		// TODO: Support disable secure port
		//	kubeSchedulerArgs = append(kubeSchedulerArgs,
		//		"--secure-port=0",
		//	)
	}

	return internalversion.Component{
		Name: "kube-scheduler",
		Links: []string{
			"kube-apiserver",
		},
		Command: []string{"kube-scheduler"},
		Volumes: volumes,
		Args:    kubeSchedulerArgs,
		Binary:  conf.Binary,
		Image:   conf.Image,
		WorkDir: conf.Workdir,
	}, nil
}
