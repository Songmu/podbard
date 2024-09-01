package primcast

import "fmt"

type Categories []string

func (cats *Categories) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	var strSlice []string

	if err := unmarshal(&str); err == nil {
		*cats = []string{str}
		return nil
	}
	err := unmarshal(&strSlice)
	if err == nil {
		*cats = strSlice
		return nil
	}
	return fmt.Errorf("failed to unmarshal field into Categories: %w", err)
}
