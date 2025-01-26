package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/jericho-yu/aid/filesystem"
	"github.com/jericho-yu/aid/honestMan"
	"github.com/jericho-yu/aid/operation"
	"github.com/jericho-yu/aid/str"
)

type (
	Config struct {
		Basic    ConfigBasic `yaml:"basic"`
		CopyFile []string    `yaml:"copy-file"`
		CopyDir  []string    `yaml:"copy-dir"`
	}

	ConfigBasic struct {
		SaveDir         string `yaml:"save-dir"`
		Name            string `yaml:"name"`
		Version         string `yaml:"version"`
		Dockerfile      string `yaml:"dockerfile"`
		AutoSaveFile    bool   `yaml:"auto-save-file"`
		AutoDeleteImage bool   `yaml:"auto-delete-image"`
	}
)

// 编译go程序
func buildGo(config Config) {
	var (
		title   = "[1/4]编译go程序"
		err     error
		saveDir = filesystem.FileSystemApp.NewByRelative(config.Basic.SaveDir).Join(config.Basic.Version)
		cmd     *exec.Cmd
		appName string
		output  []byte
		os      = os.Getenv("GOOS")
	)
	str.NewTerminalLog(title, "：%s").Info("开始")

	if !saveDir.IsExist {
		saveDir.MkDir()
	}

	appName = fmt.Sprintf(operation.Ternary[string](os == "windows", "%s/%s.exe", "%s/%s"), saveDir.GetDir(), config.Basic.Name)

	// 定义要执行的命令
	cmd = exec.Command("go", "build", "-o", appName, ".")

	// 获取命令的输出
	output, err = cmd.Output()
	if err != nil {
		str.NewTerminalLog(title, "错误：%v").Error(err)
	}

	log.Printf("OK: %s", output)
	str.NewTerminalLog(title, "成功").Success()
}

// 复制文件
func copyFile(config Config) {
	var (
		title = "[2/4]复制文件"
		err   error
		cmd   *exec.Cmd
		src   = filesystem.FileSystemApp.NewByRelative(".")
		dst   = filesystem.FileSystemApp.NewByRelative(config.Basic.SaveDir).Join(config.Basic.Version)
	)
	str.NewTerminalLog(title, "%s").Info("开始")
	str.NewTerminalLog("复制Dockerfile").Default()
	cmd = exec.Command("cp", src.Copy().Join(config.Basic.Dockerfile).GetDir(), dst.GetDir())
	_, err = cmd.Output()
	if err != nil {
		str.NewTerminalLog(title, "错误：%v").Error(err)
	}

	str.NewTerminalLog("复制其他文件").Default()
	for _, file := range config.CopyFile {
		filenames := strings.Split(file, " => ")
		if len(filenames) == 1 {
			cmd = exec.Command("cp", src.Copy().Join(filenames[0]).GetDir(), dst.Copy().Join(filenames[0]).GetDir())
		} else {
			cmd = exec.Command("cp", src.Copy().Join(filenames[0]).GetDir(), dst.Copy().Join(filenames[1]).GetDir())
		}
		_, err = cmd.Output()
		if err != nil {
			str.NewTerminalLog("复制文件错误：%s -> %v").Error(file, err)
		}
	}

	str.NewTerminalLog("复制文件夹").Default()
	for _, dir := range config.CopyDir {
		dirNames := strings.Split(dir, " => ")
		if len(dirNames) == 1 {
			cmd = exec.Command("cp", "-R", src.Copy().Join(dirNames[0]).GetDir(), dst.Copy().Join(dirNames[0]).GetDir())
		} else {
			cmd = exec.Command("cp", "-R", src.Copy().Join(dirNames[0]).GetDir(), dst.Copy().Join(dirNames[1]).GetDir())
		}
		_, err = cmd.Output()
		if err != nil {
			str.NewTerminalLog("复制文件夹错误：%s -> %v").Error(dir, err)
		}
	}

	str.NewTerminalLog(title, "成功").Success()
}

// 编译docker镜像
func buildImage(config Config) {
	var (
		title  = "[3/4]编译docker镜像"
		err    error
		cmd    *exec.Cmd
		src    = filesystem.FileSystemApp.NewByRelative(config.Basic.SaveDir).Join(config.Basic.Version)
		output []byte
	)
	str.NewTerminalLog(title, "%s").Info("开始")
	cmd = exec.Command(
		"docker",
		"build",
		"-f",
		src.Copy().Join(config.Basic.Dockerfile).GetDir(),
		"-t",
		config.Basic.Name+":"+config.Basic.Version,
		src.GetDir(),
	)
	output, err = cmd.Output()
	if err != nil {
		str.NewTerminalLog(title, "错误：%s -> %v").Error(output, err)
	}
	str.NewTerminalLog(title, "成功").Success()
}

// 保存镜像
func saveImage(config Config) {
	var (
		title = "[4/4]保存docker镜像"
		err   error
		cmd   *exec.Cmd
		src   = filesystem.FileSystemApp.NewByRelative(config.Basic.SaveDir).Join(config.Basic.Version)
	)
	str.NewTerminalLog(title, "%s").Info("开始")
	cmd = exec.Command(
		"docker",
		"save",
		"-o",
		src.Join(config.Basic.Name+"_"+config.Basic.Version+".tar").GetDir(),
		config.Basic.Name+":"+config.Basic.Version,
	)
	_, err = cmd.Output()
	if err != nil {
		str.NewTerminalLog("保存docker镜像错误：%v").Error(err)
	}
	str.NewTerminalLog(title, "成功").Success()
}

// 删除镜像
func deleteImage(config Config) {
	var (
		title = "清理docker镜像"
		err   error
		cmd   *exec.Cmd
	)
	str.NewTerminalLog(title, "%s").Info("开始")
	cmd = exec.Command("docker", "rmi", config.Basic.Name+":"+config.Basic.Version)
	_, err = cmd.Output()
	if err != nil {
		str.NewTerminalLog(title, "错误：%v").Error(err)
	}
	str.NewTerminalLog(title, "成功").Success()
}

func main() {
	var (
		err    error
		config Config
	)
	os.Setenv("AID__STR__TERMINAL_LOG__ENABLE", "true") // 打开终端日志

	if err = honestMan.App.New("docker-image-build.yaml").LoadYaml(&config); err != nil {
		str.NewTerminalLog("读取配置错误：%v").Error(err)
	}

	buildGo(config)    // 编译go程序
	copyFile(config)   // 复制文件
	buildImage(config) // 编译docker镜像
	if config.Basic.AutoSaveFile {
		saveImage(config) // 保存镜像
	}
	if config.Basic.AutoDeleteImage {
		deleteImage(config) // 清理镜像
	}
}
