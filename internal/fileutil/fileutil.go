package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	markerStart = ">>> ezproxy >>>"
	markerEnd   = "<<< ezproxy <<<"
)

func startMarker(comment string) string {
	return fmt.Sprintf("%s %s", comment, markerStart)
}

func endMarker(comment string) string {
	return fmt.Sprintf("%s %s", comment, markerEnd)
}

func UpsertMarkerBlock(path string, content string, comment string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	existing := ""
	if data, err := os.ReadFile(path); err == nil {
		existing = string(data)
	}

	block := fmt.Sprintf("%s\n%s%s\n", startMarker(comment), content, endMarker(comment))

	start := startMarker(comment)
	end := endMarker(comment)

	startIdx := strings.Index(existing, start)
	endIdx := strings.Index(existing, end)

	var result string
	if startIdx >= 0 && endIdx >= 0 {
		result = existing[:startIdx] + block + existing[endIdx+len(end):]
		if strings.HasSuffix(result, "\n\n\n") {
			result = strings.TrimRight(result, "\n") + "\n"
		}
	} else {
		if existing != "" && !strings.HasSuffix(existing, "\n") {
			existing += "\n"
		}
		if existing != "" {
			existing += "\n"
		}
		result = existing + block
	}

	return os.WriteFile(path, []byte(result), 0644)
}

func RemoveMarkerBlock(path string, comment string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	content := string(data)
	start := startMarker(comment)
	end := endMarker(comment)

	startIdx := strings.Index(content, start)
	endIdx := strings.Index(content, end)

	if startIdx < 0 || endIdx < 0 {
		return nil
	}

	before := content[:startIdx]
	after := content[endIdx+len(end):]

	before = strings.TrimRight(before, "\n")
	after = strings.TrimLeft(after, "\n")

	result := ""
	if before != "" && after != "" {
		result = before + "\n" + after + "\n"
	} else if before != "" {
		result = before + "\n"
	} else if after != "" {
		result = after + "\n"
	}

	return os.WriteFile(path, []byte(result), 0644)
}

func HasMarkerBlock(path string, comment string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), startMarker(comment))
}

func GetMarkerBlockContent(path string, comment string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	content := string(data)
	start := startMarker(comment)
	end := endMarker(comment)

	startIdx := strings.Index(content, start)
	endIdx := strings.Index(content, end)

	if startIdx < 0 || endIdx < 0 {
		return "", fmt.Errorf("no marker block found in %s", path)
	}

	blockContent := content[startIdx+len(start)+1 : endIdx]
	return blockContent, nil
}
