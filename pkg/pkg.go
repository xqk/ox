// Package pkg
// @Description: 外部应用程序可以使用的库代码
package pkg

import (
	"fmt"
	"os"
	"ox/pkg/constant"
	"path/filepath"
)

const oxVersion = "0.0.1"

var (
	appName         string
	appID           string
	hostName        string
	buildAppVersion string
	buildUser       string
	buildHost       string
	buildStatus     string
	buildTime       string
)

func init() {
	if appName == "" {
		appName = os.Getenv(constant.EnvAppName)
		if appName == "" {
			appName = filepath.Base(os.Args[0])
		}
	}

	name, err := os.Hostname()
	if err != nil {
		name = "unknown"
	}
	hostName = name
}

//
// PrintVersion
// @Description: 打印版本信息
//
func PrintVersion() {
	fmt.Printf("%-8s]> %-30s => %s\n", "ox", "name", appName)
}
