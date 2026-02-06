module github.com/nutanix/docker-machine

go 1.24.0

replace github.com/docker/docker => github.com/docker/engine v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible

require (
	github.com/docker/machine v0.16.2
	github.com/sirupsen/logrus v1.9.3
	github.com/vatesfr/xenorchestra-go-sdk v1.11.0
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/docker/docker v20.10.17+incompatible // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sourcegraph/jsonrpc2 v0.2.1 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/term v0.37.0 // indirect
	gotest.tools v2.2.0+incompatible // indirect
)
