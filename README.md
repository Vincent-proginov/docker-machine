# Xen Orchestra Rancher Node Driver

This repository contains the Rancher Node Driver for Xen Orchestra (XCP-ng). This driver allows you to provision hosts on XCP-ng/XenServer using Xen Orchestra, which Rancher uses to launch and manage Kubernetes clusters.

## Features

- Configure Xen Orchestra connection (URL, User, Password)
- Create VMs from Templates
- Configure VM Resources (vCPUs, Memory)
- Cloud-init support
- Network selection

## Installation

1. From the Home view, choose *Cluster Management* > *Drivers* in the navigation bar. From the Drivers page, select the *Node Drivers* tab.
2. Click *Add Node Driver*.
3. Complete the Add Node Driver form. Then click Create.

## Driver Args

| Arg | Description | Required | Default |
|---|---|---|---|
| `xo-url` | Xen Orchestra URL (e.g. `https://xo.example.com`) | yes | |
| `xo-username` | Xen Orchestra Username | yes | |
| `xo-password` | Xen Orchestra Password | yes | |
| `xo-insecure` | Skip TLS verification | no | `false` |
| `xo-template` | Name of the template to clone | yes | |
| `xo-vm-cpus` | Number of vCPUs | no | `2` |
| `xo-vm-mem` | Memory in MB | no | `2048` |
| `xo-cloud-config` | Cloud-init user-data | no | |
| `xo-vm-network` | Network name or UUID to attach to | no | |

## Development

### Build Instructions

build linux/amd64 binary => `make` 
build local binary => `make local`
