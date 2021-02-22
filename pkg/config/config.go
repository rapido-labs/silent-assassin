package config

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Provider struct {
	*viper.Viper
}

type IProvider interface {
	GetString(key string) string
	GetBool(key string) bool
	GetInt(key string) int
	GetInt32(key string) int32
	GetInt64(key string) int64
	GetUint(key string) uint
	GetUint32(key string) uint32
	GetUint64(key string) uint64
	GetFloat64(key string) float64
	GetTime(key string) time.Time
	GetDuration(key string) time.Duration
	GetIntSlice(key string) []int
	GetStringSlice(key string) []string
	GetStringMap(key string) map[string]interface{}
	GetStringMapString(key string) map[string]string
	GetStringMapStringSlice(key string) map[string][]string
	GetSizeInBytes(key string) uint
	SplitStringToSlice(key string, sep string) []string
}

var fetcher Provider

func Init(cfgFile string) *Provider {

	v := viper.New()
	if cfgFile == "" {
		panic("Config file path was not provided!")
	}

	v.SetConfigFile(cfgFile)

	if err := v.ReadInConfig(); err != nil {

		panic(err)
	}

	return &Provider{Viper: v}

}

// InitValue creates a config provier that extends with in-memory values
func InitValue(values ...map[string]interface{}) *Provider {
	v := viper.New()
	for _, value := range values {
		if err := v.MergeConfigMap(value); err != nil {
			panic(errors.Wrapf(err, "merge value failed %+v", value))
		}
	}
	return &Provider{Viper: v}
}

func (f *Provider) GetString(key string) string {
	return f.Viper.GetString(key)
}
func (f *Provider) GetBool(key string) bool {
	return f.Viper.GetBool(key)
}
func (f *Provider) GetInt(key string) int {
	return f.Viper.GetInt(key)
}
func (f *Provider) GetInt32(key string) int32 {
	return f.Viper.GetInt32(key)
}
func (f *Provider) GetInt64(key string) int64 {
	return f.Viper.GetInt64(key)
}
func (f *Provider) GetUint(key string) uint {
	return f.Viper.GetUint(key)
}
func (f *Provider) GetUint32(key string) uint32 {
	return f.Viper.GetUint32(key)
}
func (f *Provider) GetUint64(key string) uint64 {
	return f.Viper.GetUint64(key)
}
func (f *Provider) GetFloat64(key string) float64 {
	return f.Viper.GetFloat64(key)
}
func (f *Provider) GetTime(key string) time.Time {
	return f.Viper.GetTime(key)
}
func (f *Provider) GetDuration(key string) time.Duration {
	return f.Viper.GetDuration(key)
}
func (f *Provider) GetIntSlice(key string) []int {
	return f.Viper.GetIntSlice(key)
}
func (f *Provider) GetStringSlice(key string) []string {
	return f.Viper.GetStringSlice(key)
}
func (f *Provider) GetStringMap(key string) map[string]interface{} {
	return f.Viper.GetStringMap(key)
}
func (f *Provider) GetStringMapString(key string) map[string]string {
	return f.Viper.GetStringMapString(key)
}
func (f *Provider) GetStringMapStringSlice(key string) map[string][]string {
	return f.Viper.GetStringMapStringSlice(key)
}
func (f *Provider) GetSizeInBytes(key string) uint {
	return f.Viper.GetSizeInBytes(key)
}
func (f *Provider) SplitStringToSlice(key string, sep string) []string {
	str := f.Viper.GetString(key)
	return strings.Split(str, sep)
}
