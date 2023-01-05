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

package utils

import (
	"fmt"
	"net"
)

var (
	errGetUnusedPort        = fmt.Errorf("unable to get an unused port")
	lastUsedPort     uint32 = 32767
)

// GetUnusedPort returns an unused port
func GetUnusedPort() (uint32, error) {
	for lastUsedPort > 10000 {
		lastUsedPort--
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", lastUsedPort))
		if err != nil {
			continue
		}
		l.Close()
		return lastUsedPort, nil
	}

	return 0, errGetUnusedPort
}
