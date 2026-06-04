//go:build gui

package main

import (
	"flag"
	"log"

	"gopay/src/config"
	"gopay/src/systray"
)

func main() {
	configPath := flag.String("config", "", "配置文件路径（默认自动查找 gopay.env / config.env / /etc/gopay/config.env）")
	dbPath := flag.String("db", "", "数据库连接字符串（PostgreSQL）或文件路径（SQLite）")
	host := flag.String("host", "", "监听IP地址（默认 0.0.0.0）")
	port := flag.String("port", "", "服务端口（默认 8080）")
	migrate := flag.Bool("migrate", false, "执行数据库迁移")
	flag.Parse()

	// 查找并加载配置文件
	if cfgFile := config.FindConfigFile(*configPath); cfgFile != "" {
		if err := config.LoadConfigFile(cfgFile); err != nil {
			log.Fatalf("[config] %v", err)
		}
	}

	// 合并配置：命令行参数 > 配置文件 > 默认值
	db := config.GetConfigValue(*dbPath, "DB", "")
	h := config.GetConfigValue(*host, "HOST", "0.0.0.0")
	p := config.GetConfigValue(*port, "PORT", "8080")

	r, addr, openURL, dbFile := initRuntime(runtimeOptions{
		DBPath:  db,
		Host:    h,
		Port:    p,
		Migrate: *migrate,
	})
	printStartupLogs(addr, openURL, dbFile)

	systray.Init("GoPay", "GoPay 支付服务运行中", openURL, nil)

	systray.Run(func() {
		go func() {
			if err := r.Run(addr); err != nil {
				log.Fatalf("[server] failed to start: %v", err)
			}
		})
	})
}
