package config

import (
	"time"

	"github.com/stretchr/testify/mock"
)

type ProviderMock struct {
	mock.Mock
}

func (f *ProviderMock) GetString(key string) string {
	args := f.Called(key)
	return args.String(0)
}

func (f *ProviderMock) GetBool(key string) bool {
	args := f.Called(key)
	return args.Bool(0)
}
func (f *ProviderMock) GetInt(key string) int {
	args := f.Called(key)
	return args.Int(0)
}
func (f *ProviderMock) GetInt32(key string) int32 {
	args := f.Called(key)
	return args.Get(0).(int32)
}
func (f *ProviderMock) GetInt64(key string) int64 {
	args := f.Called(key)
	return args.Get(0).(int64)
}
func (f *ProviderMock) GetUint(key string) uint {
	args := f.Called(key)
	return args.Get(0).(uint)
}
func (f *ProviderMock) GetUint32(key string) uint32 {
	args := f.Called(key)
	return args.Get(0).(uint32)
}
func (f *ProviderMock) GetUint64(key string) uint64 {
	args := f.Called(key)
	return args.Get(0).(uint64)
}
func (f *ProviderMock) GetFloat64(key string) float64 {
	args := f.Called(key)
	return args.Get(0).(float64)
}
func (f *ProviderMock) GetTime(key string) time.Time {
	args := f.Called(key)
	return args.Get(0).(time.Time)
}
func (f *ProviderMock) GetDuration(key string) time.Duration {
	args := f.Called(key)
	return args.Get(0).(time.Duration)
}
func (f *ProviderMock) GetIntSlice(key string) []int {
	args := f.Called(key)
	return args.Get(0).([]int)
}
func (f *ProviderMock) GetStringSlice(key string) []string {
	args := f.Called(key)
	return args.Get(0).([]string)
}
func (f *ProviderMock) GetStringMap(key string) map[string]interface{} {
	args := f.Called(key)
	return args.Get(0).(map[string]interface{})
}
func (f *ProviderMock) GetStringMapString(key string) map[string]string {
	args := f.Called(key)
	return args.Get(0).(map[string]string)
}
func (f *ProviderMock) GetStringMapStringSlice(key string) map[string][]string {
	args := f.Called(key)
	return args.Get(0).(map[string][]string)
}
func (f *ProviderMock) GetSizeInBytes(key string) uint {
	args := f.Called(key)
	return args.Get(0).(uint)
}

func (f *ProviderMock) GetSliceByString(key string, sep string) []string {
	args := f.Called(key, sep)
	return args.Get(0).([]string)
}
