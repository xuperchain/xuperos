/*
 * Copyright (c) 2021, Baidu.com, Inc. All Rights Reserved.
 */

package main

import "log"

var (
	Version   = ""
	BuildTime = ""
	CommitID  = ""
)

func main() {
	cli := NewCli()
	err := cli.Init()
	if err != nil {
		log.Fatal(err)
	}
	cli.AddCommands(commands)
	cli.Execute()
}
