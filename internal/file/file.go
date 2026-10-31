package filex

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func MustLoad(filename string, v any) error {
	content, err := os.ReadFile(filename)

	if err != nil {
		return fmt.Errorf("read file error: %w", err)
	}
	err = yaml.Unmarshal(content, v)

	if err != nil {
		return fmt.Errorf("parse config content error: %w", err)
	}
	return nil
}
