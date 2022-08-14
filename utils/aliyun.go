package utils

import (
	"context"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/go-redis/redis/v8"
)

func AddContainer(port int, name string, ctx context.Context, rdb *redis.Client) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	resp, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Image: "928567a67d84",
			Cmd:   []string{},
		},
		&container.HostConfig{PortBindings: map[nat.Port][]nat.PortBinding{"8080/tcp": {{"0.0.0.0", strconv.Itoa(port)}}}}, nil, nil, name)

	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	return nil
}

func DelContainerByName(name string, ctx context.Context, rdb *redis.Client) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	err = cli.ContainerRemove(ctx, name, types.ContainerRemoveOptions{Force: true})
	if err != nil {
		panic(err)
	}
	DelPort(name, rdb, ctx)
}
