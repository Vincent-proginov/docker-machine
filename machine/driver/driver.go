package driver

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	log "github.com/sirupsen/logrus"
	xoclient "github.com/vatesfr/xenorchestra-go-sdk/client"
)

const (
	defaultVMMem = 2048
	defaultVCPUs = 2
)

type Driver struct {
	*drivers.BaseDriver
	XOURL      string
	XOUsername string
	XOPassword string
	XOInsecure bool

	// VM Specs
	VMTempl     string
	VMMem       int
	VMCPUs      int
	CloudConfig string
	VMNetwork   string

	// Internal
	VMID string
}

func NewDriver(hostname, storePath string) *Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostname,
			StorePath:   storePath,
		},
	}
}

func (d *Driver) DriverName() string {
	return "Xen Orchestra"
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:   "xo-url",
			Usage:  "Xen Orchestra URL (e.g. https://xen-orchestra.com)",
			EnvVar: "XO_URL",
		},
		mcnflag.StringFlag{
			Name:   "xo-username",
			Usage:  "Xen Orchestra Username",
			EnvVar: "XO_USERNAME",
		},
		mcnflag.StringFlag{
			Name:   "xo-password",
			Usage:  "Xen Orchestra Password",
			EnvVar: "XO_PASSWORD",
		},
		mcnflag.BoolFlag{
			Name:   "xo-insecure",
			Usage:  "Skip TLS verification",
			EnvVar: "XO_INSECURE",
		},
		mcnflag.StringFlag{
			Name:   "xo-template",
			Usage:  "Template to clone",
			EnvVar: "XO_TEMPLATE",
		},
		mcnflag.IntFlag{
			Name:   "xo-vm-cpus",
			Usage:  "Number of CPUs",
			Value:  defaultVCPUs,
			EnvVar: "XO_VM_CPUS",
		},
		mcnflag.IntFlag{
			Name:   "xo-vm-mem",
			Usage:  "Memory in MB",
			Value:  defaultVMMem,
			EnvVar: "XO_VM_MEM",
		},
		mcnflag.StringFlag{
			Name:   "xo-cloud-config",
			Usage:  "Cloud-init configuration (user-data)",
			EnvVar: "XO_CLOUD_CONFIG",
		},
		mcnflag.StringFlag{
			Name:   "xo-vm-network",
			Usage:  "Network to attach to (name or UUID)",
			EnvVar: "XO_VM_NETWORK",
		},
	}
}

func (d *Driver) SetConfigFromFlags(opts drivers.DriverOptions) error {
	d.XOURL = opts.String("xo-url")
	d.XOUsername = opts.String("xo-username")
	d.XOPassword = opts.String("xo-password")
	d.XOInsecure = opts.Bool("xo-insecure")
	d.VMTempl = opts.String("xo-template")
	d.VMCPUs = opts.Int("xo-vm-cpus")
	d.VMMem = opts.Int("xo-vm-mem")
	d.CloudConfig = opts.String("xo-cloud-config")
	d.VMNetwork = opts.String("xo-vm-network")

	if d.XOURL == "" || d.XOUsername == "" || d.XOPassword == "" || d.VMTempl == "" {
		return fmt.Errorf("xo-url, xo-username, xo-password, and xo-template are required")
	}

	return nil
}

func (d *Driver) getClient() (*xoclient.Client, error) {
	clientInt, err := xoclient.NewClient(xoclient.Config{
		Url:                d.XOURL,
		Username:           d.XOUsername,
		Password:           d.XOPassword,
		InsecureSkipVerify: d.XOInsecure,
		RetryMaxTime:       10 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	client, ok := clientInt.(*xoclient.Client)
	if !ok {
		return nil, fmt.Errorf("failed to cast XO client to *Client")
	}
	return client, nil
}

func (d *Driver) Create() error {
	log.Infof("Creating VM from template %s...", d.VMTempl)

	// Generate SSH Key
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	client, err := d.getClient()
	if err != nil {
		return err
	}

	// Prepare VM request
	vmReq := xoclient.Vm{
		NameLabel:       d.GetMachineName(),
		NameDescription: "Created by docker-machine-driver-xo",
		Template:        d.VMTempl,
		CPUs: xoclient.CPUs{
			Number: d.VMCPUs,
			Max:    d.VMCPUs,
		},
		Memory: xoclient.MemoryObject{
			Static: []int{0, d.VMMem * 1024 * 1024},
		},
	}

	if d.CloudConfig != "" {
		vmReq.CloudConfig = d.CloudConfig
	} else {
		// Generate default cloud-config
		pubKey, err := os.ReadFile(d.GetSSHKeyPath() + ".pub")
		if err != nil {
			return err
		}
		vmReq.CloudConfig = fmt.Sprintf("#cloud-config\nssh_authorized_keys:\n  - %s\n", string(pubKey))
	}

	if d.VMNetwork != "" {
		netObj, err := client.GetNetwork(xoclient.Network{NameLabel: d.VMNetwork})
		if err != nil {
			// Try as UUID
			netObj, err = client.GetNetwork(xoclient.Network{Id: d.VMNetwork})
			if err != nil {
				return fmt.Errorf("network %s not found: %v", d.VMNetwork, err)
			}
		}

		vmReq.VIFsMap = []map[string]string{
			{"network": netObj.Id},
		}
	}

	log.Infof("Creating VM...")
	// CreateVm returns *Vm
	vm, err := client.CreateVm(vmReq, 5*time.Minute)
	if err != nil {
		return err
	}
	d.VMID = vm.Id
	log.Infof("VM Created. UUID: %s", vm.Id)

	// Explicit start to be sure
	if vm.PowerState != "Running" {
		log.Infof("Starting VM...")
		if err := client.StartVm(vm.Id); err != nil {
			return err
		}
	}

	// Wait for IP
	log.Infof("Waiting for IP...")
	// Poll loop
	for i := 0; i < 60; i++ {
		vmState, err := client.GetVm(xoclient.Vm{Id: d.VMID})
		if err != nil {
			return err
		}

		// Check addresses
		for _, addr := range vmState.Addresses {
			// Check if it is an IPv4 address (simple check)
			if strings.Contains(addr, ".") && !strings.Contains(addr, ":") {
				d.IPAddress = addr
				log.Infof("Got IP: %s", addr)
				return nil
			}
		}
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("timeout waiting for IP")
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, "2376")), nil
}

func (d *Driver) GetState() (state.State, error) {
	client, err := d.getClient()
	if err != nil {
		return state.Error, err
	}

	vm, err := client.GetVm(xoclient.Vm{Id: d.VMID})
	if err != nil {
		return state.Error, err
	}

	switch vm.PowerState {
	case "Running":
		return state.Running, nil
	case "Halted":
		return state.Stopped, nil
	case "Paused":
		return state.Paused, nil
	default:
		return state.None, nil
	}
}

func (d *Driver) Start() error {
	client, err := d.getClient()
	if err != nil {
		return err
	}
	return client.StartVm(d.VMID)
}

func (d *Driver) Stop() error {
	client, err := d.getClient()
	if err != nil {
		return err
	}
	return client.HaltVm(d.VMID)
}

func (d *Driver) Kill() error {
	client, err := d.getClient()
	if err != nil {
		return err
	}
	// Force stop? The SDK has HaltVm (clean), but maybe stop with force?
	// The SDK HaltVm calls `changeVmState(id, "stop", ...)`
	// `stop` without force is usually clean?
	// The SDK also has `StopVm`?
	// looking at client/vm.go previously:
	// func (c *Client) HaltVm(id string) error { return c.changeVmState(id, "stop", ...) }
	// There might be a hard stop.
	// But `vm.stop` usually takes params?
	// Let's just use HaltVm for now or try to call custom.
	// `vm.stop` with `{ "force": true }`
	params := map[string]interface{}{
		"id":    d.VMID,
		"force": true,
	}
	var res interface{}
	return client.Call("vm.stop", params, &res)
}

func (d *Driver) Remove() error {
	client, err := d.getClient()
	if err != nil {
		return err
	}
	return client.DeleteVm(d.VMID)
}

func (d *Driver) Restart() error {
	client, err := d.getClient()
	if err != nil {
		return err
	}
	// vm.restart
	params := map[string]interface{}{
		"id": d.VMID,
	}
	var res interface{}
	return client.Call("vm.restart", params, &res)
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}
