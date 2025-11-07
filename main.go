package main

import (
	"bufio"
	"fmt"
	"github.com/moonlitxy/installServer/filebase"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var (
	serverName  string         //服务名称
	exeName     map[int]string //执行程序名称
	exeIndex    int            //文件下标
	serverIndex string         //操作序号
	exePath     string         //执行程序路径
	sys         string         //系统信息
	info        string         //系统版本

)

var (
	smsServer = []string{"com.wanwei.dclouds.sms.main.SmsServer", "org.DCloudServer.socket.main.ServiceCenterMain", "com.wanwei.dclouds.dataSync.main.DataSyncServer"}
)

func main() {
	fmt.Println("服务自动安装程序，版本号V0.0.1_20160517")
	//判断系统信息
	info = runtime.GOOS
	fmt.Println("运行系统为:", info)

	switch info {
	case "linux":

		for {
			fmt.Println("选择输入序号进行操作:\n1:Go服务\n2:java服务\n3:退出")
			fmt.Println("---------------------")
			fmt.Print("选择序号:")
			strs := ReadCmd()
			t, _ := strconv.Atoi(strs)

			//退出
			if t == 3 {
				break
			}

			if t > 3 || t <= 0 {
				fmt.Println("请输入正确的序号")
				continue
			}
			switch t {
			case 1:
				fmt.Println("选择的是Go服务操作")
				GoServerFile()
				break
			case 2:
				fmt.Println("选择的是Java服务操作")
				JavaServerFile()
				break
			}
		}

	case "windows":
		GoServerFile()
		break
	}
	fmt.Println()
}
func GoServerFile() {
	exeName = make(map[int]string, 0)

	for {
		fmt.Println("选择输入序号进行操作:\n1:添加服务\n2:删除服务\n3:启动服务\n4:停止服务\n5:查看服务状态\n6:退出")
		fmt.Println("---------------------")
		fmt.Print("选择序号:")
		strs := ReadCmd()
		t, _ := strconv.Atoi(strs)

		//退出
		if ok := strings.EqualFold(strs, "quit"); ok {
			break
		}
		if t == 6 {
			break
		}

		if t > 6 || t <= 0 {
			fmt.Println("请输入正确的序号,输入quit")
			continue
		}
		//服务操作序号
		serverIndex = strs

		if serverIndex == "1" {
			fmt.Println("正在获取可执行程序")
			strlist := filebase.GetFileList("./")

			i := 0
			for _, filename := range strlist {

				switch info {
				case "linux":
					hw, err := os.Stat(filename)
					if err != nil {
						continue
					}
					//可执行程序在Ubuntu系统为509 在centos系统为511
					if (int64(hw.Mode()) == 509 || int64(hw.Mode()) == 511) && !strings.EqualFold(hw.Name(), "installserver") {
						exeName[i] = hw.Name()
						i++
					}
				case "windows":

					suffix := filepath.Ext(filename)

					if ok := strings.EqualFold(suffix, ".exe"); ok {
						exe := filepath.Base(filename)
						if !strings.HasPrefix(strings.ToLower(exe), "installserver") && !strings.HasPrefix(strings.ToLower(exe), "nssm") {
							exeName[i] = exe
							i++
						}
					}
				}
			}
			Count := len(exeName)
			if Count == 0 {
				fmt.Println("未发现可执行程序")
				break
			}

			fmt.Println("请选择需要安装的可执行程序序号")
			for i, v := range exeName {
				fmt.Println(i+1, v)
			}
			fmt.Println("---------------------")
			for {
				strs := ReadCmd()
				t, _ := strconv.Atoi(strs)

				if ok := strings.EqualFold(strs, "quit"); ok {
					break
				}
				if t > Count || t <= 0 {
					fmt.Println("请输入正确的序号,输入quit退出")
					continue
				}
				//文件根目录路径
				sPath, _ := filebase.GetFilePath(filebase.GetLocalPath())
				//执行程序完整路径
				exeIndex = t - 1
				exePath = sPath
				break
			}
		}
	ToServerName:
		fmt.Print("输入服务名称:")

		strs = ReadCmd()
		if strs != "" {
			//判断是否输入正确的服务名称

			fmt.Print("是否采用当前服务名称,输入y确定,n重新输入")
			ACK := ReadCmd()
			if ok := strings.EqualFold(ACK, "n"); ok {
				goto ToServerName
			}
			if ok := strings.EqualFold(ACK, "y"); ok == false {
				goto ToServerName
			}

			serverName = strs
			var err error
			switch info {
			case "linux":
				switch serverIndex {
				case "1":
					//选择程序，生成可执行文件
					CreateShellService(exePath, exeName[exeIndex], serverName)
					CrrateMakefile(serverName)
				}
				//开始执行
				err = LinuxSererInstall(serverIndex, serverName)
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println("执行成功")
				}
				line := ReadServerStat(serverName)
				fmt.Println(fmt.Sprintf("服务pid:%s", line))
			case "windows":
				switch serverIndex {
				case "2", "3", "4":
					err = WindowServerInstall(serverIndex, serverName, "")
				case "1":
					err = WindowServerInstall(serverIndex, serverName, exeName[exeIndex])
				}
				if err != nil {
					fmt.Println(err, "请检查是否安装服务或服务名是否正确")
				} else {
					fmt.Println("执行成功")
				}
			}

		}

		fmt.Println("按任意键继续,输入quit退出...")
		strs = ReadCmd()
		if ok := strings.EqualFold(strs, "quit"); ok {
			break
		}
	}
}

/*
java脚本
*/
func JavaServerFile() {

	exeName = make(map[int]string, 0)
	//给程序名称赋值
	for i, v := range smsServer {
		exeName[i] = v
	}

	for {
		fmt.Println("选择输入序号进行操作:\n1:添加服务\n2:删除服务\n3:启动服务\n4:停止服务\n5:查看服务状态\n6:退出")
		fmt.Println("---------------------")
		fmt.Print("选择序号:")
		strs := ReadCmd()
		t, _ := strconv.Atoi(strs)

		//退出
		if ok := strings.EqualFold(strs, "quit"); ok {
			break
		}
		if t == 6 {
			break
		}

		if t > 6 || t <= 0 {
			fmt.Println("请输入正确的序号,输入quit")
			continue
		}
		//服务操作序号
		serverIndex = strs

		fmt.Println("程序名称序号:")

		Count := len(exeName)
		for i, v := range exeName {
			fmt.Println(i+1, v)
		}
		fmt.Println("---------------------")
		for {
			strs := ReadCmd()
			t, _ := strconv.Atoi(strs)

			if ok := strings.EqualFold(strs, "quit"); ok {
				break
			}
			if t > Count || t <= 0 {
				fmt.Println("请输入正确的序号,输入quit退出")
				continue
			}
			//文件根目录路径
			sPath, _ := filebase.GetFilePath(filebase.GetLocalPath())
			//执行程序完整路径
			exeIndex = t - 1
			exePath = sPath
			break
		}

	ToServerName:
		fmt.Print("输入服务名称:")
		strs = ReadCmd()
		if strs != "" {
			//判断是否输入正确的服务名称

			fmt.Print("是否采用当前服务名称,输入y确定,n重新输入")
			ACK := ReadCmd()
			if ok := strings.EqualFold(ACK, "n"); ok {
				goto ToServerName
			}
			if ok := strings.EqualFold(ACK, "y"); ok == false {
				goto ToServerName
			}

			serverName = strs
			var err error

			switch serverIndex {
			case "1":
				//选择程序，生成可执行文件
				CreateJavaService(exePath, exeName[exeIndex], serverName)

				CrrateMakefile(serverName)
			}
			//开始执行
			err = LinuxSererInstall(serverIndex, serverName)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("执行成功")
			}
			line := ReadServerStat(serverName)
			fmt.Println(fmt.Sprintf("服务pid:%s", line))

		}

		fmt.Println("按任意键继续,输入quit退出...")
		strs = ReadCmd()
		if ok := strings.EqualFold(strs, "quit"); ok {
			break
		}
	}
}

/*
读取输入信息
*/
func ReadCmd() string {
	run := true
	reader := bufio.NewReader(os.Stdin)
	for run {
		data, _, _ := reader.ReadLine()
		return string(data)
	}
	return ""
}
