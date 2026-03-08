package input

import (
	"fmt"
	"strings"

	bkn "github.com/kweaver-ai/bkn-specification/sdk/golang/bkn"
)

func ParseDeleteTargets(values []string) ([]bkn.DeleteTarget, error) {
	targets := make([]bkn.DeleteTarget, 0, len(values))
	for _, value := range values {
		typ, id, ok := strings.Cut(value, ":")
		typ = strings.TrimSpace(typ)
		id = strings.TrimSpace(id)
		if !ok || typ == "" || id == "" {
			return nil, fmt.Errorf("invalid target %q (expected type:id)", value)
		}
		targets = append(targets, bkn.DeleteTarget{Type: typ, ID: id})
	}
	return targets, nil
}
