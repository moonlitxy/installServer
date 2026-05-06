package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/moonlitxy/installServer/filebase"
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
		} else if serverIndex == "2" {
			// 删除服务不需要选择可执行程序，直接跳到服务名称输入
			fmt.Println("删除服务操作")
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
				// 对于删除服务，不需要创建服务文件，直接执行删除操作
				if serverIndex == "2" {
					// 删除服务
					fmt.Println("正在删除服务...")
					err = LinuxSererInstall(serverIndex, serverName)
					if err != nil {
						fmt.Printf("服务删除失败: %v\n", err)
						fmt.Println("请检查以下可能的问题:")
						fmt.Println("1. 是否有足够的权限（可能需要使用sudo运行）")
						fmt.Println("2. 服务是否存在")
					} else {
						fmt.Println("================================")
						fmt.Printf("✓ 服务 '%s' 删除成功！\n", serverName)
						fmt.Println("================================")

						// 验证服务是否已删除
						_, systemdCheck := exec.LookPath("systemctl")
						if systemdCheck == nil {
							// 使用systemctl验证服务是否已删除
							statusCmd := exec.Command("systemctl", "status", serverName)
							err = statusCmd.Run()
							if err != nil {
								fmt.Println("✓ 确认: 服务已从系统中完全删除")
							} else {
								fmt.Println("⚠ 警告: 服务可能未完全删除，请手动检查")
								fmt.Println("尝试手动检查: systemctl status", serverName)
							}
						} else {
							// 使用service命令验证服务是否已删除
							statusCmd := exec.Command("service", serverName, "status")
							err = statusCmd.Run()
							if err != nil {
								fmt.Println("✓ 确认: 服务已从系统中完全删除")
							} else {
								fmt.Println("⚠ 警告: 服务可能未完全删除，请手动检查")
								fmt.Println("尝试手动检查: service", serverName, "status")
							}
						}
					}
				} else if serverIndex == "3" || serverIndex == "4" || serverIndex == "5" {
					// 启动、停止或查看服务状态
					var operationName string
					switch serverIndex {
					case "3":
						operationName = "启动"
					case "4":
						operationName = "停止"
					case "5":
						operationName = "查看状态"
					}

					fmt.Printf("正在%s服务...\n", operationName)
					err = LinuxSererInstall(serverIndex, serverName)
					if err != nil {
						fmt.Printf("服务%s失败: %v\n", operationName, err)
						fmt.Println("请检查以下可能的问题:")
						fmt.Println("1. 是否有足够的权限（可能需要使用sudo运行）")
						fmt.Println("2. 服务是否存在")
					} else {
						// 根据操作类型显示不同的成功消息
						switch serverIndex {
						case "3":
							fmt.Println("================================")
							fmt.Printf("✓ 服务 '%s' 启动成功！\n", serverName)
							fmt.Println("================================")
							// 获取并显示服务PID
							line := ReadServerStat(serverName)
							if line != "" {
								fmt.Printf("✓ 服务PID: %s\n", line)
							} else {
								fmt.Println("⚠ 无法获取服务PID")
							}
						case "4":
							fmt.Println("================================")
							fmt.Printf("✓ 服务 '%s' 停止成功！\n", serverName)
							fmt.Println("================================")
						case "5":
							fmt.Println("================================")
							fmt.Printf("✓ 服务 '%s' 状态查询完成！\n", serverName)
							fmt.Println("================================")
							line := ReadServerStat(serverName)
							if line != "" {
								fmt.Printf("✓ 服务PID: %s\n", line)
								fmt.Println("✓ 服务状态: 运行中")
							} else {
								fmt.Println("✓ 服务状态: 未运行或不存在")
							}
						}
					}
				} else {
					// 添加服务
					switch serverIndex {
					case "1":
						fmt.Println("正在安装服务...")
						// 检查系统是否使用systemd
						_, systemdCheck := exec.LookPath("systemctl")
						if systemdCheck == nil {
							// 使用systemd方式安装服务
							fmt.Println("检测到systemd，使用systemd方式安装服务")
							err = CreateSystemdService(exePath, exeName[exeIndex], serverName)
							if err != nil {
								fmt.Printf("创建systemd服务文件失败: %v\n", err)
								fmt.Println("请检查以下可能的问题:")
								fmt.Println("1. 是否有足够的权限（可能需要使用sudo运行）")
								fmt.Println("2. /etc/systemd/system/目录是否可写")
								fmt.Println("3. 可执行文件路径是否正确")
								break
							}

							// 启用服务
							enableCmd := exec.Command("systemctl", "enable", serverName)
							err = enableCmd.Run()
							if err != nil {
								fmt.Printf("启用服务失败: %v\n", err)
								fmt.Println("尝试手动启用: systemctl enable", serverName)
							} else {
								fmt.Println("✓ 服务已启用")
							}
						} else {
							// 使用传统sysvinit方式安装服务
							fmt.Println("未检测到systemd，使用传统sysvinit方式安装服务")
							err = CreateShellService(exePath, exeName[exeIndex], serverName)
							if err != nil {
								fmt.Printf("创建shell服务脚本失败: %v\n", err)
								break
							}

							err = CrrateMakefile(serverName)
							if err != nil {
								fmt.Printf("创建Makefile失败: %v\n", err)
								break
							}
						}

						//开始执行安装流程
						err = LinuxSererInstall(serverIndex, serverName)
						if err != nil {
							fmt.Printf("服务安装失败: %v\n", err)
							fmt.Println("请检查以下可能的问题:")
							fmt.Println("1. 是否有足够的权限（可能需要使用sudo运行）")
							fmt.Println("2. 系统是否支持systemd或sysvinit")
							fmt.Println("3. 服务名称是否已被占用")
							fmt.Println("4. 可执行文件路径是否正确")
						} else {
							fmt.Println("================================")
							fmt.Printf("✓ 服务 '%s' 安装成功！\n", serverName)
							fmt.Println("================================")

							// 验证服务是否正确安装
							_, systemdCheck := exec.LookPath("systemctl")
							if systemdCheck == nil {
								// 使用systemctl验证服务状态
								statusCmd := exec.Command("systemctl", "status", serverName)
								err = statusCmd.Run()
								if err != nil {
									fmt.Printf("⚠ 警告: 服务可能未正确安装，systemctl status失败: %v\n", err)
									fmt.Println("尝试手动检查: systemctl status", serverName)
								} else {
									fmt.Println("✓ 服务状态检查通过")
								}
							} else {
								// 使用service命令验证服务状态
								statusCmd := exec.Command("service", serverName, "status")
								err = statusCmd.Run()
								if err != nil {
									fmt.Printf("⚠ 警告: 服务可能未正确安装，service status失败: %v\n", err)
									fmt.Println("尝试手动检查: service", serverName, "status")
								} else {
									fmt.Println("✓ 服务状态检查通过")
								}
							}
						}

						// 尝试获取服务PID
						line := ReadServerStat(serverName)
						if line != "" {
							fmt.Printf("✓ 服务PID: %s\n", line)
						} else {
							fmt.Println("⚠ 无法获取服务PID，服务可能未运行")
							fmt.Println("尝试手动启动服务:")
							_, systemdCheck := exec.LookPath("systemctl")
							if systemdCheck == nil {
								fmt.Println("systemctl start", serverName)
							} else {
								fmt.Println("service", serverName, "start")
							}
						}
					}
				}
			case "windows":
				var operationName string
				switch serverIndex {
				case "1":
					operationName = "安装"
				case "2":
					operationName = "删除"
				case "3":
					operationName = "启动"
				case "4":
					operationName = "停止"
				}

				fmt.Printf("正在%s服务...\n", operationName)
				switch serverIndex {
				case "2", "3", "4":
					err = WindowServerInstall(serverIndex, serverName, "")
				case "1":
					err = WindowServerInstall(serverIndex, serverName, exeName[exeIndex])
				}
				if err != nil {
					fmt.Println(err, "请检查是否安装服务或服务名是否正确")
				} else {
					fmt.Println("================================")
					fmt.Printf("✓ 服务 '%s' %s成功！\n", serverName, operationName)
					fmt.Println("================================")
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
