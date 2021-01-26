/*
 * Copyright (c) 2021, Baidu.com, Inc. All Rights Reserved.
 */

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/xuperchain/xupercore/bcs/ledger/xledger/utils"
	"github.com/xuperchain/xupercore/kernel/common/xconfig"
	"github.com/xuperchain/xupercore/lib/logs"
	_ "github.com/xuperchain/xupercore/lib/storage/kvdb/leveldb"
	xutils "github.com/xuperchain/xupercore/lib/utils"

	"github.com/spf13/cobra"
)

// CreateChainCommand create chain cmd
type CreateChainCommand struct {
	cli *Cli
	cmd *cobra.Command
	// 创世块配置文件
	GenesisConf string
	// 环境配置文件
	EnvConf string
}

// NewCreateChainVersion new create chain cmd
func NewCreateChainVersion(cli *Cli) *cobra.Command {
	c := new(CreateChainCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "createChain",
		Short: "Operate a blockchain: [OPTIONS].",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.createChain(ctx)
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *CreateChainCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.GenesisConf,
		"genesis_conf", "g", "./data/genesis/xuper.json", "genesis config file path")
	c.cmd.Flags().StringVarP(&c.EnvConf,
		"env_conf", "e", "./conf/env.yaml", "env config file path")
}

func (c *CreateChainCommand) createChain(ctx context.Context) error {
	log.Printf("start create chain.bc_name:%s genesis_conf:%s env_conf:%s\n",
		c.cli.RootOptions.Name, c.GenesisConf, c.EnvConf)

	if !xutils.FileIsExist(c.GenesisConf) || !xutils.FileIsExist(c.EnvConf) {
		log.Printf("config file not exist.genesis_conf:%s env_conf:%s\n", c.GenesisConf, c.EnvConf)
		return fmt.Errorf("config file not exist")
	}

	econf, err := xconfig.LoadEnvConf(c.EnvConf)
	if err != nil {
		log.Printf("load env config failed.env_conf:%s err:%v\n", c.EnvConf, err)
		return fmt.Errorf("load env config failed")
	}

	logs.InitLog(econf.GenConfFilePath(econf.LogConf), econf.GenDirAbsPath(econf.LogDir))
	err = utils.CreateLedger(c.cli.RootOptions.Name, c.GenesisConf, econf)
	if err != nil {
		log.Printf("create ledger failed.err:%v\n", err)
		return fmt.Errorf("create ledger failed")
	}

	log.Printf("create ledger succ.bc_name:%s\n", c.cli.RootOptions.Name)
	return nil
}

func init() {
	AddCommand(NewCreateChainVersion)
}
