package main

import (
	"bufio"
	"fmt"
	"github.com/spf13/viper"
	_ "io/fs"
	"log"
	"os/exec"
	"path/filepath"
	"runtime"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	abs, err := filepath.Abs(".")
	if err != nil {
		log.Fatalln("获取目录失败", err)
	}
	fmt.Println("当前执行目录: ", abs)
	viper.AddConfigPath(".")
	viper.AddConfigPath("../")
	viper.SetConfigType("yaml")
	viper.SetConfigType("yml")
	viper.SetConfigName("build")
	readErr := viper.ReadInConfig()
	if readErr != nil {
		log.Fatalln(readErr)
		return
	}
	isWindows := runtime.GOOS == "windows"
	output := viper.GetString("go.build.output")
	configs := viper.GetStringMapString("go.build.package")
	for k, v := range configs {
		if isWindows {
			v += ".exe"
		}
		if output != "" {
			v = filepath.Join(output, v)
		}

		goBuild(filepath.Join(abs, k), filepath.Join(abs, v))
	}
}

func goBuild(path string, name string) {
	commandLine := fmt.Sprintf("go build -C %s -o %s", path, name)
	fmt.Println("当前执命令: ", commandLine)
	sysType := runtime.GOOS
	var cmd *exec.Cmd
	if sysType == "windows" {
		cmd = exec.Command("cmd", "/c", commandLine)
	} else {
		cmd = exec.Command("sh", "-c", commandLine)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get stdout pipe: %s", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalf("Failed to get stderr pipe: %s", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start command: %s", err)
	}

	// 使用 goroutine 实时读取标准输出
	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			fmt.Printf("stdout: %s\n", scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Error reading stdout: %s", err)
		}
	}()

	// 使用 goroutine 实时读取标准错误
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			fmt.Printf("stderr: %s\n", scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Error reading stderr: %s", err)
		}
	}()

	// 等待命令完成
	if err := cmd.Wait(); err != nil {
		log.Fatalf("Command finished with error: %s", err)
	}
}
