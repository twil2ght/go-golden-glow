package executor

import "fmt"

func Validate(params Parameters, keys ...string) error {
	for _, key := range keys {
		if _, ok := params[key]; !ok {
			return fmt.Errorf("key:%s not found", key)
		}
	}
	return nil
}
