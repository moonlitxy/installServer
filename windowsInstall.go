package main

import (
	"fmt"
	"os/exec"
)

func WindowServerInstall(index string, serverName string, exeName string) error {

	var command string

	//根据输入序号和服务名称创建字符串
	switch index {
	case "1":
		command = fmt.Sprintf("nssm install %s %s", serverName, exeName)
	case "2":
		command = fmt.Sprintf("nssm remove %s confirm", serverName)
	case "3":
		command = fmt.Sprintf("net start %s", serverName)
	case "4":
		command = fmt.Sprintf("nssm stop %s", serverName)
	}
	if command != "" {
		comm := exec.Command("cmd", "/C", command)
		err := comm.Run()
		if err != nil {
			return err
		}
	}
	return nil
}
