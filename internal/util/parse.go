// Package util
package util

import (
    "mime"
    "strings"
    "path/filepath"
    "mime/multipart"
)

func GetFileName(url string) string {
    base := filepath.Base(url)
    return strings.TrimSuffix(base, filepath.Ext(base))
}

func GetFileExt(url string) string {
    return filepath.Ext(filepath.Base(url))
}

func GetContentType(header *multipart.FileHeader) string {
    contentType := header.Header.Get("Content-Type")
    if contentType == "" {
        return "application/octet-stream"
    }

    return contentType
}

func MakeContentType(url string) string {
    base := filepath.Base(url)

    contentType := mime.TypeByExtension(filepath.Ext(base))
    if contentType == "" {
        // 기본값이고, 바이너리이므로 다운로드를 유도한다.
        contentType = "application/octet-stream"
    }

    return contentType
}
