package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	// 架构配置
	architectures := []struct {
		name   string
		env    []string
		output string
		upx    bool
	}{
		{
			name: "ARMv7",
			env: []string{
				"GOOS=linux", "GOARCH=arm", "GOARM=7", "CGO_ENABLED=0",
			},
			output: "hwddns-armv7",
			upx:    true,
		},
		{
			name: "ARMv8",
			env: []string{
				"GOOS=linux", "GOARCH=arm64", "CGO_ENABLED=0",
			},
			output: "hwddns-armv8",
			upx:    true,
		},
		{
			name: "MIPS",
			env: []string{
				"GOOS=linux", "GOARCH=mips", "GOMIPS=softfloat", "CGO_ENABLED=0",
			},
			output: "hwddns-mips-softfpu",
			upx:    true,
		},
		{
			name: "MIPSEL",
			env: []string{
				"GOOS=linux", "GOARCH=mipsle", "GOMIPS=softfloat", "CGO_ENABLED=0",
			},
			output: "hwddns-mipsel-softfpu",
			upx:    true,
		},
		{
			name: "Windows",
			env: []string{
				"GOOS=windows", "GOARCH=amd64", "CGO_ENABLED=0",
			},
			output: "hwddns-windows-amd64.exe",
			upx:    true,
		},
		{
			name: "Linux AMD64",
			env: []string{
				"GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0",
			},
			output: "hwddns-linux-amd64",
			upx:    true,
		},
	}

	// 获取版本信息
	version := getVersionInfo()
	fmt.Printf("构建版本: %s\n", version)

	// 创建输出目录
	outputDir := "build"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("创建输出目录失败: %v\n", err)
		os.Exit(1)
	}

	// 检查UPX是否安装
	upxInstalled := checkUpxInstalled()
	if !upxInstalled {
		fmt.Println("警告: UPX未安装，将跳过压缩步骤")
		fmt.Println("请安装UPX: sudo apt install upx")
	}

	// 构建所有架构
	for _, arch := range architectures {
		fmt.Printf("\n=== 构建 %s ===\n", arch.name)

		outputPath := filepath.Join(outputDir, arch.output)

		// 构建命令
		cmd := exec.Command("go", "build",
			"-ldflags", fmt.Sprintf("-X main.buildVersion=%s -s -w", version),
			"-trimpath",
			"-o", outputPath,
			"./main.go") // 构建项目根目录的 main.go

		cmd.Env = append(os.Environ(), arch.env...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// 执行构建
		if err := cmd.Run(); err != nil {
			fmt.Printf("错误: 构建 %s 失败: %v\n", arch.name, err)
			continue
		}

		// 获取文件大小
		if info, err := os.Stat(outputPath); err == nil {
			size := float64(info.Size()) / (1024 * 1024)
			fmt.Printf("构建成功: %s (%.2f MB)\n", outputPath, size)
		}

		// UPX压缩
		if arch.upx && upxInstalled {
			compressedPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + "-upx" + filepath.Ext(outputPath)

			upxCmd := exec.Command("upx",
				"--best",
				"--lzma",
				"-o", compressedPath,
				outputPath)

			upxCmd.Stdout = os.Stdout
			upxCmd.Stderr = os.Stderr

			fmt.Printf("压缩 %s...\n", outputPath)
			if err := upxCmd.Run(); err != nil {
				fmt.Printf("警告: UPX压缩失败: %v\n", err)
			} else {
				// 获取压缩后大小
				if info, err := os.Stat(compressedPath); err == nil {
					size := float64(info.Size()) / (1024 * 1024)
					fmt.Printf("压缩成功: %s (%.2f MB)\n", compressedPath, size)
				}
			}
		}
	}

	fmt.Println("\n所有架构构建完成!")
}

// 获取版本信息
func getVersionInfo() string {
	// 尝试获取Git提交信息
	commitCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	if commit, err := commitCmd.Output(); err == nil {
		return strings.TrimSpace(string(commit))
	}

	// 使用构建时间作为备选
	return time.Now().Format("20060102")
}

// 检查UPX是否安装
func checkUpxInstalled() bool {
	cmd := exec.Command("upx", "--version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
