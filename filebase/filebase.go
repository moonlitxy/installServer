package filebase

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

/*****************
文件的读取和写入
******************/

// <summary>
// 读取本地运行路径
// </summary>
// <returns></returns>
func GetLocalPath() string {
	var filePath string
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return ""
	}
	filePath, err = filepath.Abs(file)
	if err != nil {
		return ""
	}
	filePath = strings.Replace(filePath, "\\", "/", -1)
	return filePath
}

/*
根据文件地址获取文件路径
如：/usr/app_install/db/bin/db
返回：/usr/app_install/db/bin,db
*/
func GetFilePath(fileAddr string) (string, string) {
	fileAddr = strings.Replace(fileAddr, "\\", "/", -1)
	strs := strings.Split(fileAddr, "/")
	if len(strs) <= 0 {
		return fileAddr, ""
	}
	sName := strs[len(strs)-1]
	sPath := ""
	for s := 0; s < len(strs)-1; s++ {
		sPath += strs[s] + "/"
	}
	os.MkdirAll(sPath, 0777)
	os.Chmod(sPath, 0777)
	return sPath, sName
}

/** 创建文件路径
 */
func CreateDir(filePath string) error {
	fPath := changefilePath(filePath)
	os.MkdirAll(fPath, 0777)
	os.Chmod(fPath, 0777)
	return nil
}

/*
******
文件的读取
******
*/
func ReadData(fileAddr string) string {
	fFile := changefilePath(fileAddr)
	inputFile, inputError := ioutil.ReadFile(fFile)
	if inputError != nil {
		log.Println("An error occurred on opening the inputfile\n" +
			"Does the file exist?\n" +
			"Have you got acces to it?\n")
		return ""
	}
	return string(inputFile)
}

/*
******
文件的写入
******
*/
func WriteData(fileAddr string, strContent string) bool {
	var flag bool
	fFile := changefilePath(fileAddr)
	outputFile, outputError := os.OpenFile(fFile, os.O_WRONLY|os.O_CREATE, 0777)

	if outputError != nil {
		fmt.Printf("An error occurred with file opening or creation\n")
		return flag
	}
	os.Chmod(fFile, 0777)
	defer outputFile.Close()

	outputWriter := bufio.NewWriter(outputFile)
	outputString := strContent

	outputWriter.WriteString(outputString)

	outputWriter.Flush()

	flag = true
	return flag
}

/*
* 写文件数据
直接覆盖原数据
*/
func WriteDataByte(fileAddr string, Databyte []byte) bool {
	var flag bool
	fFile := changefilePath(fileAddr)
	outputFile, outputError := os.OpenFile(fFile, os.O_WRONLY|os.O_CREATE, 0777)

	if outputError != nil {
		fmt.Printf("An error occurred with file opening or creation\n")
		fmt.Println(fFile)
		return flag
	}
	defer outputFile.Close()

	outputWriter := bufio.NewWriter(outputFile)
	outputWriter.Write(Databyte)
	outputWriter.Flush()

	flag = true
	return flag
}

func AppendDataByte(fileAddr string, Databyte []byte) bool {
	var flag bool
	fFile := changefilePath(fileAddr)
	var outputFile *os.File
	var outputError error
	if CheckFileIsExist(fFile) {
		outputFile, outputError = os.OpenFile(fFile, os.O_APPEND, 0777)
	} else {
		outputFile, outputError = os.OpenFile(fFile, os.O_WRONLY|os.O_CREATE, 0777)
	}

	if outputError != nil {
		fmt.Printf("An error occurred with file opening or creation\n")
		fmt.Println(fFile)
		return flag
	}
	defer outputFile.Close()

	outputWriter := bufio.NewWriter(outputFile)
	outputWriter.Write(Databyte)
	outputWriter.Flush()

	flag = true
	return flag
}

/*
获取文件大小
fileAddr := 文件路径
*/
func GetFileLens(fileAddr string) (int64, error) {
	fFile := changefilePath(fileAddr)
	fs, err := os.Stat(fFile)
	if err != nil {
		return -1, err
	}
	return fs.Size(), nil
}

/*
获取文件数据
fileAddr=文件本地路径
sIndex=获取序号，sIndex 从1开始，每次返回1024字节数据
*/
func GetFileData(fileAddr string, sIndex int) ([]byte, error) {
	fFile := changefilePath(fileAddr)
	fi, err := os.Open(fFile)
	if err != nil {
		return nil, err
	}
	defer fi.Close()
	buf := make([]byte, 1024)
	var lens int64
	lens = (int64)(sIndex-1) * 1024
	if lens < 0 {
		return nil, fmt.Errorf("序号错误")
	}

	n, err := fi.ReadAt(buf, lens)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if n == 0 {
		return nil, fmt.Errorf("文件已读取完毕")
	}
	return buf[:n], nil
}

/*
获取文件所有数据
fileAddr := 文件地址
*/
func GetFileDataAll(fileAddr string) ([]byte, error) {
	fFile := changefilePath(fileAddr)
	fi, err := os.Open(fFile)
	//fmt.Println(fFile)
	if err != nil {
		return []byte{}, err
	}
	defer fi.Close()
	return ioutil.ReadAll(fi)

}

/** 判断文件是否存在
 *   存在返回 true 不存在返回false
 */
func CheckFileIsExist(filePath string) bool {
	var exist = true
	fFile := changefilePath(filePath)
	if _, err := os.Stat(fFile); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

/** 获取文件夹中所有文件
 */
func GetFileList(sPath string) []string {
	fFile := changefilePath(sPath)
	//获取所有文件
	dir, err := os.Open(fFile)
	if err != nil {
		return nil
	}
	defer dir.Close()
	// 读取文件列表
	fis, err := dir.Readdir(0)
	if err != nil {
		panic(err)
	}
	var fileList []string

	// 遍历文件列表
	for _, fi := range fis {
		fileList = append(fileList, fi.Name())
	}
	return fileList
}

/*
	拷贝文件

src = 源文件
des = 目标文件
*/
func CopyFile(src, des string) (w int64, err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	desFile, err := os.Create(des)
	if err != nil {
		return 0, err
	}
	defer desFile.Close()

	return io.Copy(desFile, srcFile)
}

/** 删除文件
 */
func DelFile(FileAddress string) error {
	return os.Remove(FileAddress)
}

/*
* 文件路径替换
由于"./"路径在服务等情况无法正常获取文件，因此在碰到此类路径需要先转换成绝对路径
*/
func changefilePath(sourcePath string) (desPath string) {
	desPath = sourcePath
	if strings.HasPrefix(sourcePath, "./") {
		path := GetLocalPath()
		pathDir, _ := GetFilePath(path)
		desPath = strings.Replace(sourcePath, "./", pathDir, -1)
	}
	return
}
