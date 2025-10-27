package projects

import (
	"context"
	"io"
	"strings"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	composev2 "github.com/docker/compose/v2/pkg/compose"
)

type Client struct {
	svc       api.Compose
	dockerCli command.Cli
}

func NewClient(ctx context.Context) (*Client, error) {
	cli, err := command.NewDockerCli()
	if err != nil {
		return nil, err
	}
	opts := flags.NewClientOptions()
	if err := cli.Initialize(opts); err != nil {
		return nil, err
	}
	svc := composev2.NewComposeService(cli)
	return &Client{svc: svc, dockerCli: cli}, nil
}

func (c *Client) Close() error {
	if c == nil || c.dockerCli == nil {
		return nil
	}
	if apiClient := c.dockerCli.Client(); apiClient != nil {
		_ = apiClient.Close()
	}
	return nil
}

type writerConsumer struct{ out io.Writer }

func (w writerConsumer) Register(container string)    {}
func (w writerConsumer) Start(container string)       {}
func (w writerConsumer) Stop(container string)        {}
func (w writerConsumer) Status(container, msg string) {}
func (w writerConsumer) Log(container, msg string) {
	if w.out == nil {
		return
	}
	// Include container/service name in the format expected by NormalizeProjectLine
	output := msg
	if container != "" {
		output = container + " | " + msg
	}
	if !strings.HasSuffix(output, "\n") {
		output += "\n"
	}
	_, _ = io.WriteString(w.out, output)
}
func (w writerConsumer) Err(container, msg string) {
	if w.out == nil {
		return
	}
	// Include container/service name in the format expected by NormalizeProjectLine
	output := msg
	if container != "" {
		output = container + " | " + msg
	}
	if !strings.HasSuffix(output, "\n") {
		output += "\n"
	}
	_, _ = io.WriteString(w.out, output)
}
