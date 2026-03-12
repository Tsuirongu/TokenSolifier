package services

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFormatInputJSON(t *testing.T) {
	service := &YaegiService{}

	// 测试基本JSON
	input := `{"textList": ["Hello World"]}`
	expected := "`{\"textList\": [\"Hello World\"]}`"

	result := service.formatInputJSON(input)

	if result != expected {
		t.Errorf("formatInputJSON failed.\nExpected: %s\nGot: %s", expected, result)
	}

	// 验证生成的Go字符串在JSON解析时是有效的
	innerJSON := strings.Trim(result, "`")
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(innerJSON), &parsed); err != nil {
		t.Errorf("Formatted JSON is not valid: %v\nResult: %s\nInner: %s", err, result, innerJSON)
	}
}

func TestFormatInputJSONWithQuotes(t *testing.T) {
	service := &YaegiService{}

	// 测试包含双引号的JSON
	// 前端发送: {"textList": ["Hello \"World\""]}
	input := `{"textList": ["Hello \"World\""]}`

	// 应该生成反引号字符串，内部不需要额外转义双引号
	expected := "`{\"textList\": [\"Hello \\\"World\\\"\"]}`"

	result := service.formatInputJSON(input)

	if result != expected {
		t.Errorf("formatInputJSON failed.\nExpected: %s\nGot: %s", expected, result)
	}

	// 验证生成的Go字符串在JSON解析时是有效的
	innerJSON := strings.Trim(result, "`")
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(innerJSON), &parsed); err != nil {
		t.Errorf("Formatted JSON is not valid: %v\nResult: %s\nInner: %s", err, result, innerJSON)
	}

	// 验证解析结果
	if textList, ok := parsed["textList"].([]interface{}); ok {
		if len(textList) > 0 {
			if textList[0] != "Hello \"World\"" {
				t.Errorf("Expected first text to be 'Hello \"World\"', got %v", textList[0])
			}
		}
	}
}
