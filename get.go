package sail

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

type GetError struct {
	Err   error
	Infos []string
}

func (g *GetError) Error() string {
	return fmt.Sprintf("%s: %v", g.Err, g.Infos)
}

var (
	ErrDuplicateKey = errors.New("ErrDuplicateKey")
)

func (s *Sail) Get(key string) (interface{}, error) {
	return s.rangeVipers(key)
}

func (s *Sail) MustGet(key string) interface{} {
	v, _ := s.Get(key)
	return v
}

func (s *Sail) GetWithName(key string, name string) interface{} {
	if v, ok := s.vipers[name]; ok {
		return v.Get(key)
	}
	return nil
}

func (s *Sail) GetString(key string) (string, error) {
	v, err := s.Get(key)
	return cast.ToString(v), err
}

func (s *Sail) MustGetString(key string) string {
	v, _ := s.Get(key)
	return cast.ToString(v)
}

func (s *Sail) GetStringWithName(key string, name string) string {
	if v, ok := s.vipers[name]; ok {
		return v.GetString(key)
	}
	return ""
}

func (s *Sail) GetBool(key string) (bool, error) {
	v, err := s.Get(key)
	return cast.ToBool(v), err
}

func (s *Sail) MustGetBool(key string) bool {
	v, _ := s.Get(key)
	return cast.ToBool(v)
}

func (s *Sail) GetBoolWithName(key string, name string) bool {
	if v, ok := s.vipers[name]; ok {
		return v.GetBool(key)
	}
	return false
}

func (s *Sail) GetInt(key string) (int, error) {
	v, err := s.Get(key)
	return cast.ToInt(v), err
}

func (s *Sail) MustGetInt(key string) int {
	v, _ := s.Get(key)
	return cast.ToInt(v)
}

func (s *Sail) GetIntWithName(key string, name string) int {
	if v, ok := s.vipers[name]; ok {
		return v.GetInt(key)
	}
	return 0
}

func (s *Sail) GetInt32(key string) (int32, error) {
	v, err := s.Get(key)
	return cast.ToInt32(v), err
}

func (s *Sail) MustGetInt32(key string) int32 {
	v, _ := s.Get(key)
	return cast.ToInt32(v)
}

func (s *Sail) GetInt32WithName(key string, name string) int32 {
	if v, ok := s.vipers[name]; ok {
		return v.GetInt32(key)
	}
	return 0
}

func (s *Sail) GetInt64(key string) (int64, error) {
	v, err := s.Get(key)
	return cast.ToInt64(v), err
}

func (s *Sail) MustGetInt64(key string) int64 {
	v, _ := s.Get(key)
	return cast.ToInt64(v)
}

func (s *Sail) GetInt64WithName(key string, name string) int64 {
	if v, ok := s.vipers[name]; ok {
		return v.GetInt64(key)
	}
	return 0
}

func (s *Sail) GetUint(key string) (uint, error) {
	v, err := s.Get(key)
	return cast.ToUint(v), err
}

func (s *Sail) MustGetUint(key string) uint {
	v, _ := s.Get(key)
	return cast.ToUint(v)
}

func (s *Sail) GetUintWithName(key string, name string) uint {
	if v, ok := s.vipers[name]; ok {
		return v.GetUint(key)
	}
	return 0
}

func (s *Sail) GetFloat64(key string) (float64, error) {
	v, err := s.Get(key)
	return cast.ToFloat64(v), err
}

func (s *Sail) MustGetFloat64(key string) float64 {
	v, _ := s.Get(key)
	return cast.ToFloat64(v)
}

func (s *Sail) GetFloat64WithName(key string, name string) float64 {
	if v, ok := s.vipers[name]; ok {
		return v.GetFloat64(key)
	}
	return 0
}

func (s *Sail) GetTime(key string) (time.Time, error) {
	v, err := s.Get(key)
	return cast.ToTime(v), err
}

func (s *Sail) MustGetTime(key string) time.Time {
	v, _ := s.Get(key)
	return cast.ToTime(v)
}

func (s *Sail) GetTimeWithName(key string, name string) time.Time {
	if v, ok := s.vipers[name]; ok {
		return v.GetTime(key)
	}
	return time.Time{}
}

func (s *Sail) GetDuration(key string) (time.Duration, error) {
	v, err := s.Get(key)
	return cast.ToDuration(v), err
}

func (s *Sail) MustGetDuration(key string) time.Duration {
	v, _ := s.Get(key)
	return cast.ToDuration(v)
}

func (s *Sail) GetDurationWithName(key string, name string) time.Duration {
	if v, ok := s.vipers[name]; ok {
		return v.GetDuration(key)
	}
	return 0
}

func (s *Sail) GetIntSlice(key string) ([]int, error) {
	v, err := s.Get(key)
	return cast.ToIntSlice(v), err
}

func (s *Sail) MustGetIntSlice(key string) []int {
	v, _ := s.Get(key)
	return cast.ToIntSlice(v)
}

func (s *Sail) GetIntSliceWithName(key string, name string) []int {
	if v, ok := s.vipers[name]; ok {
		return v.GetIntSlice(key)
	}
	return nil
}

func (s *Sail) GetStringSlice(key string) ([]string, error) {
	v, err := s.Get(key)
	return cast.ToStringSlice(v), err
}

func (s *Sail) MustGetStringSlice(key string) []string {
	v, _ := s.Get(key)
	return cast.ToStringSlice(v)
}

func (s *Sail) GetStringSliceWithName(key string, name string) []string {
	if v, ok := s.vipers[name]; ok {
		return v.GetStringSlice(key)
	}
	return nil
}

func (s *Sail) GetStringMap(key string) (map[string]interface{}, error) {
	v, err := s.Get(key)
	return cast.ToStringMap(v), err
}

func (s *Sail) MustGetStringMap(key string) map[string]interface{} {
	v, _ := s.Get(key)
	return cast.ToStringMap(v)
}

func (s *Sail) GetStringMapWithName(key string, name string) map[string]interface{} {
	if v, ok := s.vipers[name]; ok {
		return v.GetStringMap(key)
	}
	return nil
}

func (s *Sail) GetStringMapString(key string) (map[string]string, error) {
	v, err := s.Get(key)
	return cast.ToStringMapString(v), err
}

func (s *Sail) MustGetStringMapString(key string) map[string]string {
	v, _ := s.Get(key)
	return cast.ToStringMapString(v)
}

func (s *Sail) GetStringMapStringWithName(key string, name string) map[string]string {
	if v, ok := s.vipers[name]; ok {
		return v.GetStringMapString(key)
	}
	return nil
}

func (s *Sail) GetStringMapStringSlice(key string) (map[string][]string, error) {
	v, err := s.Get(key)
	return cast.ToStringMapStringSlice(v), err
}

func (s *Sail) MustGetStringMapStringSlice(key string) map[string][]string {
	v, _ := s.Get(key)
	return cast.ToStringMapStringSlice(v)
}

func (s *Sail) GetStringMapStringSliceWithName(key string, name string) map[string][]string {
	if v, ok := s.vipers[name]; ok {
		return v.GetStringMapStringSlice(key)
	}
	return nil
}

func (s *Sail) GetSizeInBytes(key string) (uint, error) {
	v, err := s.GetString(key)
	return parseSizeInBytes(v), err
}

func (s *Sail) MustGetSizeInBytes(key string) uint {
	v, _ := s.GetString(key)
	return parseSizeInBytes(v)
}

func (s *Sail) GetSizeInBytesWithName(key string, name string) uint {
	if v, ok := s.vipers[name]; ok {
		return v.GetSizeInBytes(key)
	}
	return 0
}

func (s *Sail) GetViperWithName(name string) *viper.Viper {
	v, ok := s.vipers[name]
	if !ok {
		return nil
	}
	return v
}

func (s *Sail) MergeVipers() (*viper.Viper, error) {
	newViper := viper.New()
	s.lock.RLock()
	defer s.lock.RUnlock()

	for _, v := range s.vipers {
		err := newViper.MergeConfigMap(v.AllSettings())
		if err != nil {
			return nil, err
		}
	}
	return newViper, nil
}

func (s *Sail) MergeVipersWithName() (*viper.Viper, error) {
	newViper := viper.New()
	s.lock.RLock()
	defer s.lock.RUnlock()

	for k, v := range s.vipers {
		dataMap := v.AllSettings()

		err := newViper.MergeConfigMap(map[string]interface{}{
			k: dataMap,
		})
		if err != nil {
			return nil, err
		}
	}
	return newViper, nil
}

func (s *Sail) rangeVipers(key string) (interface{}, error) {
	if len(s.vipers) == 0 {
		return nil, nil
	}
	var result interface{}
	var duResult []string

	s.lock.RLock()
	for k, v := range s.vipers {
		if ok := v.IsSet(key); ok {
			vv := v.Get(key)
			if result == nil {
				result = vv
			}
			duResult = append(duResult, k)
		}
	}
	s.lock.RUnlock()

	if len(duResult) > 1 {
		return result, &GetError{
			Err:   ErrDuplicateKey,
			Infos: duResult,
		}
	}
	return result, nil
}

// parseSizeInBytes converts strings like 1GB or 12 mb into an unsigned integer number of bytes
func parseSizeInBytes(sizeStr string) uint {
	sizeStr = strings.TrimSpace(sizeStr)
	lastChar := len(sizeStr) - 1
	multiplier := uint(1)

	if lastChar > 0 {
		if sizeStr[lastChar] == 'b' || sizeStr[lastChar] == 'B' {
			if lastChar > 1 {
				switch unicode.ToLower(rune(sizeStr[lastChar-1])) {
				case 'k':
					multiplier = 1 << 10
					sizeStr = strings.TrimSpace(sizeStr[:lastChar-1])
				case 'm':
					multiplier = 1 << 20
					sizeStr = strings.TrimSpace(sizeStr[:lastChar-1])
				case 'g':
					multiplier = 1 << 30
					sizeStr = strings.TrimSpace(sizeStr[:lastChar-1])
				default:
					multiplier = 1
					sizeStr = strings.TrimSpace(sizeStr[:lastChar])
				}
			}
		}
	}

	size := cast.ToInt(sizeStr)
	if size < 0 {
		size = 0
	}

	return safeMul(uint(size), multiplier)
}

func safeMul(a, b uint) uint {
	c := a * b
	if a > 1 && b > 1 && c/b != a {
		return 0
	}
	return c
}
