package svc_cfgfile

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/walleframe/walle/app"
	"github.com/walleframe/walle/services/configcentra"
)

type ConfigItem struct {
	Value configcentra.ConfigValue
	Ntfs  []configcentra.ConfigUpdateNotify
}

type ConfigFileService struct {
	ctrl   *viper.Viper
	values []ConfigItem
	ntfs   []configcentra.ConfigUpdateNotify
}

var _ configcentra.ConfigCentra = (*ConfigFileService)(nil)

func (svc *ConfigFileService) Init(s app.Stoper) (err error) {
	svc.ctrl = viper.New()
	cfg := svc.ctrl
	cfg.BindPFlags(pflag.CommandLine)

	// 环境变量
	cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	cfg.AutomaticEnv()
	// 配置文件
	configFile := cfg.GetString("config_file")
	log.Println("config file name:", configFile)
	cfg.SetConfigName(strings.TrimSuffix(filepath.Base(configFile), filepath.Ext(configFile)))
	cfg.SetConfigType(strings.TrimPrefix(filepath.Ext(configFile), "."))
	cfg.AddConfigPath(filepath.Dir(configFile))

	// 设置默认值
	for _, vc := range svc.values {
		vc.Value.SetDefaultValue(svc)
	}

	if dumpConfigFlag {
		cfg.WriteConfigAs(fmt.Sprintf("dump_%s.toml", appName))
		s.Stop()
		return nil
	}

	// 读取文件
	err = cfg.ReadInConfig()
	if err != nil {
		// 不要求一定要有配置文件，使用环境变量也行. 但是有配置文件,读取出错不行.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Println(err)
			log.Printf("%#v\n", err)
			return err
		}
		log.Println("config file not exists,use default config.", configFile, err)
		err = nil
	} else {
		// 读取成功，监控文件变化
		cfg.WatchConfig()
		// 配置变动更新
		cfg.OnConfigChange(func(in fsnotify.Event) {
			svc.onUpdateConfig()
		})
	}

	// 初始化读取配置
	for _, vc := range svc.values {
		vc.Value.RefreshValue(svc)
	}

	return
}
func (svc *ConfigFileService) Start(s app.Stoper) error {
	return nil
}
func (svc *ConfigFileService) Stop() {
	return
}
func (svc *ConfigFileService) Finish() {
	return
}

func (svc *ConfigFileService) onUpdateConfig() {
	for _, vc := range svc.values {
		vc.Value.RefreshValue(svc)
		for _, ntf := range vc.Ntfs {
			ntf(svc)
		}
	}
	for _, ntf := range svc.ntfs {
		ntf(svc)
	}
	return
}

// register custom config value
func (svc *ConfigFileService) RegisterConfig(v configcentra.ConfigValue, ntf []configcentra.ConfigUpdateNotify) {
	svc.values = append(svc.values, ConfigItem{
		Value: v,
		Ntfs:  ntf,
	})
}

// watch config update
func (svc *ConfigFileService) WatchConfigUpdate(ntf []configcentra.ConfigUpdateNotify) {
	svc.ntfs = append(svc.ntfs, ntf...)
}

func (svc *ConfigFileService) SetDefault(key string, doc string, value interface{}) {
	// NOTE: config file ingore doc info.
	svc.ctrl.SetDefault(key, value)
}
func (svc *ConfigFileService) GetString(key string) string {
	return svc.ctrl.GetString(key)
}
func (svc *ConfigFileService) GetBool(key string) bool {
	return svc.ctrl.GetBool(key)
}
func (svc *ConfigFileService) GetInt(key string) int {
	return svc.ctrl.GetInt(key)
}
func (svc *ConfigFileService) GetInt32(key string) int32 {
	return svc.ctrl.GetInt32(key)
}
func (svc *ConfigFileService) GetInt64(key string) int64 {
	return svc.ctrl.GetInt64(key)
}
func (svc *ConfigFileService) GetUint(key string) uint {
	return svc.ctrl.GetUint(key)
}
func (svc *ConfigFileService) GetUint16(key string) uint16 {
	return svc.ctrl.GetUint16(key)
}
func (svc *ConfigFileService) GetUint32(key string) uint32 {
	return svc.ctrl.GetUint32(key)
}
func (svc *ConfigFileService) GetUint64(key string) uint64 {
	return svc.ctrl.GetUint64(key)
}
func (svc *ConfigFileService) GetFloat64(key string) float64 {
	return svc.ctrl.GetFloat64(key)
}
func (svc *ConfigFileService) GetTime(key string) time.Time {
	return svc.ctrl.GetTime(key)
}
func (svc *ConfigFileService) GetDuration(key string) time.Duration {
	return svc.ctrl.GetDuration(key)
}
func (svc *ConfigFileService) GetIntSlice(key string) []int {
	return svc.ctrl.GetIntSlice(key)
}
func (svc *ConfigFileService) GetStringSlice(key string) []string {
	return svc.ctrl.GetStringSlice(key)
}
