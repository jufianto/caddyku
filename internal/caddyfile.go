package internal

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

const beginPrefix = "# BEGIN caddyku:"
const endPrefix = "# END caddyku:"

func caddyfileBlock(label string, entries []DomainEntry) string {
	var sb strings.Builder
	sb.WriteString(beginPrefix + label + "\n")
	for _, e := range entries {
		sb.WriteString(fmt.Sprintf("%s {\n    reverse_proxy %s\n}\n\n", e.Domain, e.Upstream))
	}
	sb.WriteString(endPrefix + label + "\n")
	return sb.String()
}

func ReadCaddyfile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func WriteCaddyfile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func AddBlock(path, label string, entries []DomainEntry) error {
	content, err := ReadCaddyfile(path)
	if err != nil {
		return err
	}
	begin := beginPrefix + label
	if strings.Contains(content, begin) {
		content = replaceBlock(content, label, entries)
	} else {
		if len(content) > 0 && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		if len(content) > 0 {
			content += "\n"
		}
		content += caddyfileBlock(label, entries)
	}
	return WriteCaddyfile(path, content)
}

func RemoveBlock(path, label string) error {
	content, err := ReadCaddyfile(path)
	if err != nil {
		return err
	}
	begin := beginPrefix + label
	if !strings.Contains(content, begin) {
		return fmt.Errorf("no managed block found for label %q", label)
	}
	content = deleteBlock(content, label)
	return WriteCaddyfile(path, content)
}

func SyncBlocks(path string, blocks map[string][]DomainEntry) error {
	content, err := ReadCaddyfile(path)
	if err != nil {
		return err
	}

	existingLabels := findBlockLabels(content)
	for _, label := range existingLabels {
		if _, ok := blocks[label]; !ok {
			content = deleteBlock(content, label)
		}
	}

	for label, entries := range blocks {
		begin := beginPrefix + label
		if strings.Contains(content, begin) {
			content = replaceBlock(content, label, entries)
		} else {
			if len(content) > 0 && !strings.HasSuffix(content, "\n") {
				content += "\n"
			}
			if len(content) > 0 {
				content += "\n"
			}
			content += caddyfileBlock(label, entries)
		}
	}

	return WriteCaddyfile(path, content)
}

func ListDomains(path string) ([]string, error) {
	content, err := ReadCaddyfile(path)
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`(?m)^(\S+)\s*\{`)
	matches := re.FindAllStringSubmatch(content, -1)
	var domains []string
	for _, m := range matches {
		domains = append(domains, m[1])
	}
	return domains, nil
}

func replaceBlock(content, label string, entries []DomainEntry) string {
	begin := beginPrefix + label
	end := endPrefix + label
	startIdx := strings.Index(content, begin)
	endIdx := strings.Index(content, end)
	if startIdx == -1 || endIdx == -1 {
		return content
	}
	endIdx += len(end)
	if endIdx < len(content) && content[endIdx] == '\n' {
		endIdx++
	}
	return content[:startIdx] + caddyfileBlock(label, entries) + content[endIdx:]
}

func deleteBlock(content, label string) string {
	begin := beginPrefix + label
	end := endPrefix + label
	startIdx := strings.Index(content, begin)
	endIdx := strings.Index(content, end)
	if startIdx == -1 || endIdx == -1 {
		return content
	}
	endIdx += len(end)
	if endIdx < len(content) && content[endIdx] == '\n' {
		endIdx++
	}
	if startIdx > 0 && content[startIdx-1] == '\n' {
		startIdx--
	}
	return content[:startIdx] + content[endIdx:]
}

func findBlockLabels(content string) []string {
	re := regexp.MustCompile(`(?m)^# BEGIN caddyku:(.+)$`)
	matches := re.FindAllStringSubmatch(content, -1)
	var labels []string
	for _, m := range matches {
		labels = append(labels, strings.TrimSpace(m[1]))
	}
	return labels
}
