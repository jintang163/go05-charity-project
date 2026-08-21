package receipt

import (
	"fmt"
	"strings"
	"time"
)

func Code(now time.Time, seq string) string {
	day := now.UTC().Format("20060102")
	seq = strings.ToUpper(strings.TrimSpace(seq))
	if len(seq) > 8 {
		seq = seq[:8]
	}
	if seq == "" {
		seq = "X"
	}
	return fmt.Sprintf("CP-%s-%s", day, seq)
}

func ValidFormat(code string) bool {
	code = strings.TrimSpace(code)
	if !strings.HasPrefix(code, "CP-") {
		return false
	}
	parts := strings.Split(code, "-")
	return len(parts) == 3 && len(parts[1]) == 8 && len(parts[2]) > 0
}
