package docker

import (
	"context"
	"emperror.dev/errors"
	"github.com/docker/docker/api/types"
	"github.com/pelican-dev/wings/environment"
)

func CreateNetwork(name string, config types.NetworkCreate) (string, error) {
	cli, err := environment.Docker()
	if err != nil {
		return "", err
	}

	network, err := cli.NetworkCreate(context.Background(), name, config)
	if err != nil {
		return "", errors.Wrap(err, "network/docker: failed to create network")
	}

	return network.ID, nil
}

func RemoveNetwork(networkID string) error {
	cli, err := environment.Docker()
	if err != nil {
		return err
	}

	err = cli.NetworkRemove(context.Background(), networkID)
	if err != nil {
		return errors.Wrap(err, "network/docker: failed to remove network")
	}

	return nil
}
