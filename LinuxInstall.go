package main

import (
	"bytes"
	"fmt"
	"installServer/filebase"
	"os"
	"os/exec"
	"strings"
)

// Linux发行版类型
type LinuxDistro int

const (
	DistroUnknown LinuxDistro = iota
	DistroCentOS
	DistroUbuntu
	DistroDebian
)

// 检测Linux发行版
func detectLinuxDistro() LinuxDistro {
	// 读取/etc/os-release文件
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		content := string(data)
		if strings.Contains(content, "CentOS") || strings.Contains(content, "Red Hat") || strings.Contains(content, "Rocky") || strings.Contains(content, "Kylin") {
			return DistroCentOS
		}
		if strings.Contains(content, "Ubuntu") {
			return DistroUbuntu
		}
		if strings.Contains(content, "Debian") {
			return DistroDebian
		}
	}

	// 备用检测方法：检查是否存在特定文件
	if _, err := os.Stat("/etc/redhat-release"); err == nil {
		return DistroCentOS
	}
	if _, err := os.Stat("/etc/debian_version"); err == nil {
		// 进一步区分Ubuntu和Debian
		if data, err := os.ReadFile("/etc/lsb-release"); err == nil {
			if strings.Contains(string(data), "Ubuntu") {
				return DistroUbuntu
			}
		}
		return DistroDebian
	}

	return DistroUnknown
}

func LinuxSererInstall(index string, serverName string) error {

	var command string

	//根据输入序号和服务名称创建字符串
	switch index {
	case "1": //注册服务
		command = "make"
	case "2": //删除服务
		removeServiceAll(serverName)
		command = fmt.Sprintf("rm -f /etc/init.d/%s", serverName)
	case "3": //启动服务
		command = fmt.Sprintf("service %s start", serverName)
	case "4": //停止服务
		command = fmt.Sprintf("service %s stop", serverName)
	}
	if command != "" {
		comm := exec.Command("/bin/sh", "-c", command)
		err := comm.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

/*
读取服务状态
*/
func ReadServerStat(serverName string) string {
	var command string
	var buff bytes.Buffer
	command = fmt.Sprintf("service %s stat", serverName)
	comm := exec.Command("/bin/sh", "-c", command)
	comm.Stdout = &buff
	err := comm.Run()
	if err != nil {
		return ""
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
	distro := detectLinuxDistro()

	// 根据发行版使用不同的删除命令
	switch distro {
	case DistroCentOS:
		// CentOS使用chkconfig删除服务
		exec.Command("/bin/sh", "-c", fmt.Sprintf("chkconfig --del %s 2>/dev/null || true", serverName)).Run()
	case DistroUbuntu, DistroDebian:
		// Ubuntu/Debian使用update-rc.d删除服务
		exec.Command("/bin/sh", "-c", fmt.Sprintf("update-rc.d -f %s remove 2>/dev/null || true", serverName)).Run()
	default:
		// 未知发行版，手动删除符号链接
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
			exec.Command("/bin/sh", "-c", "rm -f "+command[i]).Run()
		}
	}
}

/*
创建service文件
*/
func CreateShellService(filePath string, exeName string, fileName string) error {
	var fData bytes.Buffer
	fileAddr := fmt.Sprintf("%s%s", filePath, exeName)
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
	fData.WriteString("kill -9 $pid\n")
	fData.WriteString(";;\n")
	fData.WriteString("restart)\n")
	fData.WriteString(fmt.Sprintf("pid=`ps -ef|grep %s |grep -v grep|awk '{print $2}'`\n", fileAddr))
	fData.WriteString("sleep 3\n")
	fData.WriteString("kill -9 $pid\n")
	fData.WriteString("sleep 3\n")
	fData.WriteString(fmt.Sprintf("cd  %s \n", filePath))
	fData.WriteString(fmt.Sprintf("nohup  %s >/dev/null 2>&1 &\n", fileAddr))
	fData.WriteString(";;\n")
	fData.WriteString("stat)\n")
	fData.WriteString(fmt.Sprintf("pid=`ps -ef|grep %s |grep -v grep|awk '{print $2}'`\n", fileAddr))
	fData.WriteString("sleep 1\n")
	fData.WriteString("echo $pid\n")
	fData.WriteString(";;\n")
	fData.WriteString("esac")
	filebase.WriteDataByte(fileName, fData.Bytes())

	return nil
}

/*
创建javaservice文件
*/
func CreateJavaService(filePath string, serverName string, fileName string) error {
	var fData bytes.Buffer

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
	fData.WriteString("if [[ $pid -gt 0 ]]\n")
	fData.WriteString("then\n")
	fData.WriteString("    kill -9 $pid\n")
	fData.WriteString("    sleep 3\n")
	fData.WriteString("fi\n")
	fData.WriteString(";;\n")
	fData.WriteString("restart)\n")
	fData.WriteString(fmt.Sprintf("pid=`ps -ef|grep %s |grep -v grep|awk '{print $2}'`\n", serverName))
	fData.WriteString("if [[ $pid -gt 0 ]]\n")
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
	fData.WriteString("if [[ $pid -gt 0 ]]\n")
	fData.WriteString("then\n")
	fData.WriteString("echo $pid\n")
	fData.WriteString("fi\n")
	fData.WriteString(";;\n")
	fData.WriteString("esac")
	filebase.WriteDataByte(fileName, fData.Bytes())

	return nil
}

/*
创建makefile文件
兼容CentOS、Ubuntu、Debian系统
*/
func CrrateMakefile(serverName string) error {
	var fData bytes.Buffer
	distro := detectLinuxDistro()

	fData.WriteString("install:\n")
	fData.WriteString(fmt.Sprintf("	cp ./%s /etc/init.d/%s\n", serverName, serverName))
	fData.WriteString(fmt.Sprintf("	chmod +x /etc/init.d/%s\n", serverName))

	// 根据发行版使用不同的注册方式
	switch distro {
	case DistroCentOS:
		// CentOS使用chkconfig注册服务
		fData.WriteString(fmt.Sprintf("	# CentOS/RHEL系统使用chkconfig\n"))
		fData.WriteString(fmt.Sprintf("	-chkconfig --add %s 2>/dev/null || true\n", serverName))
		fData.WriteString(fmt.Sprintf("	-chkconfig %s on 2>/dev/null || true\n", serverName))
		// 备用方案：手动创建符号链接
		fData.WriteString("	# 备用方案：手动创建符号链接\n")
		fData.WriteString("	mkdir -p /etc/rc0.d /etc/rc1.d /etc/rc2.d /etc/rc3.d /etc/rc4.d /etc/rc5.d /etc/rc6.d\n")
		fData.WriteString(fmt.Sprintf("	ln -sf /etc/init.d/%s /etc/rc1.d/S20%s 2>/dev/null || true\n", serverName, serverName))
		fData.WriteString(fmt.Sprintf("	ln -sf /etc/init.d/%s /etc/rc2.d/S20%s 2>/dev/null || true\n", serverName, serverName))
		fData.WriteString(fmt.Sprintf("	ln -sf /etc/init.d/%s /etc/rc3.d/S20%s 2>/dev/null || true\n", serverName, serverName))
		fData.WriteString(fmt.Sprintf("	ln -sf /etc/init.d/%s /etc/rc4.d/S20%s 2>/dev/null || true\n", serverName, serverName))
		fData.WriteString(fmt.Sprintf("	ln -sf /etc/init.d/%s /etc/rc5.d/S20%s 2>/dev/null || true\n", serverName, serverName))
		fData.WriteString(fmt.Sprintf("	ln -sf /etc/init.d/%s /etc/rc0.d/S60%s 2>/dev/null || true\n", serverName, serverName))
		fData.WriteString(fmt.Sprintf("	ln -sf /etc/init.d/%s /etc/rc6.d/S60%s 2>/dev/null || true\n", serverName, serverName))
	case DistroUbuntu, DistroDebian:
		// Ubuntu/Debian使用update-rc.d注册服务
		fData.WriteString(fmt.Sprintf("	# Ubuntu/Debian系统使用update-rc.d\n"))
		fData.WriteString(fmt.Sprintf("	-update-rc.d %s defaults 20 2>/dev/null || true\n", serverName))
		// 备用方案：手动创建符号链接
		fData.WriteString("	# 备用方案：手动创建符号链接\n")
		fData.WriteString("	mkdir -p /etc/rc0.d /etc/rc1.d /etc/rc2.d /etc/rc3.d /etc/rc4.d /etc/rc5.d /etc/rc6.d\n")
		fData.WriteString(fmt.Sprintf("	ln -sf /etc/init.d/%s /etc/rc1.d/S20%s 2>/dev/null || true\n", serverName, serverName))
		fData.WriteString(fmt.Sprintf("	ln -sf /etc/init.d/%s /etc/rc2.d/S20%s 2>/dev/null || true\n", serverName, serverName))
		fData.WriteString(fmt.Sprintf("	ln -sf /etc/init.d/%s /etc/rc3.d/S20%s 2>/dev/null || true\n", serverName, serverName))
		fData.WriteString(fmt.Sprintf("	ln -sf /etc/init.d/%s /etc/rc4.d/S20%s 2>/dev/null || true\n", serverName, serverName))
		fData.WriteString(fmt.Sprintf("	ln -sf /etc/init.d/%s /etc/rc5.d/S20%s 2>/dev/null || true\n", serverName, serverName))
		fData.WriteString(fmt.Sprintf("	ln -sf /etc/init.d/%s /etc/rc0.d/S60%s 2>/dev/null || true\n", serverName, serverName))
		fData.WriteString(fmt.Sprintf("	ln -sf /etc/init.d/%s /etc/rc6.d/S60%s 2>/dev/null || true\n", serverName, serverName))
	default:
		// 未知发行版，使用通用方法
		fData.WriteString("	# 未知发行版，使用通用方法创建符号链接\n")
		fData.WriteString("	mkdir -p /etc/rc0.d /etc/rc1.d /etc/rc2.d /etc/rc3.d /etc/rc4.d /etc/rc5.d /etc/rc6.d\n")
		fData.WriteString(fmt.Sprintf("	ln -sf /etc/init.d/%s /etc/rc1.d/S20%s\n", serverName, serverName))
		fData.WriteString(fmt.Sprintf("	ln -sf /etc/init.d/%s /etc/rc2.d/S20%s\n", serverName, serverName))
		fData.WriteString(fmt.Sprintf("	ln -sf /etc/init.d/%s /etc/rc3.d/S20%s\n", serverName, serverName))
		fData.WriteString(fmt.Sprintf("	ln -sf /etc/init.d/%s /etc/rc4.d/S20%s\n", serverName, serverName))
		fData.WriteString(fmt.Sprintf("	ln -sf /etc/init.d/%s /etc/rc5.d/S20%s\n", serverName, serverName))
		fData.WriteString(fmt.Sprintf("	ln -sf /etc/init.d/%s /etc/rc0.d/S60%s\n", serverName, serverName))
		fData.WriteString(fmt.Sprintf("	ln -sf /etc/init.d/%s /etc/rc6.d/S60%s\n", serverName, serverName))
		// 尝试使用update-rc.d（Ubuntu/Debian）
		fData.WriteString(fmt.Sprintf("	-update-rc.d %s defaults 20 2>/dev/null || true\n", serverName))
		// 尝试使用chkconfig（CentOS/RHEL）
		fData.WriteString(fmt.Sprintf("	-chkconfig --add %s 2>/dev/null || true\n", serverName))
		fData.WriteString(fmt.Sprintf("	-chkconfig %s on 2>/dev/null || true\n", serverName))
	}

	fData.WriteString("	echo \"control complete\"")
	filebase.WriteDataByte("makefile", fData.Bytes())

	return nil
}
