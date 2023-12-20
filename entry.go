package svc_cfgfile

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/pflag"
	"github.com/walleframe/walle/services/configcentra"
)

var appName string
var dumpConfigFlag bool

// UseConfigFile use config file to start app
func UseConfigFile() {
	configcentra.ConfigCentraBackend = &ConfigFileBackend{}

	appName = filepath.Base(os.Args[0])
	if runtime.GOOS == "windows" {
		appName = strings.TrimSuffix(appName, ".exe")
	}

	// 命令行参数绑定
	pflag.BoolVar(&dumpConfigFlag, "dump_config", false, "dump service config")
	pflag.String("config_file", fmt.Sprintf("./conf/%s.toml", appName), "app config file")
}
