package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/moonlitxy/installServer/filebase"
)

func LinuxSererInstall(index string, serverName string) error {

	var command string
	var err error

	// 检查系统是否使用systemd
	_, systemdCheck := exec.LookPath("systemctl")

	//根据输入序号和服务名称创建字符串
	switch index {
	case "1": //注册服务
		// 如果系统使用systemd，服务已经在main.go中创建并启用，无需额外操作
		if systemdCheck == nil {
			// 检查服务文件是否存在
			serviceFile := fmt.Sprintf("/etc/systemd/system/%s.service", serverName)
			if _, err := os.Stat(serviceFile); err != nil {
				return fmt.Errorf("systemd服务文件不存在: %v", err)
			}

			// 重新加载systemd配置
			reloadCmd := exec.Command("systemctl", "daemon-reload")
			err = reloadCmd.Run()
			if err != nil {
				fmt.Printf("警告: systemctl daemon-reload失败: %v\n", err)
			}

			// 启动服务
			startCmd := exec.Command("systemctl", "start", serverName)
			err = startCmd.Run()
			if err != nil {
				return fmt.Errorf("启动systemd服务失败: %v", err)
			}

			return nil
		}

		// 非systemd系统，使用传统方式
		// 检查make命令是否存在
		_, err = exec.LookPath("make")
		if err != nil {
			// 如果make不存在，尝试直接执行安装命令
			command = fmt.Sprintf("mv ./%s /etc/init.d/%s && chmod +x /etc/init.d/%s", serverName, serverName, serverName)
			comm := exec.Command("/bin/sh", "-c", command)
			err = comm.Run()
			if err != nil {
				return fmt.Errorf("执行安装命令失败: %v", err)
			}

			// 创建符号链接
			links := []string{
				fmt.Sprintf("ln -s /etc/init.d/%s /etc/rc1.d/S20%s", serverName, serverName),
				fmt.Sprintf("ln -s /etc/init.d/%s /etc/rc2.d/S20%s", serverName, serverName),
				fmt.Sprintf("ln -s /etc/init.d/%s /etc/rc3.d/S20%s", serverName, serverName),
				fmt.Sprintf("ln -s /etc/init.d/%s /etc/rc4.d/S20%s", serverName, serverName),
				fmt.Sprintf("ln -s /etc/init.d/%s /etc/rc5.d/S20%s", serverName, serverName),
				fmt.Sprintf("ln -s /etc/init.d/%s /etc/rc0.d/S60%s", serverName, serverName),
				fmt.Sprintf("ln -s /etc/init.d/%s /etc/rc6.d/S60%s", serverName, serverName),
			}

			for _, linkCmd := range links {
				comm := exec.Command("/bin/sh", "-c", linkCmd)
				comm.Run() // 忽略错误，因为某些链接可能已存在
			}

			// 为Ubuntu 22.04添加systemctl兼容支持
			comm = exec.Command("/bin/sh", "-c", "systemctl daemon-reload 2>/dev/null || true")
			comm.Run()
		} else {
			// 如果make存在，使用make命令
			command = fmt.Sprintf("make")
			comm := exec.Command("/bin/sh", "-c", command)
			err = comm.Run()
			if err != nil {
				return fmt.Errorf("执行make命令失败: %v", err)
			}
		}
	case "2": //删除服务
		if systemdCheck == nil {
			// 使用systemd删除服务
			fmt.Println("检测到systemd，使用systemd方式删除服务")

			// 停止服务
			stopCmd := exec.Command("systemctl", "stop", serverName)
			err := stopCmd.Run()
			if err != nil {
				fmt.Printf("停止服务失败: %v\n", err)
				// 继续执行删除操作，即使停止失败
			}

			// 禁用服务
			disableCmd := exec.Command("systemctl", "disable", serverName)
			err = disableCmd.Run()
			if err != nil {
				fmt.Printf("禁用服务失败: %v\n", err)
				// 继续执行删除操作，即使禁用失败
			}

			// 删除服务文件
			serviceFile := fmt.Sprintf("/etc/systemd/system/%s.service", serverName)
			rmCmd := exec.Command("rm", "-f", serviceFile)
			err = rmCmd.Run()
			if err != nil {
				return fmt.Errorf("删除服务文件失败: %v", err)
			}

			// 重新加载systemd配置
			reloadCmd := exec.Command("systemctl", "daemon-reload")
			err = reloadCmd.Run()
			if err != nil {
				fmt.Printf("重新加载systemd配置失败: %v\n", err)
				// 继续执行，即使重新加载失败
			}
		} else {
			// 使用传统sysvinit方式删除服务
			fmt.Println("未检测到systemd，使用传统sysvinit方式删除服务")

			// 停止服务
			stopCmd := exec.Command("service", serverName, "stop")
			err := stopCmd.Run()
			if err != nil {
				fmt.Printf("停止服务失败: %v\n", err)
				// 继续执行删除操作，即使停止失败
			}

			// 删除服务
			removeServiceAll(serverName)

			// 删除/etc/init.d/下的服务文件
			initDPath := "/etc/init.d/" + serverName
			rmCmd := exec.Command("rm", "-f", initDPath)
			err = rmCmd.Run()
			if err != nil {
				fmt.Printf("删除/etc/init.d/服务文件失败: %v\n", err)
				// 继续执行，即使删除失败
			}
		}
	case "3": //启动服务
		if systemdCheck == nil {
			// 使用systemctl启动服务
			command = fmt.Sprintf("systemctl start %s", serverName)
		} else {
			// 使用service启动服务
			command = fmt.Sprintf("service %s start", serverName)
		}
		comm := exec.Command("/bin/sh", "-c", command)
		err = comm.Run()
		if err != nil {
			return fmt.Errorf("启动服务失败: %v", err)
		}
	case "4": //停止服务
		if systemdCheck == nil {
			// 使用systemctl停止服务
			command = fmt.Sprintf("systemctl stop %s", serverName)
		} else {
			// 使用service停止服务
			command = fmt.Sprintf("service %s stop", serverName)
		}
		comm := exec.Command("/bin/sh", "-c", command)
		err = comm.Run()
		if err != nil {
			return fmt.Errorf("停止服务失败: %v", err)
		}
	}
	return nil
}

/*
读取服务状态
修复以支持Ubuntu 22.04和CentOS 8的systemd服务管理
*/
func ReadServerStat(serverName string) string {
	var command string
	var buff bytes.Buffer

	// 检查系统是否使用systemd
	_, systemdCheck := exec.LookPath("systemctl")

	if systemdCheck == nil {
		// 使用systemd方式获取服务状态
		command = fmt.Sprintf("systemctl status %s | grep 'Active:' | awk '{print $2}'", serverName)
		comm := exec.Command("/bin/sh", "-c", command)
		comm.Stdout = &buff
		err := comm.Run()

		if err == nil {
			status, _ := buff.ReadString('\n')
			status = strings.TrimSpace(status)
			if status == "active" {
				// 服务正在运行，获取PID
				buff.Reset()
				command = fmt.Sprintf("systemctl status %s | grep 'Main PID:' | awk '{print $3}'", serverName)
				comm := exec.Command("/bin/sh", "-c", command)
				comm.Stdout = &buff
				comm.Run()

				pid, _ := buff.ReadString('\n')
				return strings.TrimSpace(pid)
			}
		}

		// 如果systemctl方式失败，尝试使用ps命令查找进程
		buff.Reset()
		command = fmt.Sprintf("ps -ef|grep %s |grep -v grep|awk '{print $2}'", serverName)
		comm = exec.Command("/bin/sh", "-c", command)
		comm.Stdout = &buff
		comm.Run()
	} else {
		// 非systemd系统，使用传统方式
		// 首先尝试使用status命令（sysvinit兼容）
		command = fmt.Sprintf("service %s status || service %s stat", serverName, serverName)
		comm := exec.Command("/bin/sh", "-c", command)
		comm.Stdout = &buff
		err := comm.Run()

		// 如果失败，尝试直接使用ps命令查找进程
		if err != nil {
			buff.Reset()
			command = fmt.Sprintf("ps -ef|grep /etc/init.d/%s |grep -v grep|awk '{print $2}'", serverName)
			comm = exec.Command("/bin/sh", "-c", command)
			comm.Stdout = &buff
			comm.Run()
		}
	}

	line, errs := buff.ReadString('\n')
	if errs != nil {
		return ""
	}
	return strings.TrimSpace(line)
}

/*
删除快捷方式
*/
func removeServiceAll(serverName string) {
	var command = new([7]string)
	command[0] = fmt.Sprintf("/etc/rc1.d/S20%s", serverName)
	command[1] = fmt.Sprintf("/etc/rc2.d/S20%s", serverName)
	command[2] = fmt.Sprintf("/etc/rc3.d/S20%s", serverName)
	command[3] = fmt.Sprintf("/etc/rc4.d/S20%s", serverName)
	command[4] = fmt.Sprintf("/etc/rc5.d/S20%s", serverName)
	command[5] = fmt.Sprintf("/etc/rc0.d/S60%s", serverName)
	command[6] = fmt.Sprintf("/etc/rc6.d/S60%s", serverName)

	length := len(command)
	for i := 0; i < length; i++ {
		exec.Command("/bin/sh", "-c", "rm -r "+command[i]).Run()
	}
}

/*
创建service文件
*/
func CreateShellService(filePath string, exeName string, fileName string) error {
	var fData bytes.Buffer
	fileAddr := fmt.Sprintf("%s%s", filePath, exeName)
	// 修复脚本格式，移除case语句中的冒号错误
	fData.WriteString("#!/bin/sh\n\n")
	fData.WriteString("ARG=$1\n\n")
	fData.WriteString("case $ARG in\n")
	fData.WriteString("start)\n")
	fData.WriteString(fmt.Sprintf("cd  %s \n", filePath))
	fData.WriteString(fmt.Sprintf("nohup  %s >/dev/null 2>&1 &\n", fileAddr))
	fData.WriteString(";;\n")
	fData.WriteString("stop)\n")
	fData.WriteString(fmt.Sprintf("pid=`ps -ef|grep %s |grep -v grep|awk '{print $2}'`\n", fileAddr))
	fData.WriteString("sleep 3\n")
	fData.WriteString("if [ ! -z \"$pid\" ]; then\n")
	fData.WriteString("  kill -9 $pid\n")
	fData.WriteString("fi\n")
	fData.WriteString(";;\n")
	fData.WriteString("restart)\n")
	fData.WriteString(fmt.Sprintf("pid=`ps -ef|grep %s |grep -v grep|awk '{print $2}'`\n", fileAddr))
	fData.WriteString("sleep 3\n")
	fData.WriteString("if [ ! -z \"$pid\" ]; then\n")
	fData.WriteString("  kill -9 $pid\n")
	fData.WriteString("fi\n")
	fData.WriteString("sleep 3\n")
	fData.WriteString(fmt.Sprintf("cd  %s \n", filePath))
	fData.WriteString(fmt.Sprintf("nohup  %s >/dev/null 2>&1 &\n", fileAddr))
	fData.WriteString(";;\n")
	fData.WriteString("stat)\n")
	fData.WriteString(fmt.Sprintf("pid=`ps -ef|grep %s |grep -v grep|awk '{print $2}'`\n", fileAddr))
	fData.WriteString("sleep 1\n")
	fData.WriteString("echo $pid\n")
	fData.WriteString(";;\n")
	// 添加status命令以兼容systemctl
	fData.WriteString("status)\n")
	fData.WriteString(fmt.Sprintf("pid=`ps -ef|grep %s |grep -v grep|awk '{print $2}'`\n", fileAddr))
	fData.WriteString("sleep 1\n")
	fData.WriteString("echo $pid\n")
	fData.WriteString(";;\n")
	fData.WriteString("esac")
	filebase.WriteDataByte(fileName, fData.Bytes())

	// 添加执行权限
	os.Chmod(fileName, 0755)

	return nil
}

/*
创建javaservice文件
*/
func CreateJavaService(filePath string, serverName string, fileName string) error {
	var fData bytes.Buffer

	// 修复脚本格式，移除case语句中的冒号错误
	fData.WriteString("#!/bin/sh\n\n")
	fData.WriteString("ARG=$1\n\n")
	fData.WriteString("case $ARG in\n")
	fData.WriteString("start)\n")
	fData.WriteString(fmt.Sprintf("export currpath=\"%s\"\n", filePath))
	fData.WriteString(fmt.Sprintf("export CLASSPATH=$CLASSPATH:$currpath\n"))
	fData.WriteString(fmt.Sprintf("for jar in $currpath/lib/*.jar\n"))
	fData.WriteString(fmt.Sprintf("do\n"))
	fData.WriteString(fmt.Sprintf("      CLASSPATH=$CLASSPATH:$jar\n"))
	fData.WriteString(fmt.Sprintf("done\n"))
	fData.WriteString(fmt.Sprintf("export CLASSPATH\n"))
	fData.WriteString(fmt.Sprintf("java -cp $CLASSPATH %s &\n", serverName))
	fData.WriteString(";;\n")
	fData.WriteString("stop)\n")
	fData.WriteString(fmt.Sprintf("pid=`ps -ef|grep %s |grep -v grep|awk '{print $2}'`\n", serverName))
	fData.WriteString("if [ ! -z \"$pid\" ]\n")
	fData.WriteString("then\n")
	fData.WriteString("    kill -9 $pid\n")
	fData.WriteString("    sleep 3\n")
	fData.WriteString("fi\n")
	fData.WriteString(";;\n")
	fData.WriteString("restart)\n")
	fData.WriteString(fmt.Sprintf("pid=`ps -ef|grep %s |grep -v grep|awk '{print $2}'`\n", serverName))
	fData.WriteString("if [ ! -z \"$pid\" ]\n")
	fData.WriteString("then\n")
	fData.WriteString("    kill -9 $pid\n")
	fData.WriteString("    sleep 3\n")
	fData.WriteString("fi\n")
	fData.WriteString(fmt.Sprintf("export currpath=\"%s\"\n", filePath))
	fData.WriteString(fmt.Sprintf("export CLASSPATH=$CLASSPATH:$currpath\n"))
	fData.WriteString(fmt.Sprintf("for jar in $currpath/lib/*.jar\n"))
	fData.WriteString(fmt.Sprintf("do\n"))
	fData.WriteString(fmt.Sprintf("      CLASSPATH=$CLASSPATH:$jar\n"))
	fData.WriteString(fmt.Sprintf("done\n"))
	fData.WriteString(fmt.Sprintf("export CLASSPATH\n"))
	fData.WriteString(fmt.Sprintf("java -cp $CLASSPATH %s &\n", serverName))
	fData.WriteString(";;\n")
	fData.WriteString("stat)\n")
	fData.WriteString(fmt.Sprintf("pid=`ps -ef|grep %s |grep -v grep|awk '{print $2}'`\n", serverName))
	fData.WriteString("    sleep 3\n")
	fData.WriteString("if [ ! -z \"$pid\" ]\n")
	fData.WriteString("then\n")
	fData.WriteString("echo $pid\n")
	fData.WriteString("fi\n")
	fData.WriteString(";;\n")
	// 添加status命令以兼容systemctl
	fData.WriteString("status)\n")
	fData.WriteString(fmt.Sprintf("pid=`ps -ef|grep %s |grep -v grep|awk '{print $2}'`\n", serverName))
	fData.WriteString("    sleep 3\n")
	fData.WriteString("if [ ! -z \"$pid\" ]\n")
	fData.WriteString("then\n")
	fData.WriteString("echo $pid\n")
	fData.WriteString("fi\n")
	fData.WriteString(";;\n")
	fData.WriteString("esac")
	filebase.WriteDataByte(fileName, fData.Bytes())

	// 添加执行权限
	os.Chmod(fileName, 0755)

	return nil
}

/*
创建systemd服务文件
修复以支持CentOS 8和Ubuntu 22.04
*/
func CreateSystemdService(filePath string, exeName string, serverName string) error {
	var fData bytes.Buffer
	fileAddr := fmt.Sprintf("%s%s", filePath, exeName)

	// 创建systemd服务文件
	fData.WriteString("[Unit]\n")
	fData.WriteString(fmt.Sprintf("Description=%s Service\n", serverName))
	fData.WriteString(fmt.Sprintf("Documentation=man:%s(8)\n", serverName))
	fData.WriteString("After=network.target\n\n")

	fData.WriteString("[Service]\n")
	fData.WriteString("Type=simple\n")
	fData.WriteString(fmt.Sprintf("ExecStart=%s\n", fileAddr))
	fData.WriteString("ExecReload=/bin/kill -HUP $MAINPID\n")
	fData.WriteString("Restart=always\n")
	fData.WriteString("RestartSec=10\n")
	fData.WriteString("User=root\n")
	fData.WriteString("Group=root\n")
	fData.WriteString("WorkingDirectory=" + filePath + "\n")
	fData.WriteString("Environment=PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\n\n")

	fData.WriteString("[Install]\n")
	fData.WriteString("WantedBy=multi-user.target\n")

	// 写入systemd服务文件
	serviceFileName := fmt.Sprintf("/etc/systemd/system/%s.service", serverName)
	ok := filebase.WriteDataByte(serviceFileName, fData.Bytes())
	if !ok {
		return fmt.Errorf("写入systemd服务文件失败")
	}

	// 设置文件权限
	err := os.Chmod(serviceFileName, 0644)
	if err != nil {
		return fmt.Errorf("设置systemd服务文件权限失败: %v", err)
	}

	// 重新加载systemd配置
	reloadCmd := exec.Command("systemctl", "daemon-reload")
	err = reloadCmd.Run()
	if err != nil {
		return fmt.Errorf("systemctl daemon-reload失败: %v", err)
	}

	return nil
}

/*
创建makefile文件
*/
func CrrateMakefile(serverName string) error {
	var fData bytes.Buffer

	// 修复Makefile，解决目录切换问题
	// 使用 && 确保命令顺序执行，或者在同一行执行多个命令
	fData.WriteString("install:\n\n")
	// 移动文件并设置执行权限
	fData.WriteString(fmt.Sprintf("\tmv ./%s /etc/init.d/%s && chmod +x /etc/init.d/%s\n\n", serverName, serverName, serverName))
	// 使用 && 连接命令，确保在同一shell上下文中执行
	fData.WriteString(fmt.Sprintf("\tcd /etc/init.d/ && \\\n"))
	fData.WriteString(fmt.Sprintf("\tln -s /etc/init.d/%s /etc/rc1.d/S20%s && \\\n", serverName, serverName))
	fData.WriteString(fmt.Sprintf("\tln -s /etc/init.d/%s /etc/rc2.d/S20%s && \\\n", serverName, serverName))
	fData.WriteString(fmt.Sprintf("\tln -s /etc/init.d/%s /etc/rc3.d/S20%s && \\\n", serverName, serverName))
	fData.WriteString(fmt.Sprintf("\tln -s /etc/init.d/%s /etc/rc4.d/S20%s && \\\n", serverName, serverName))
	fData.WriteString(fmt.Sprintf("\tln -s /etc/init.d/%s /etc/rc5.d/S20%s && \\\n", serverName, serverName))
	fData.WriteString(fmt.Sprintf("\tln -s /etc/init.d/%s /etc/rc0.d/S60%s && \\\n", serverName, serverName))
	fData.WriteString(fmt.Sprintf("\tln -s /etc/init.d/%s /etc/rc6.d/S60%s\n\n", serverName, serverName))

	// 为Ubuntu 22.04添加systemctl兼容支持
	fData.WriteString("\t# 为systemctl添加兼容性支持\n")
	fData.WriteString("\tsystemctl daemon-reload 2>/dev/null || true\n\n")

	fData.WriteString("\techo \"control complete\"")
	filebase.WriteDataByte("makefile", fData.Bytes())

	return nil
}
