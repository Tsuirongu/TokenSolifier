package services

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// UploadTempFile 上传文件到临时目录
func UploadTempFile(fileName string, base64Content string) (string, error) {
	// 创建临时文件目录
	tempDir := filepath.Join(os.TempDir(), "loji-uploads")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("创建临时目录失败: %w", err)
	}

	// 生成唯一文件名
	timestamp := time.Now().UnixNano()
	ext := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, ext)
	uniqueFileName := fmt.Sprintf("%d_%s%s", timestamp, baseName, ext)
	filePath := filepath.Join(tempDir, uniqueFileName)

	// 解码Base64内容
	decoded, err := base64.StdEncoding.DecodeString(base64Content)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %w", err)
	}

	// 写入文件
	if err := ioutil.WriteFile(filePath, decoded, 0644); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	log.Printf("文件已保存到临时目录: %s (大小: %d bytes)", filePath, len(decoded))
	return filePath, nil
}

// CleanupTempFiles 清理临时文件
func CleanupTempFiles(filePaths []string) error {
	var errs []string
	for _, path := range filePaths {
		// 只清理在临时目录中的文件
		if !strings.Contains(path, "loji-uploads") {
			log.Printf("跳过非临时文件: %s", path)
			continue
		}

		if err := os.Remove(path); err != nil {
			log.Printf("删除临时文件失败 %s: %v", path, err)
			errs = append(errs, fmt.Sprintf("%s: %v", path, err))
		} else {
			log.Printf("已删除临时文件: %s", path)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("部分文件删除失败: %s", strings.Join(errs, "; "))
	}
	return nil
}

// CleanupOldTempFiles 清理超过指定时间的临时文件
func CleanupOldTempFiles(maxAge time.Duration) error {
	tempDir := filepath.Join(os.TempDir(), "loji-uploads")
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		return nil // 目录不存在，无需清理
	}

	files, err := ioutil.ReadDir(tempDir)
	if err != nil {
		return fmt.Errorf("读取临时目录失败: %w", err)
	}

	now := time.Now()
	var errs []string
	count := 0

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// 检查文件修改时间
		if now.Sub(file.ModTime()) > maxAge {
			filePath := filepath.Join(tempDir, file.Name())
			if err := os.Remove(filePath); err != nil {
				errs = append(errs, fmt.Sprintf("%s: %v", file.Name(), err))
			} else {
				count++
			}
		}
	}

	log.Printf("清理了 %d 个过期临时文件", count)

	if len(errs) > 0 {
		return fmt.Errorf("部分文件清理失败: %s", strings.Join(errs, "; "))
	}
	return nil
}

// GetOutputFile 读取输出文件内容
func GetOutputFile(filePath string) ([]byte, error) {
	// 安全检查：只允许读取临时目录下的文件
	tempDir := os.TempDir()
	if !strings.HasPrefix(filePath, tempDir) && !strings.Contains(filePath, "loji-outputs") {
		return nil, fmt.Errorf("只能访问临时目录中的文件")
	}

	// 读取文件
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	return content, nil
}

// GetOutputFileBase64 读取输出文件并返回Base64编码
func GetOutputFileBase64(filePath string) (string, error) {
	content, err := GetOutputFile(filePath)
	if err != nil {
		return "", err
	}

	// 获取文件扩展名并推断MIME类型
	ext := strings.ToLower(filepath.Ext(filePath))
	mimeType := getMimeType(ext)

	// 编码为Base64并添加data URL前缀
	base64Content := base64.StdEncoding.EncodeToString(content)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64Content), nil
}

// CopyOutputFile 将输出文件复制到目标路径
func CopyOutputFile(sourcePath, targetPath string) error {
	// 安全检查：只允许复制临时目录下的文件
	tempDir := os.TempDir()
	if !strings.HasPrefix(sourcePath, tempDir) && !strings.Contains(sourcePath, "loji-outputs") {
		return fmt.Errorf("只能复制临时目录中的文件")
	}

	// 读取源文件
	content, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("读取源文件失败: %w", err)
	}

	// 确保目标目录存在
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 写入目标文件
	if err := ioutil.WriteFile(targetPath, content, 0644); err != nil {
		return fmt.Errorf("写入目标文件失败: %w", err)
	}

	log.Printf("文件已复制: %s -> %s", sourcePath, targetPath)
	return nil
}

// CleanupOutputFiles 清理输出文件
func CleanupOutputFiles(filePaths []string) error {
	tempDir := os.TempDir()
	var errs []string
	for _, path := range filePaths {
		// 只清理在临时目录中的文件
		if !strings.HasPrefix(path, tempDir) && !strings.Contains(path, "loji-outputs") {
			log.Printf("跳过非临时文件: %s", path)
			continue
		}

		if err := os.Remove(path); err != nil {
			log.Printf("删除输出文件失败 %s: %v", path, err)
			errs = append(errs, fmt.Sprintf("%s: %v", path, err))
		} else {
			log.Printf("已删除输出文件: %s", path)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("部分文件删除失败: %s", strings.Join(errs, "; "))
	}
	return nil
}

// getMimeType 根据文件扩展名获取MIME类型
func getMimeType(ext string) string {
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".bmp":  "image/bmp",
		".svg":  "image/svg+xml",
		".pdf":  "application/pdf",
		".txt":  "text/plain",
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".xml":  "application/xml",
		".zip":  "application/zip",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".mp4":  "video/mp4",
		".avi":  "video/x-msvideo",
	}

	if mimeType, ok := mimeTypes[ext]; ok {
		return mimeType
	}
	return "application/octet-stream"
}
