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

	//根据输入序号和服务名称创建字符串
	switch index {
	case "1": //注册服务
		command = fmt.Sprintf("make")
	case "2": //删除服务
		removeServiceAll(serverName)
		os.Chdir("/etc/init.d/")
		command = fmt.Sprintf("rm -R %s", serverName)
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
	fData.WriteString("#!/bin/sh\n\n")
	fData.WriteString("ARG=$1\n\n")
	fData.WriteString("case $ARG in\n")
	fData.WriteString("start):\n")
	fData.WriteString(fmt.Sprintf("cd  %s \n", filePath))
	fData.WriteString(fmt.Sprintf("nohup  %s >/dev/null 2>&1 &\n", fileAddr))
	fData.WriteString(";;\n")
	fData.WriteString("stop):\n")
	fData.WriteString(fmt.Sprintf("pid=`ps -ef|grep %s |grep -v grep|awk '{print $2}'`\n", fileAddr))
	fData.WriteString("sleep 3\n")
	fData.WriteString("kill -9 $pid\n")
	fData.WriteString(";;\n")
	fData.WriteString("restart):\n")
	fData.WriteString(fmt.Sprintf("pid=`ps -ef|grep %s |grep -v grep|awk '{print $2}'`\n", fileAddr))
	fData.WriteString("sleep 3\n")
	fData.WriteString("kill -9 $pid\n")
	fData.WriteString("sleep 3\n")
	fData.WriteString(fmt.Sprintf("cd  %s \n", filePath))
	fData.WriteString(fmt.Sprintf("nohup  %s >/dev/null 2>&1 &\n", fileAddr))
	fData.WriteString(";;\n")
	fData.WriteString("stat):\n")
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
	fData.WriteString("start):\n")
	fData.WriteString(fmt.Sprintf("export currpath=\"%s\"\n", filePath))
	fData.WriteString(fmt.Sprintf("export CLASSPATH=$CLASSPATH:$currpath\n"))
	fData.WriteString(fmt.Sprintf("for jar in $currpath/lib/*.jar\n"))
	fData.WriteString(fmt.Sprintf("do\n"))
	fData.WriteString(fmt.Sprintf("      CLASSPATH=$CLASSPATH:$jar\n"))
	fData.WriteString(fmt.Sprintf("done\n"))
	fData.WriteString(fmt.Sprintf("export CLASSPATH\n"))
	fData.WriteString(fmt.Sprintf("java -cp $CLASSPATH %s &\n", serverName))
	fData.WriteString(";;\n")
	fData.WriteString("stop):\n")
	fData.WriteString(fmt.Sprintf("pid=`ps -ef|grep %s |grep -v grep|awk '{print $2}'`\n", serverName))
	fData.WriteString("if [[ $pid -gt 0 ]]\n")
	fData.WriteString("then\n")
	fData.WriteString("    kill -9 $pid\n")
	fData.WriteString("    sleep 3\n")
	fData.WriteString("fi\n")
	fData.WriteString(";;\n")
	fData.WriteString("restart):\n")
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
	fData.WriteString("stat):\n")
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
*/
func CrrateMakefile(serverName string) error {
	var fData bytes.Buffer

	fData.WriteString("install:\n\n")
	fData.WriteString(fmt.Sprintf("	mv ./%s /etc/init.d/%s\n\n", serverName, serverName))
	fData.WriteString("	cd /etc/init.d/\n\n")
	fData.WriteString(fmt.Sprintf("	ln -s /etc/init.d/%s /etc/rc1.d/S20%s\n\n", serverName, serverName))
	fData.WriteString(fmt.Sprintf("	ln -s /etc/init.d/%s /etc/rc2.d/S20%s\n\n", serverName, serverName))
	fData.WriteString(fmt.Sprintf("	ln -s /etc/init.d/%s /etc/rc3.d/S20%s\n\n", serverName, serverName))
	fData.WriteString(fmt.Sprintf("	ln -s /etc/init.d/%s /etc/rc4.d/S20%s\n\n", serverName, serverName))
	fData.WriteString(fmt.Sprintf("	ln -s /etc/init.d/%s /etc/rc5.d/S20%s\n\n", serverName, serverName))
	fData.WriteString(fmt.Sprintf("	ln -s /etc/init.d/%s /etc/rc0.d/S60%s\n\n", serverName, serverName))
	fData.WriteString(fmt.Sprintf("	ln -s /etc/init.d/%s /etc/rc6.d/S60%s\n\n", serverName, serverName))

	fData.WriteString("	echo \"control complete\"")
	filebase.WriteDataByte("makefile", fData.Bytes())

	return nil
}
