package tests

import (
	"fmt"
	"os/exec"
)

// MachineConfig structure
type MachineConfig struct {
	VMMem       int
	VMCPUs      int
	MachineName string
	VMNetwork   string
	SSHUser     string
	SSHPass     string
	XOUser      string
	XOPass      string
	XOURL       string
	XOTemplate  string
}

// CreateMachine test new machine creation
func (c MachineConfig) CreateMachine() ([]byte, error) {
	vmMem := c.VMMem
	if vmMem == 0 {
		vmMem = 1024
	}
	vmCPUs := c.VMCPUs
	if vmCPUs == 0 {
		vmCPUs = 1
	}
	args := []string{
		"create",
		"-d",
		"xo",
		"--xo-username",
		c.XOUser,
		"--xo-password",
		c.XOPass,
		"--xo-url",
		c.XOURL,
		"--xo-template",
		c.XOTemplate,
		"--xo-vm-mem",
		fmt.Sprintf("%d", vmMem),
		"--xo-vm-cpus",
		fmt.Sprintf("%d", vmCPUs),
	}

	if c.VMNetwork != "" {
		args = append(args, "--xo-vm-network", c.VMNetwork)
	}

	args = append(args, c.MachineName)

	cmd := exec.Command("docker-machine", args...)

	fmt.Println(cmd.Args)

	return cmd.CombinedOutput()
}

// DeleteMachine test delete machine creation
func (c MachineConfig) DeleteMachine() ([]byte, error) {
	args := []string{
		"rm",
		c.MachineName,
	}
	cmd := exec.Command("docker-machine", args...)
	return cmd.CombinedOutput()
}

// StartMachine test start machine
func (c MachineConfig) StartMachine() ([]byte, error) {
	args := []string{
		"start",
		c.MachineName,
	}
	cmd := exec.Command("docker-machine", args...)
	return cmd.CombinedOutput()
}

// StopMachine test stop machine
func (c MachineConfig) StopMachine() ([]byte, error) {
	args := []string{
		"stop",
		c.MachineName,
	}
	cmd := exec.Command("docker-machine", args...)
	return cmd.CombinedOutput()
}
