package config

import (
	"time"

	"github.com/spf13/viper"
)

type Provider struct {
	viper *viper.Viper
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

	return &Provider{viper: v}

}

func (f *Provider) GetString(key string) string {
	return f.viper.GetString(key)
}
func (f *Provider) GetBool(key string) bool {
	return f.viper.GetBool(key)
}
func (f *Provider) GetInt(key string) int {
	return f.viper.GetInt(key)
}
func (f *Provider) GetInt32(key string) int32 {
	return f.viper.GetInt32(key)
}
func (f *Provider) GetInt64(key string) int64 {
	return f.viper.GetInt64(key)
}
func (f *Provider) GetUint(key string) uint {
	return f.viper.GetUint(key)
}
func (f *Provider) GetUint32(key string) uint32 {
	return f.viper.GetUint32(key)
}
func (f *Provider) GetUint64(key string) uint64 {
	return f.viper.GetUint64(key)
}
func (f *Provider) GetFloat64(key string) float64 {
	return f.viper.GetFloat64(key)
}
func (f *Provider) GetTime(key string) time.Time {
	return f.viper.GetTime(key)
}
func (f *Provider) GetDuration(key string) time.Duration {
	return f.viper.GetDuration(key)
}
func (f *Provider) GetIntSlice(key string) []int {
	return f.viper.GetIntSlice(key)
}
func (f *Provider) GetStringSlice(key string) []string {
	return f.viper.GetStringSlice(key)
}
func (f *Provider) GetStringMap(key string) map[string]interface{} {
	return f.viper.GetStringMap(key)
}
func (f *Provider) GetStringMapString(key string) map[string]string {
	return f.viper.GetStringMapString(key)
}
func (f *Provider) GetStringMapStringSlice(key string) map[string][]string {
	return f.viper.GetStringMapStringSlice(key)
}
func (f *Provider) GetSizeInBytes(key string) uint {
	return f.viper.GetSizeInBytes(key)
}
