package aaronsql

import "strings"

var defaultModelDBTagKey = "db"

func SetDefaultModelDBTagKey(key string) {
	if key == "" {
		panic("default model db tag key cannot be empty")
	}
	defaultModelDBTagKey = key
}

func parseTagString(tagStr string) map[string]string {
	if tagStr == "" {
		return nil
	}
	tagMap := make(map[string]string)
	parts := strings.Split(tagStr, TAG_DEFAULT_PART_QUOTE)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if kv := strings.SplitN(part, TAG_DEFAULT_KEY_VALUE_QUOTE, 2); len(kv) == 2 {
			tagMap[kv[0]] = kv[1]
		} else if len(kv) == 1 {
			tagMap[kv[0]] = ""
		}
	}
	return tagMap
}
