//go:build windows

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
	"context"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

func startProcess(ctx context.Context, name string, arg ...string) *exec.Cmd {
	cmd := command(ctx, name, arg...)
	cmd.SysProcAttr.CreationFlags |= windows.CREATE_NEW_CONSOLE
	return cmd
}

func command(ctx context.Context, name string, arg ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
	return cmd
}

func isRunning(pid int) bool {
	_, err := os.FindProcess(pid)
	return err == nil
}
