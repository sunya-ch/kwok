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

package binary

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"
)

//go:embed prometheus.yaml.tpl
var prometheusYamlTpl string

var prometheusYamlTemplate = template.Must(template.New("_").Parse(prometheusYamlTpl))

func BuildPrometheus(conf BuildPrometheusConfig) (string, error) {
	buf := bytes.NewBuffer(nil)
	err := prometheusYamlTemplate.Execute(buf, conf)
	if err != nil {
		return "", fmt.Errorf("build prometheus error: %s", err)
	}
	return buf.String(), nil
}

type BuildPrometheusConfig struct {
	ProjectName               string
	SecretPort                bool
	AdminCrtPath              string
	AdminKeyPath              string
	PrometheusPort            uint32
	EtcdPort                  uint32
	KubeApiserverPort         uint32
	KubeControllerManagerPort uint32
	KubeSchedulerPort         uint32
	KwokControllerPort        uint32
}
