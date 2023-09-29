package config

import (
	"fmt"
	"strconv"
	"strings"
)

// SetEnvToParamIfNeed parses and assign string val to pointer of 'param' hidden under interface type.
// In case of unknown type returns error.
func SetEnvToParamIfNeed(param interface{}, val string) error {
	if val == "" {
		return nil
	}

	switch param := param.(type) {
	case *int:
		if envVal, err := strconv.Atoi(val); err == nil {
			*param = envVal
		} else {
			return err
		}
	case *int64:
		if envVal, err := Atoi64(val); err == nil {
			*param = envVal
		} else {
			return err
		}
	case *string:
		*param = val
	case *[]string:
		*param = strings.Split(val, ",")
	default:
		return fmt.Errorf("wrong input param type")
	}

	return nil
}

// Atoi64 converts string into int64.
func Atoi64(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

//goland:noinspection HttpUrlsUsage
//func AddHTTPPrefixIfNeed(value string) string {
//	if !strings.HasPrefix(value, "http://") {
//		return "http://" + value
//	}
//
//	return value
//}
