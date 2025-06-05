package helper

import (
	"fmt"
	"strings"
	"time"
)

func BuildUpdatePartialQuery(updates map[string]interface{}) (string, []interface{}) {
	var setClauses []string
	var args []interface{}
	var updatedAt time.Time

	for key, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", key, len(args)+1))
		args = append(args, value)
	}

	updatedAt = time.Now()
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", len(args)+1))
	args = append(args, updatedAt)

	return strings.Join(setClauses, ", "), args
}
