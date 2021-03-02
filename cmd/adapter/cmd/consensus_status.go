package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

// 本文件封装了和共识模块有关的client调用接口, 具体格式为:
// xchain-cli consensus status 当前共识状态

const statusBucket = "$consensus"

type ConsensusStatusCommand struct {
	cli *Cli
	cmd *cobra.Command
}

// NewConsensusCommand new consensus cmd
func NewConsensusStatusCommand(cli *Cli) *cobra.Command {
	c := new(ConsensusStatusCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "status",
		Short: "get consensus status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.getStatus(ctx)
		},
	}
	return c.cmd
}

func (c *ConsensusStatusCommand) getStatus(ctx context.Context) error {
	/*
		client := c.cli.XchainClient()
		req := &pb.CommonIn{
			Header: &pb.Header{
				Logid: utils.GenLogId(),
			},
			Bcname: c.cli.RootOptions.Name,
		}
		_, err := client.GetSystemStatus()
	*/
	return nil
}
