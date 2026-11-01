package cli

import (
	"fmt"
	"strconv"
	"strings"
)

func parsePositiveID(flagName, raw string) (uint64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, usageErr("%s is required", flagName)
	}
	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, usageErr("%s must be a positive base-10 integer", flagName)
	}
	if n == 0 {
		return 0, usageErr("%s must be greater than 0", flagName)
	}
	return n, nil
}

func projectPathFromID(projectID uint64, suffix string) (string, error) {
	seg := strconv.FormatUint(projectID, 10)
	p := fmt.Sprintf("/api/project/%s%s", seg, suffix)
	return normalizeAPIPath(p)
}

func projectTaskPath(projectID, taskID uint64, suffix string) (string, error) {
	p := fmt.Sprintf("/api/project/%s/tasks/%s%s",
		strconv.FormatUint(projectID, 10),
		strconv.FormatUint(taskID, 10),
		suffix,
	)
	return normalizeAPIPath(p)
}
