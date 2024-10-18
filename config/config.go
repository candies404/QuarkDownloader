package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

// Config 配置结构体定义
type Config struct {
	Quark struct {
		UseDirName string `yaml:"useDirName"`
		Cookie     string `yaml:"cookie"`
	} `yaml:"quark"`
	SharesLinks []struct {
		Link struct {
			URL      string `yaml:"url"`
			PassCode string `yaml:"passCode"`
		} `yaml:"link"`
	} `yaml:"sharesLinks"`
	LocalSaveDir   string  `yaml:"localSaveDir"`
	Delay          float32 `yaml:"delay"`
	DownloadFilter struct {
		Include string `yaml:"include"`
		Exclude string `yaml:"exclude"`
	} `yaml:"downloadFilter"`
}

var Cfg Config // 全局变量保存配置

// init 函数在程序启动时自动执行
func init() {
	// 调用解析配置文件函数
	err := parseConfig("config.yaml")
	if err != nil {
		log.Fatalf("初始化时读取配置文件出错: %v", err)
	}
}

// 解析配置文件的函数
func parseConfig(filePath string) error {
	// 打开配置文件
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("无法打开配置文件: %v", err)
	}
	defer file.Close()

	// 读取文件内容并解析为配置结构体
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&Cfg)
	if err != nil {
		return fmt.Errorf("无法解析配置文件: %v", err)
	}
	if Cfg.Delay == 0 {
		Cfg.Delay = 1
	}
	// 解析成功，返回 nil 错误
	return nil
}
