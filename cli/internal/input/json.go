package input

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func ReadJSON(path string, target any) error {
	var reader io.Reader
	if path == "-" {
		reader = os.Stdin
	} else {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		reader = f
	}
	if err := json.NewDecoder(reader).Decode(target); err != nil {
		return fmt.Errorf("decode json from %q: %w", path, err)
	}
	return nil
}
