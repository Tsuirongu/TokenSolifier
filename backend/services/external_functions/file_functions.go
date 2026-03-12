package external_functions

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ReadFile 读取文件内容(支持Base64或文件路径)
// 参数 fileData 可以是:
//  1. Base64数据: "data:mime/type;base64,xxxxx"
//  2. 文件路径: "/path/to/file"
//
// 返回文件的字节内容
func ReadFile(fileData string) ([]byte, error) {
	if fileData == "" {
		return nil, fmt.Errorf("fileData is empty")
	}

	// 判断是Base64还是文件路径
	if strings.HasPrefix(fileData, "data:") {
		// Base64数据
		parts := strings.SplitN(fileData, ",", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid Base64 data format")
		}
		return base64.StdEncoding.DecodeString(parts[1])
	} else {
		// 文件路径
		return ioutil.ReadFile(fileData)
	}
}

// ReadFileAsString 读取文件为字符串
// 适用于文本文件
func ReadFileAsString(fileData string) (string, error) {
	content, err := ReadFile(fileData)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// GetFileExtension 获取文件扩展名
// 从fileData中提取文件扩展名
func GetFileExtension(fileData string) string {
	if strings.HasPrefix(fileData, "data:") {
		// 从MIME类型中提取
		parts := strings.Split(fileData, ";")
		if len(parts) > 0 {
			mimeType := strings.TrimPrefix(parts[0], "data:")
			// 简单映射常见MIME类型到扩展名
			switch mimeType {
			case "image/jpeg":
				return ".jpg"
			case "image/png":
				return ".png"
			case "image/gif":
				return ".gif"
			case "text/plain":
				return ".txt"
			case "application/pdf":
				return ".pdf"
			default:
				return ""
			}
		}
	} else {
		// 从文件路径中提取
		lastDot := strings.LastIndex(fileData, ".")
		if lastDot != -1 {
			return fileData[lastDot:]
		}
	}
	return ""
}

// GetFileSize 获取文件大小(字节)
func GetFileSize(fileData string) (int, error) {
	content, err := ReadFile(fileData)
	if err != nil {
		return 0, err
	}
	return len(content), nil
}

// WriteOutputFile 将字节数据写入输出文件
// 参数:
//
//	data: 要写入的字节数据
//	fileName: 文件名（可选，如果为空则自动生成）
//
// 返回: 文件的访问路径
func WriteOutputFile(data []byte, fileName string) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("data is empty")
	}

	// 创建输出文件目录
	outputDir := filepath.Join(os.TempDir(), "loji-outputs")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 生成文件名
	if fileName == "" {
		fileName = fmt.Sprintf("output_%d", time.Now().UnixNano())
	}

	// 生成唯一文件名（添加时间戳避免冲突）
	ext := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, ext)
	timestamp := time.Now().UnixNano()
	uniqueFileName := fmt.Sprintf("%s_%d%s", baseName, timestamp, ext)
	filePath := filepath.Join(outputDir, uniqueFileName)

	// 写入文件
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	return filePath, nil
}

// WriteOutputFileFromBase64 从Base64字符串写入输出文件
// 参数:
//
//	base64Data: Base64编码的数据（可以带或不带data:前缀）
//	fileName: 文件名
//
// 返回: 文件的访问路径
func WriteOutputFileFromBase64(base64Data string, fileName string) (string, error) {
	// 移除可能的data:前缀
	if strings.HasPrefix(base64Data, "data:") {
		parts := strings.SplitN(base64Data, ",", 2)
		if len(parts) == 2 {
			base64Data = parts[1]
		}
	}

	// 解码Base64
	decoded, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %w", err)
	}

	return WriteOutputFile(decoded, fileName)
}
