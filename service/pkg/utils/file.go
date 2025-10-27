package utils

import (
	"net/url"
	"path"
)

func GetFilenameFromURL(rawURL string) string {
	if rawURL == "" {
		return ""
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	// 直接从路径中提取文件名
	filename := path.Base(parsedURL.Path)

	// 处理无文件名的情况（如以斜杠结尾的路径）
	if filename == "." || filename == "/" {
		return ""
	}

	return filename
}
