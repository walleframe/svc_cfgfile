package svc_cfgfile

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cast"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/walleframe/walle/app"
	"github.com/walleframe/walle/services/configcentra"
	"go.uber.org/multierr"
)

type ConfigItem struct {
	Value configcentra.ConfigValue
	Ntfs  []configcentra.ConfigUpdateNotify
}

type ConfigFileBackend struct {
	ctrl   *viper.Viper
	values []ConfigItem
	ntfs   []configcentra.ConfigUpdateNotify
}

var _ configcentra.ConfigCentra = (*ConfigFileBackend)(nil)

func (svc *ConfigFileBackend) Init(s app.Stoper) (err error) {
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
		err = multierr.Append(err, vc.Value.RefreshValue(svc))
	}

	return
}
func (svc *ConfigFileBackend) Start(s app.Stoper) error {
	return nil
}
func (svc *ConfigFileBackend) Stop() {
	return
}
func (svc *ConfigFileBackend) Finish() {
	return
}

func (svc *ConfigFileBackend) onUpdateConfig() {
	for _, vc := range svc.values {
		if err := vc.Value.RefreshValue(svc); err != nil {
			log.Println("update config failed", err)
			continue
		}
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
func (svc *ConfigFileBackend) RegisterConfig(v configcentra.ConfigValue, ntf []configcentra.ConfigUpdateNotify) {
	svc.values = append(svc.values, ConfigItem{
		Value: v,
		Ntfs:  ntf,
	})
}

// watch config update
func (svc *ConfigFileBackend) WatchConfigUpdate(ntf []configcentra.ConfigUpdateNotify) {
	svc.ntfs = append(svc.ntfs, ntf...)
}

// object support
func (svc *ConfigFileBackend) UseObject() bool {
	return false
}

func (svc *ConfigFileBackend) SetObject(key string, doc string, obj interface{}) {
	panic("not support set object")
}

func (svc *ConfigFileBackend) GetObject(key string, obj interface{}) (err error) {
	return errors.New("config file backend not support object")
}

func (svc *ConfigFileBackend) SetDefault(key string, doc string, value interface{}) {
	// NOTE: config file ingore doc info.
	svc.ctrl.SetDefault(key, value)
}

func (svc *ConfigFileBackend) GetString(key string) (string, error) {
	return cast.ToStringE(svc.ctrl.Get(key))
}
func (svc *ConfigFileBackend) GetBool(key string) (bool, error) {
	return cast.ToBoolE(svc.ctrl.Get(key))
}
func (svc *ConfigFileBackend) GetInt(key string) (int, error) {
	return cast.ToIntE(svc.ctrl.Get(key))
}
func (svc *ConfigFileBackend) GetInt32(key string) (int32, error) {
	return cast.ToInt32E(svc.ctrl.Get(key))
}
func (svc *ConfigFileBackend) GetInt64(key string) (int64, error) {
	return cast.ToInt64E(svc.ctrl.Get(key))
}
func (svc *ConfigFileBackend) GetUint(key string) (uint, error) {
	return cast.ToUintE(svc.ctrl.Get(key))
}
func (svc *ConfigFileBackend) GetUint16(key string) (uint16, error) {
	return cast.ToUint16E(svc.ctrl.Get(key))
}
func (svc *ConfigFileBackend) GetUint32(key string) (uint32, error) {
	return cast.ToUint32E(svc.ctrl.Get(key))
}
func (svc *ConfigFileBackend) GetUint64(key string) (uint64, error) {
	return cast.ToUint64E(svc.ctrl.Get(key))
}
func (svc *ConfigFileBackend) GetFloat64(key string) (float64, error) {
	return cast.ToFloat64E(svc.ctrl.Get(key))
}
func (svc *ConfigFileBackend) GetTime(key string) (time.Time, error) {
	return cast.ToTimeE(svc.ctrl.Get(key))
}
func (svc *ConfigFileBackend) GetDuration(key string) (time.Duration, error) {
	return cast.ToDurationE(svc.ctrl.Get(key))
}
func (svc *ConfigFileBackend) GetIntSlice(key string) ([]int, error) {
	return cast.ToIntSliceE(svc.ctrl.Get(key))
}
func (svc *ConfigFileBackend) GetStringSlice(key string) ([]string, error) {
	return cast.ToStringSliceE(svc.ctrl.Get(key))
}
