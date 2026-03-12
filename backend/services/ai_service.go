package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"loji-app/backend/models"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// PluginGenerationResult 插件生成结果
type PluginGenerationResult struct {
	Code           string                       `json:"code"`
	InputDesc      models.PluginInputDesc       `json:"inputDesc"`
	OutputDesc     models.PluginOutputDesc      `json:"outputDesc"`
	Dependencies   []models.PluginDependency    `json:"dependencies"`
	RelatedPlugins []models.PluginRelatedPlugin `json:"relatedPlugins"`
	Names          PluginNames                  `json:"names"`
}

// PluginNames 插件名称
type PluginNames struct {
	ChineseName string `json:"chineseName"`
	EnglishName string `json:"englishName"`
}

// PluginInputOutputResult 输入输出描述和名称生成结果
type PluginInputOutputResult struct {
	InputDesc  models.PluginInputDesc  `json:"input"`
	OutputDesc models.PluginOutputDesc `json:"output"`
	Names      PluginNames             `json:"names"`
}

// ProgressCallback 进度回调函数类型
type ProgressCallback func(stage string, progress int, message string)

// AIService AI服务
type AIService struct {
	apiKey               string
	apiURL               string
	ioDescPromptTemplate string
	codePromptTemplate   string
	progressCallback     ProgressCallback // 进度回调函数
	configService        *ConfigService   // 配置服务
}

// NewAIService 创建AI服务实例
func NewAIService() *AIService {
	// 创建配置服务实例
	configService := NewConfigService()

	// 从配置服务读取API配置，如果配置为空则使用环境变量作为后备
	var apiKey, apiURL string

	// 获取API Key
	if configItem, err := configService.GetConfig("OPENAI_API_KEY"); err == nil && configItem.Value != "" {
		apiKey = configItem.Value
	} else {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	// 获取API URL
	if configItem, err := configService.GetConfig("OPENAI_API_URL"); err == nil && configItem.Value != "" {
		apiURL = configItem.Value
	} else {
		apiURL = os.Getenv("OPENAI_API_URL")
		if apiURL == "" {
			apiURL = "https://api.openai.com/v1/chat/completions"
		}
	}

	// 加载输入输出描述生成prompt模板
	ioDescPromptTemplate := loadIoDescPromptTemplate()

	// 加载代码生成prompt模板
	codePromptTemplate := loadCodePromptTemplate()

	return &AIService{
		apiKey:               apiKey,
		apiURL:               apiURL,
		ioDescPromptTemplate: ioDescPromptTemplate,
		codePromptTemplate:   codePromptTemplate,
		progressCallback:     nil, // 默认无进度回调
		configService:        configService,
	}
}

// SetProgressCallback 设置进度回调函数
func (s *AIService) SetProgressCallback(callback ProgressCallback) {
	s.progressCallback = callback
}

// getConfigValue 获取配置值，如果配置为空则返回环境变量作为后备
func (s *AIService) getConfigValue(key string) string {
	if configItem, err := s.configService.GetConfig(key); err == nil && configItem.Value != "" {
		return configItem.Value
	}
	return os.Getenv(key)
}

// PluginGenerationConfig 插件生成配置（从app包导入，但这里重新定义避免循环依赖）
type PluginGenerationConfig struct {
	UseAvailablePlugins bool // 是否使用其他插件
	UseAICapabilities   bool // 是否使用AI能力
}

// GeneratePluginCode 生成插件代码和输入输出描述（包含插件筛选）
func (s *AIService) GeneratePluginCode(requirement string, availablePlugins []models.Plugin, genConfig *PluginGenerationConfig) (*PluginGenerationResult, error) {
	// 如果API key未设置，尝试从配置服务获取
	if s.apiKey == "" {
		s.apiKey = s.getConfigValue("OPENAI_API_KEY")
		if s.apiKey == "" {
			return nil, fmt.Errorf("API key is not set")
		}
	}

	// 第一步：AI筛选相关插件
	selectedPlugins, err := s.selectRelevantPlugins(requirement, availablePlugins)
	if err != nil {
		log.Printf("Warning: failed to select relevant plugins: %v, continuing without related plugins", err)
		selectedPlugins = []models.Plugin{}
	}

	// 进度报告：输入输出格式化阶段
	if s.progressCallback != nil {
		s.progressCallback("输入输出格式化", 50, "正在通过AI生成输入输出描述和插件名称...")
	}
	log.Printf("AIService: Sending progress event: 输入输出格式化 50%%")

	// 第二步：生成输入输出描述和名称（独立于插件筛选）
	inputDesc, outputDesc, names, err := s.generateInputOutputDescByAI(requirement)
	if err != nil {
		return nil, fmt.Errorf("failed to generate input/output description: %w", err)
	}

	// 进度报告：代码生成阶段
	if s.progressCallback != nil {
		s.progressCallback("代码生成", 80, "正在通过AI生成插件代码...")
	}
	log.Printf("AIService: Sending progress event: 代码生成 80%%")

	// 第三步：根据输入输出描述和相关插件生成代码
	code, err := s.generateCodeByAI(requirement, inputDesc, outputDesc, selectedPlugins, genConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plugin code: %w", err)
	}

	// 进度报告：依赖提取阶段
	if s.progressCallback != nil {
		s.progressCallback("依赖提取", 90, "正在通过AI分析代码依赖...")
	}
	log.Printf("AIService: Sending progress event: 依赖提取 90%%")

	// 第四步：通过AI分析代码中的依赖
	dependencies, err := s.extractDependenciesByAI(code)
	if err != nil {
		log.Printf("Warning: failed to extract dependencies by AI: %v, falling back to string matching", err)
		// 如果AI提取失败，回退到原来的字符串匹配方法
		dependencies = s.extractDependenciesFromCode(code)
	}

	// 第五步：转换相关插件为PluginRelatedPlugin格式
	relatedPlugins := make([]models.PluginRelatedPlugin, len(selectedPlugins))
	for i, plugin := range selectedPlugins {
		relatedPlugins[i] = models.PluginRelatedPlugin{
			RelatedPluginID:   plugin.ID,
			RelatedPluginName: plugin.Name,
		}
	}

	// 第六步：验证和修复生成的代码
	validatedCode, err := s.validateAndFixCode(code, inputDesc, outputDesc, selectedPlugins, genConfig, dependencies)
	if err != nil {
		return nil, fmt.Errorf("failed to validate and fix plugin code: %w", err)
	}

	return &PluginGenerationResult{
		Code:           validatedCode,
		InputDesc:      inputDesc,
		OutputDesc:     outputDesc,
		Dependencies:   dependencies,
		RelatedPlugins: relatedPlugins,
		Names:          names,
	}, nil
}

// generateInputOutputDescByAI 通过AI生成输入输出描述和插件名称
func (s *AIService) generateInputOutputDescByAI(requirement string) (models.PluginInputDesc, models.PluginOutputDesc, PluginNames, error) {
	// 构造prompt（只基于需求，不包含插件信息）
	prompt := strings.Replace(s.ioDescPromptTemplate, "{{REQUIREMENT}}", requirement, 1)

	modelType := s.getConfigValue("OPENAI_MODEL_TYPE")
	requestBody := map[string]interface{}{
		"model": modelType,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "你是一个专业的插件接口设计师，能够根据用户需求分析并定义清晰的输入输出字段规范。",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.3, // 降低温度以获得更稳定的结构化输出
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return models.PluginInputDesc{}, models.PluginOutputDesc{}, PluginNames{}, err
	}

	req, err := http.NewRequest("POST", s.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return models.PluginInputDesc{}, models.PluginOutputDesc{}, PluginNames{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return models.PluginInputDesc{}, models.PluginOutputDesc{}, PluginNames{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.PluginInputDesc{}, models.PluginOutputDesc{}, PluginNames{}, err
	}

	if resp.StatusCode != 200 {
		return models.PluginInputDesc{}, models.PluginOutputDesc{}, PluginNames{}, fmt.Errorf("AI API error: %s", string(body))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return models.PluginInputDesc{}, models.PluginOutputDesc{}, PluginNames{}, err
	}

	if len(response.Choices) == 0 {
		return models.PluginInputDesc{}, models.PluginOutputDesc{}, PluginNames{}, fmt.Errorf("no response from AI")
	}

	content := response.Choices[0].Message.Content

	// 尝试解析JSON格式的输入输出描述和名称
	inputDesc, outputDesc, names, err := s.parseInputOutputDesc(content)
	if err != nil {
		// 如果解析失败，回退到基于关键词的生成和默认名称
		log.Printf("Failed to parse AI response, using fallback: %v", err)
		return models.PluginInputDesc{}, models.PluginOutputDesc{}, PluginNames{
			ChineseName: "未命名工具",
			EnglishName: "UnnamedTool",
		}, fmt.Errorf("failed to parse AI response as JSON, falling back to keyword-based generation: %v", err)
	}

	// 确保至少有一个输入和输出字段
	if len(inputDesc.TextList) == 0 && len(inputDesc.ImageList) == 0 && len(inputDesc.FileList) == 0 &&
		len(inputDesc.AudioList) == 0 && len(inputDesc.VideoList) == 0 && len(inputDesc.DocumentList) == 0 && len(inputDesc.OtherList) == 0 {
		inputDesc.TextList = []*models.InputItem{
			{
				Content:     "输入数据",
				ContentType: models.InputTypeText,
				Description: "输入数据",
				Options:     []models.SelectOption{},
			},
		}
	}

	if len(outputDesc.TextList) == 0 && len(outputDesc.ImageList) == 0 && len(outputDesc.FileList) == 0 &&
		len(outputDesc.AudioList) == 0 && len(outputDesc.VideoList) == 0 && len(outputDesc.DocumentList) == 0 && len(outputDesc.OtherList) == 0 {
		outputDesc.TextList = []*models.InputItem{
			{
				Content:     "输出结果",
				ContentType: models.InputTypeText,
				Description: "输出结果",
				Options:     []models.SelectOption{},
			},
		}
	}
	return inputDesc, outputDesc, names, nil
}

// generateCodeByAI 根据输入输出描述生成Go代码
func (s *AIService) generateCodeByAI(requirement string, inputDesc models.PluginInputDesc, outputDesc models.PluginOutputDesc, selectedPlugins []models.Plugin, genConfig *PluginGenerationConfig) (string, error) {
	// 构造输入输出规范的JSON字符串
	inputIO := models.PluginIO{}
	for _, item := range inputDesc.TextList {
		inputIO.TextList = append(inputIO.TextList, item.Content)
	}
	for _, item := range inputDesc.ImageList {
		inputIO.ImageList = append(inputIO.ImageList, item.Content)
	}
	for _, item := range inputDesc.FileList {
		inputIO.FileList = append(inputIO.FileList, item.Content)
	}
	for _, item := range inputDesc.AudioList {
		inputIO.AudioList = append(inputIO.AudioList, item.Content)
	}
	for _, item := range inputDesc.DocumentList {
		inputIO.DocumentList = append(inputIO.DocumentList, item.Content)
	}
	for _, item := range inputDesc.OtherList {
		inputIO.OtherList = append(inputIO.OtherList, item.Content)
	}
	outputIO := models.PluginIO{}
	for _, item := range outputDesc.TextList {
		outputIO.TextList = append(outputIO.TextList, item.Content)
	}
	for _, item := range outputDesc.ImageList {
		outputIO.ImageList = append(outputIO.ImageList, item.Content)
	}
	for _, item := range outputDesc.FileList {
		outputIO.FileList = append(outputIO.FileList, item.Content)
	}
	for _, item := range outputDesc.AudioList {
		outputIO.AudioList = append(outputIO.AudioList, item.Content)
	}
	for _, item := range outputDesc.DocumentList {
		outputIO.DocumentList = append(outputIO.DocumentList, item.Content)
	}
	for _, item := range outputDesc.OtherList {
		outputIO.OtherList = append(outputIO.OtherList, item.Content)
	}

	ioSpec := map[string]interface{}{
		"input":  inputIO,
		"output": outputIO,
	}

	ioSpecJson, err := json.MarshalIndent(ioSpec, "", "  ")
	if err != nil {
		return "", err
	}

	// 构造prompt
	prompt := strings.Replace(s.codePromptTemplate, "{{REQUIREMENT}}", requirement, 1)
	prompt = strings.Replace(prompt, "{{INPUT_OUTPUT_SPEC}}", string(ioSpecJson), 1)

	log.Printf("Prompt: %s", prompt)

	// 根据配置添加插件和AI能力说明
	if len(selectedPlugins) > 0 && genConfig.UseAvailablePlugins {
		pluginInfo := s.buildPluginInfoForCodePrompt(selectedPlugins)
		prompt += "\n\n" + pluginInfo
	}

	// 根据AI能力配置添加说明
	if genConfig.UseAICapabilities {
		prompt += "\n\n注意：可以使用AI能力。当用户需求涉及AI相关功能时，可以使用llm.InvokeLLM(prompt string) (string, error)函数来调用AI服务。（llm包已存在）"
	}

	modelType := s.getConfigValue("OPENAI_MODEL_TYPE")
	requestBody := map[string]interface{}{
		"model": modelType,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "你是一个专业的代码生成器，能够根据明确的输入输出规范生成高质量的Yaegi解释器可执行代码,所有代码都必须是逻辑，不能有模拟数据。代码不要有任何语法错误",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", s.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("AI API error: %s", string(body))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	content := response.Choices[0].Message.Content
	code := extractCode(content)

	return code, nil
}

// extractInputOutputDesc 从AI响应中提取输入输出描述
func (s *AIService) parseInputOutputDesc(content string) (models.PluginInputDesc, models.PluginOutputDesc, PluginNames, error) {
	var inputDesc models.PluginInputDesc
	var outputDesc models.PluginOutputDesc
	var names PluginNames

	// 尝试提取JSON格式的内容
	// 首先尝试找到JSON代码块
	jsonRe := regexp.MustCompile("```json\\s*\\n([\\s\\S]*?)\\n```")
	jsonMatches := jsonRe.FindStringSubmatch(content)
	if len(jsonMatches) > 1 {
		content = jsonMatches[1]
	}

	// 尝试直接解析为JSON
	var ioDesc PluginInputOutputResult

	err := json.Unmarshal([]byte(content), &ioDesc)
	if err != nil {
		return inputDesc, outputDesc, names, err
	}

	return ioDesc.InputDesc, ioDesc.OutputDesc, ioDesc.Names, nil
}

// extractDependenciesByAI 通过AI从代码中提取依赖包
func (s *AIService) extractDependenciesByAI(code string) ([]models.PluginDependency, error) {
	// 如果API key未设置，尝试从配置服务获取

	prompt := fmt.Sprintf(`请分析以下Go代码片段，提取其中的所有import语句，并返回JSON格式的依赖包列表。

代码内容：
%s

请注意：
1. 只提取实际导入的包名，不要包含默认导入的包名，例如"encoding/json"、"fmt"、"log"、"strings"、"strconv"、"errors"、"os"、"io"、"bytes"、"context"、"reflect"、"time"、"regexp"
2. 返回格式必须是JSON数组，每个元素包含"package"字段，格式如下：
[{"package": "github.com/example/package"}, {"package": "golang.org/x/example"}]
2. 例如，如果包里涉及zip.Decode()，则需要导入"archive/zip"包。
3. 排除"llm/llm","lojifile/lojifile","loji/loji"包，这些包已导入，无需重复导入，标准库的包也不需要
4. 如果代码的注释里写了需要导入的包，那么直接输出这些包，不要自己判断。
4. 如果没有外部依赖，返回空数组[]`, code)

	modelType := s.getConfigValue("OPENAI_MODEL_TYPE")
	requestBody := map[string]interface{}{
		"model": modelType,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "你是一个专业的Go代码分析工具，能够准确提取代码中的依赖包信息。只返回JSON格式的结果，不要包含其他说明文字。",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.1, // 降低温度以获得更稳定的结构化输出
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", s.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	choices, ok := response["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("invalid AI response format")
	}

	content, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid AI response content")
	}

	// 解析AI返回的JSON依赖列表
	var dependencies []models.PluginDependency
	if err := json.Unmarshal([]byte(content), &dependencies); err != nil {
		// 如果解析失败，尝试提取JSON部分
		re := regexp.MustCompile(`\[[\s\S]*\]`)
		jsonMatch := re.FindString(content)
		if jsonMatch != "" {
			if err := json.Unmarshal([]byte(jsonMatch), &dependencies); err != nil {
				log.Printf("Warning: failed to parse AI response as dependency list: %v, content: %s", err, content)
				// 如果AI解析失败，回退到原来的字符串匹配方法
				return s.extractDependenciesFromCode(code), nil
			}
		} else {
			log.Printf("Warning: no JSON found in AI response, content: %s", content)
			// 如果AI解析失败，回退到原来的字符串匹配方法
			return s.extractDependenciesFromCode(code), nil
		}
	}
	var newDependencies []models.PluginDependency
	for _, dependency := range dependencies {
		if strings.Contains(dependency.Package, "lojifile") || strings.Contains(dependency.Package, "llm") {
			continue
		}
		newDependencies = append(newDependencies, dependency)
	}

	return newDependencies, nil
}

// extractDependenciesFromCode 从代码中提取依赖包（备用方法）
func (s *AIService) extractDependenciesFromCode(code string) []models.PluginDependency {
	var dependencies []models.PluginDependency
	packageSet := make(map[string]bool)

	// 简单的字符串匹配查找import语句
	lines := strings.Split(code, "\n")
	inImportBlock := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 检查是否进入import块
		if strings.HasPrefix(line, "import") && strings.Contains(line, "(") {
			inImportBlock = true
			continue
		}

		// 检查是否退出import块
		if inImportBlock && strings.Contains(line, ")") {
			inImportBlock = false
			continue
		}

		// 在import块内处理每一行
		if inImportBlock && line != "" && !strings.HasPrefix(line, "//") && !strings.HasPrefix(line, "/*") {
			// 提取包名（去掉引号）
			if strings.Contains(line, `"`) {
				// 处理别名导入，如：pkg "github.com/user/package"
				if strings.Contains(line, " ") {
					parts := strings.Fields(line)
					if len(parts) >= 2 && strings.Contains(parts[1], `"`) {
						pkg := strings.Trim(parts[1], `"`)
						if pkg != "" {
							packageSet[pkg] = true
						}
					}
				} else {
					// 处理普通导入，如："github.com/user/package"
					parts := strings.Fields(line)
					if len(parts) > 0 {
						pkg := strings.Trim(parts[0], `"`)
						if pkg != "" {
							packageSet[pkg] = true
						}
					}
				}
			}
		}

		// 也匹配单行import
		if strings.HasPrefix(line, `import "`) && strings.HasSuffix(line, `"`) {
			pkg := strings.TrimPrefix(line, `import "`)
			pkg = strings.TrimSuffix(pkg, `"`)
			if pkg != "" {
				packageSet[pkg] = true
			}
		}
	}

	// 过滤掉系统包，并转换为依赖列表
	for pkg := range packageSet {
		// 过滤掉标准库中的基础包（这些会被yaegi自动加载）
		if pkg != "encoding/json" && pkg != "fmt" && pkg != "log" &&
			pkg != "strings" && pkg != "strconv" && pkg != "errors" &&
			pkg != "os" && pkg != "io" && pkg != "bytes" && pkg != "context" &&
			pkg != "reflect" && pkg != "time" && pkg != "regexp" && pkg != "path/filepath" {
			dependencies = append(dependencies, models.PluginDependency{
				Package: pkg,
				Version: "", // 可以后续扩展版本信息
			})
		}
	}

	return dependencies
}

// extractCode 从AI响应中提取代码
func extractCode(content string) string {
	// 提取代码块
	re := regexp.MustCompile("```(?:go|golang)?\\s*\\n([\\s\\S]*?)\\n```")
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return strings.TrimSpace(content)
}

// loadIoDescPromptTemplate 加载输入输出描述生成prompt模板
func loadIoDescPromptTemplate() string {
	// 不从文件加载
	// homeDir, err := os.UserHomeDir()
	// if err == nil {
	// 	promptPath := filepath.Join(homeDir, ".loji-app", "prompts", "plugin_generator.txt")
	// 	log.Printf("Prompt path: %s", promptPath)
	// 	if content, err := os.ReadFile(promptPath); err == nil {
	// 		log.Printf("Prompt template: %s", string(content))
	// 		return string(content)
	// 	}
	// }
	// log.Printf("Prompt template: %v", err)

	// 返回默认模板
	return `请根据以下需求生成Go WebAssembly插件的名称和输入输出字段描述：

需求：{{REQUIREMENT}}

请分析需求并提供JSON格式的完整描述，包括插件名称和输入输出字段。格式如下：
{
  "names": {
    "chineseName": "中文名称",
    "englishName": "EnglishFunctionName"
  },
  "input": {
    "textList": [
      {
        "content": "字段名称1",
        "contentType": "text|number|boolean|select|radio|checkbox",
        "description": "字段描述，用途说明",
        "options": [
          {"label": "显示文本1", "value": "实际值1"},
          {"label": "显示文本2", "value": "实际值2"}
        ]
      },
	  {
        "content": "字段名称2",
        "contentType": "text|number|boolean|select|radio|checkbox",
        "description": "字段描述，用途说明",
        "options": [
          {"label": "编码", "value": "encode"},
          {"label": "解码", "value": "decode"}
        ]
      }
    ],
    "imageList": [],
    "fileList": [],
    "audioList": [],
    "videoList": [],
    "documentList": [],
    "otherList": []
  },
  "output": {
    "textList": [
	  {
        "content": "字段名称1",
        "contentType": "text|number",
        "description": "字段描述，用途说明"
      },
	  {
        "content": "字段名称2",
        "contentType": "text|number",
        "description": "字段描述，用途说明"
      }
	],
    "imageList": [],
    "fileList": [],
    "audioList": [],
    "videoList": [],
    "documentList": [],
    "otherList": []
  }
}

要求：
0. 根据需求分析所需的输入和输出类型
1. names: 插件名称
   - chineseName: 描述性中文名称，用于界面显示，简洁明了
   - englishName: 开头大写的英文函数名，符合Go语言命名规范
2. input/output: 输入输出字段描述
3. 输入字段格式：
   - content: 字段名称，要清晰明了
   - contentType: 支持以下类型：
     * 文本类：
       - "text": 文本输入框
       - "number": 数字输入框
       - "boolean": 布尔开关（true/false）
       - "select": 单选下拉框（需要提供options数组）
       - "radio": 单选按钮组（需要提供options数组）
       - "checkbox": 多选框组（需要提供options数组）
     * 文件类：
       - "file": 通用文件（支持所有文件类型）
       - "image": 图片文件（如JPG、PNG、GIF）
       - "document": 文档文件（如PDF、Word、TXT）
       - "audio": 音频文件（如MP3、WAV）
       - "video": 视频文件（如MP4、AVI）
   - description: 字段的用途说明，告知用户这个字段的作用，简洁明了
   - options: 当contentType为select/radio/checkbox时必须提供选项数组
     * 格式: [{"label": "显示文本", "value": "实际值"}]
     * label: 前端展示给用户的文本（可以是中文）
     * value: 实际传递给插件的值（通常是英文或代码）
     * 示例: [{"label": "编码", "value": "encode"}, {"label": "解码", "value": "decode"}]

4. **重要**：字段分类规则
   - textList: 所有文本类型（text、number、boolean、select、radio、checkbox等）
   - imageList: 图片文件（contentType为"image"）
   - audioList: 音频文件（contentType为"audio"）
   - videoList: 视频文件（contentType为"video"）
   - fileList: 通用文件（contentType为"file"）
   - documentList: 文档文件（contentType为"document"）
   - otherList: 其他特殊类型

5. **文件类型使用场景**：
   - 如果需求中提到"上传文件"、"读取文件"、"文件处理"，使用fileList
   - 如果需求中提到"图片"、"照片"、"图像"，使用imageList
   - 如果需求中提到"文档"、"PDF"、"Word"，使用documentList
   - 如果需求中提到"音频"、"音乐"、"声音"，使用audioList
   - 如果需求中提到"视频"，使用videoList

6. **文件数据格式**：
   - 所有文件都会上传到临时目录，以文件路径形式传入
   - 格式示例: "/tmp/loji-uploads/1699876543210_example.txt"
   - 插件可以使用外部函数 ReadFile(fileData) 来读取文件内容
   - 插件可以使用 ReadFileAsString(fileData) 来读取文本文件内容为字符串
   - 支持任意大小的文件

7. 输出字段格式：与输入字段相同规则
8. 只返回JSON，不要有其他解释文字
9. 字段名称要清晰明了，能够准确描述用途
10. description要简洁地说明字段用途，帮助用户理解需要输入什么
11. 根据需求智能选择合适的输入类型：
   - 是/否、开启/关闭、启用/禁用等二选一场景使用boolean
   - 多个固定选项中选一个使用select或radio
   - 多个选项可以同时选择使用checkbox
   - 普通文本使用text，数值使用number
   - 文件上传使用file/image/document/audio/video
12. 重要：在插件执行时，所有非多媒体类型的输入值都会以字符串数组（[]string）的形式传入PluginInput.TextList，插件代码需要自行处理类型转换

请根据需求 "{{REQUIREMENT}}" 生成相应的名称和输入输出字段描述。
`
}

// loadCodePromptTemplate 加载代码生成prompt模板
func loadCodePromptTemplate() string {
	// 不用从文件加载
	// homeDir, err := os.UserHomeDir()
	// if err == nil {
	// 	promptPath := filepath.Join(homeDir, ".loji-app", "prompts", "plugin_code_generator.txt")
	// 	if content, err := os.ReadFile(promptPath); err == nil {
	// 		return string(content)
	// 	}
	// }

	// 返回默认模板（Yaegi版本）
	return `请根据以下需求和输入输出规范生成Go插件代码，该代码将在Yaegi解释器中直接执行：

需求：{{REQUIREMENT}}

输入输出规范：
{{INPUT_OUTPUT_SPEC}}

要求：
1. 生成适合Yaegi解释器执行的标准Go代码,不要自己写函数，只需要有逻辑代码，其他交给Yaegi解释器执行
2. 使用PluginInput结构体获取输入数据(pluginInput不是指针类型)
3. 使用PluginOutput结构体设置输出结果(pluginOutput不是指针类型)
4. 代码要简洁、高效、易于理解，不要有语法错误。
5. 包含适当的错误处理
6. 使用标准库，避免外部依赖
7. 输入输出都使用JSON格式，严格按照提供的输入输出规范定义数据结构。代码的输入类型要和定义的输入类型一致,代码的输出结果也要和定义的输出结果一致。
8. 只返回核心业务逻辑代码，不要包含package声明，但需要包含必要的import语句
9. 不要有main函数，由Yaegi执行框架负责
10. 如果报错，PluginOutput里需要带上输入值的JsonString，方便用户排查问题
11. 入参会按照输入输出规范定义的结构体进行类型断言，参数与入参描述顺序一致。
12. 生成代码里不需要有import字段，但需要在注释里说明需要导入的包，比如：json.Marshal() // import "encoding/json"
13. 重要：无论输入描述中的contentType是什么（boolean/select/radio/checkbox/number/text），所有输入字段的值都以[]string类型传入。如果需要其他类型，必须在代码中进行类型转换，例如：
    - boolean类型：需要将字符串"true"/"false"转换为bool
    - number类型：需要使用strconv.Atoi()或strconv.ParseFloat()转换
    - select/radio/checkbox：直接使用字符串值即可
14. 生成的代码是一个函数内的代码，因此一定不要自己声明函数，要么使用已有固定函数，要么直接在代码里实现。换句话说，代码里不能有func这个关键字。
15. 如果需要组装换行的的字符串，尽量在一行代码内实现，golang在双引号号里直接换行会报错。
16. **文件处理**：系统提供文件处理外部函数，注意，不支持PDF文件的读取（在lojifile/lojifile包中）：(lojifile/lojifile包已导入，无需重复导入)
    - lojifile.ReadFile(filePath string) ([]byte, error) - 读取文件内容为字节数组
    - lojifile.ReadFileAsString(filePath string) (string, error) - 读取纯文本文件内容为字符串，目前并不支持PDF文件的读取
    - lojifile.GetFileExtension(filePath string) string - 获取文件扩展名
    - lojifile.GetFileSize(filePath string) (int, error) - 获取文件大小
    文件路径从PluginInput的FileList/ImageList/DocumentList/AudioList/VideoList中获取
    格式示例: "/tmp/loji-uploads/1699876543210_example.txt"

代码结构参考：
` + "```go" + `
// 主执行逻辑
// pluginInput 是一个 PluginInput 结构体，可以直接访问字段
// pluginOutput 是一个 PluginOutput 结构体，直接设置字段即可

PluginInput定义为：
type PluginInput struct {
	TextList     []string 
	ImageList    []string
	FileList     []string
	AudioList    []string
	VideoList    []string
	DocumentList []string 
	OtherList    []string
}

PluginOutput定义为：
type PluginOutput struct {
	TextList     []string
	ImageList    []string
	FileList     []string
	AudioList    []string
	VideoList    []string
	DocumentList []string
	OtherList    []string
	PluginError  string
}

// 获取输入数据，例如：
if len(pluginInput.TextList) > 0 {
    text := pluginInput.TextList[0]
    // 处理业务逻辑
}

// 设置输出结果，例如：
pluginOutput.TextList = []string{"处理结果"}
pluginOutput.PluginError = "" // 如果没有错误，设置为空字符串
` + "```" + `

请根据需求和输入输出规范生成相应的Go代码。请注意：
- pluginInput 和 pluginOutput 是具体的结构体类型，可以直接访问字段
- 直接设置pluginOutput的字段，不要返回JSON字符串
- 如果处理出错，将错误信息设置到pluginOutput.PluginError字段
- 使用标准Go语法，Yaegi会处理类型转换`
}

// selectRelevantPlugins AI筛选相关插件
func (s *AIService) selectRelevantPlugins(requirement string, availablePlugins []models.Plugin) ([]models.Plugin, error) {
	if len(availablePlugins) == 0 {
		log.Printf("No available plugins provided, skipping plugin selection")
		return []models.Plugin{}, nil
	}

	// 构建插件筛选prompt
	prompt := s.buildPluginSelectionPrompt(requirement, availablePlugins)

	// 调用AI进行筛选
	modelType := s.getConfigValue("OPENAI_MODEL_TYPE")
	requestBody := map[string]interface{}{
		"model": modelType,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "你是一个专业的插件分析助手，能够根据用户需求从可用插件中筛选出需要的插件。",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	// 发送请求
	resp, err := s.makeAPIRequest(jsonData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	// 解析AI响应
	var openaiResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &openaiResponse); err != nil {
		return nil, err
	}

	if len(openaiResponse.Choices) == 0 {
		return []models.Plugin{}, nil
	}

	// 解析AI返回的插件名称列表
	selectedNames := s.parseSelectedPluginNames(openaiResponse.Choices[0].Message.Content)

	// 根据名称匹配插件
	var selectedPlugins []models.Plugin
	for _, name := range selectedNames {
		for _, plugin := range availablePlugins {
			if plugin.Name == name {
				selectedPlugins = append(selectedPlugins, plugin)
				break
			}
		}
	}

	return selectedPlugins, nil
}

// buildPluginSelectionPrompt 构建插件筛选prompt
func (s *AIService) buildPluginSelectionPrompt(requirement string, availablePlugins []models.Plugin) string {
	var pluginList strings.Builder
	pluginList.WriteString("可用插件列表：\n")

	for i, plugin := range availablePlugins {
		pluginList.WriteString(fmt.Sprintf("%d. %s - %s\n", i+1, plugin.Name, plugin.Description))
	}

	return fmt.Sprintf(`请分析用户需求，从以下可用插件中筛选出需要的插件：

用户需求：%s

%s

请返回一个JSON数组，包含需要用到的插件名称，例如：["PluginName1", "PluginName2"]

如果不需要任何插件，返回空数组：[]`, requirement, pluginList.String())
}

// parseSelectedPluginNames 解析AI返回的插件名称列表
func (s *AIService) parseSelectedPluginNames(content string) []string {
	var names []string

	// 尝试解析JSON数组
	if err := json.Unmarshal([]byte(content), &names); err != nil {
		log.Printf("Failed to parse plugin names as JSON: %v, content: %s", err, content)
		// 如果JSON解析失败，尝试简单文本解析
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.Contains(line, "[") && !strings.Contains(line, "]") {
				// 移除引号和逗号
				line = strings.Trim(line, `",`)
				if line != "" {
					names = append(names, line)
				}
			}
		}
	}

	return names
}

// buildPluginInfoForPrompt 为prompt构建插件信息
func (s *AIService) buildPluginInfoForPrompt(selectedPlugins []models.Plugin) string {
	var info strings.Builder

	for i, plugin := range selectedPlugins {
		info.WriteString(fmt.Sprintf("%d. %s - %s\n", i+1, plugin.Name, plugin.Description))
		info.WriteString(fmt.Sprintf("   调用方式：loji.%s(input)\n", plugin.Name))
		info.WriteString("   输入：PluginInput结构体，输出：PluginOutput结构体\n\n")
	}

	return info.String()
}

// buildPluginInfoForCodePrompt 为代码生成prompt构建插件信息
func (s *AIService) buildPluginInfoForCodePrompt(selectedPlugins []models.Plugin) string {
	var info strings.Builder
	info.WriteString("可用的插件函数（已注册在loji包中）：\n\n")

	for i, plugin := range selectedPlugins {
		info.WriteString(fmt.Sprintf("%d. 插件：%s\n", i+1, plugin.Name))
		info.WriteString(fmt.Sprintf("   描述：%s\n", plugin.Description))
		info.WriteString(fmt.Sprintf("   调用：result := loji.%s(input)\n", plugin.Name))

		// 详细说明输入参数结构
		info.WriteString("   输入参数（PluginInput结构体）：\n")
		s.buildPluginInputDetails(&info, plugin)

		info.WriteString("   返回值（PluginOutput结构体）：\n")
		s.buildPluginOutputDetails(&info, plugin)

		info.WriteString("   错误处理：检查result.PluginError字段\n\n")
	}

	info.WriteString("PluginInput字段组装说明,PluginInput在当前包内，前面不需要包名，直接使用：\n")
	info.WriteString("- textList: []string 文本字段列表，按顺序对应插件的文本输入要求\n")
	info.WriteString("- imageList: []string 图片URL列表，按顺序对应插件的图片输入要求\n")
	info.WriteString("- fileList: []string 文件路径列表，按顺序对应插件的文件输入要求\n")
	info.WriteString("- audioList: []string 音频URL列表，按顺序对应插件的音频输入要求\n")
	info.WriteString("- videoList: []string 视频URL列表，按顺序对应插件的视频输入要求\n")
	info.WriteString("- documentList: []string 文档URL列表，按顺序对应插件的文档输入要求\n")
	info.WriteString("- otherList: []string 其他类型数据列表，按顺序对应插件的其他输入要求\n\n")

	info.WriteString("组装示例：\n")
	info.WriteString("input := PluginInput{\n")
	info.WriteString("    textList: []string{\"第一个文本参数\", \"第二个文本参数\"},\n")
	info.WriteString("    imageList: []string{\"https://example.com/image1.jpg\"},\n")
	info.WriteString("    fileList: []string{},\n")
	info.WriteString("    audioList: []string{},\n")
	info.WriteString("    videoList: []string{},\n")
	info.WriteString("    documentList: []string{},\n")
	info.WriteString("    otherList: []string{},\n")
	info.WriteString("}\n\n")

	info.WriteString("结果处理示例：\n\n")
	info.WriteString("output := PluginOutput{\n")
	info.WriteString("    textList: []string{\"处理结果\"},\n")
	info.WriteString("    imageList: []string{},\n")
	info.WriteString("    fileList: []string{},\n")
	info.WriteString("    audioList: []string{},\n")
	info.WriteString("    videoList: []string{},\n")
	info.WriteString("    documentList: []string{},\n")
	info.WriteString("    otherList: []string{},\n")
	info.WriteString("    pluginError: \"\",\n")
	info.WriteString("}\n\n")

	return info.String()
}

// buildPluginInputDetails 构建插件输入参数的详细信息
func (s *AIService) buildPluginInputDetails(info *strings.Builder, plugin models.Plugin) {
	if plugin.Input == nil {
		info.WriteString("      无输入参数要求\n")
		return
	}

	// 文本输入
	if len(plugin.Input.TextList) > 0 {
		info.WriteString("      textList (文本输入):\n")
		for j, field := range plugin.Input.TextList {
			fieldName := field.Content
			if fieldName == "" {
				fieldName = fmt.Sprintf("字段%d", j+1)
			}
			info.WriteString(fmt.Sprintf("        %d. %s (%s) - %s\n", j+1, fieldName, s.getContentTypeName(string(field.ContentType)), field.Description))
		}
	}

	// 图片输入
	if len(plugin.Input.ImageList) > 0 {
		info.WriteString("      imageList (图片输入):\n")
		for j, field := range plugin.Input.ImageList {
			fieldName := field.Content
			if fieldName == "" {
				fieldName = fmt.Sprintf("图片字段%d", j+1)
			}
			info.WriteString(fmt.Sprintf("        %d. %s - %s\n", j+1, fieldName, field.Description))
		}
	}

	// 文件输入
	if len(plugin.Input.FileList) > 0 {
		info.WriteString("      fileList (文件输入):\n")
		for j, field := range plugin.Input.FileList {
			fieldName := field.Content
			if fieldName == "" {
				fieldName = fmt.Sprintf("文件字段%d", j+1)
			}
			info.WriteString(fmt.Sprintf("        %d. %s - %s\n", j+1, fieldName, field.Description))
		}
	}

	// 音频输入
	if len(plugin.Input.AudioList) > 0 {
		info.WriteString("      audioList (音频输入):\n")
		for j, field := range plugin.Input.AudioList {
			fieldName := field.Content
			if fieldName == "" {
				fieldName = fmt.Sprintf("音频字段%d", j+1)
			}
			info.WriteString(fmt.Sprintf("        %d. %s - %s\n", j+1, fieldName, field.Description))
		}
	}

	// 视频输入
	if len(plugin.Input.VideoList) > 0 {
		info.WriteString("      videoList (视频输入):\n")
		for j, field := range plugin.Input.VideoList {
			fieldName := field.Content
			if fieldName == "" {
				fieldName = fmt.Sprintf("视频字段%d", j+1)
			}
			info.WriteString(fmt.Sprintf("        %d. %s - %s\n", j+1, fieldName, field.Description))
		}
	}

	// 文档输入
	if len(plugin.Input.DocumentList) > 0 {
		info.WriteString("      documentList (文档输入):\n")
		for j, field := range plugin.Input.DocumentList {
			fieldName := field.Content
			if fieldName == "" {
				fieldName = fmt.Sprintf("文档字段%d", j+1)
			}
			info.WriteString(fmt.Sprintf("        %d. %s - %s\n", j+1, fieldName, field.Description))
		}
	}

	// 其他输入
	if len(plugin.Input.OtherList) > 0 {
		info.WriteString("      otherList (其他输入):\n")
		for j, field := range plugin.Input.OtherList {
			fieldName := field.Content
			if fieldName == "" {
				fieldName = fmt.Sprintf("其他字段%d", j+1)
			}
			info.WriteString(fmt.Sprintf("        %d. %s - %s\n", j+1, fieldName, field.Description))
		}
	}
}

// buildPluginOutputDetails 构建插件输出参数的详细信息
func (s *AIService) buildPluginOutputDetails(info *strings.Builder, plugin models.Plugin) {
	if plugin.Output == nil {
		info.WriteString("      无输出参数说明\n")
		return
	}

	// 文本输出
	if len(plugin.Output.TextList) > 0 {
		info.WriteString("      textList (文本输出):\n")
		for j, field := range plugin.Output.TextList {
			fieldName := field.Content
			if fieldName == "" {
				fieldName = fmt.Sprintf("输出字段%d", j+1)
			}
			info.WriteString(fmt.Sprintf("        %d. %s (%s) - %s\n", j+1, fieldName, s.getContentTypeName(string(field.ContentType)), field.Description))
		}
	}

	// 图片输出
	if len(plugin.Output.ImageList) > 0 {
		info.WriteString("      imageList (图片输出):\n")
		for j, field := range plugin.Output.ImageList {
			fieldName := field.Content
			if fieldName == "" {
				fieldName = fmt.Sprintf("图片输出%d", j+1)
			}
			info.WriteString(fmt.Sprintf("        %d. %s - %s\n", j+1, fieldName, field.Description))
		}
	}

	// 文件输出
	if len(plugin.Output.FileList) > 0 {
		info.WriteString("      fileList (文件输出):\n")
		for j, field := range plugin.Output.FileList {
			fieldName := field.Content
			if fieldName == "" {
				fieldName = fmt.Sprintf("文件输出%d", j+1)
			}
			info.WriteString(fmt.Sprintf("        %d. %s - %s\n", j+1, fieldName, field.Description))
		}
	}

	// 音频输出
	if len(plugin.Output.AudioList) > 0 {
		info.WriteString("      audioList (音频输出):\n")
		for j, field := range plugin.Output.AudioList {
			fieldName := field.Content
			if fieldName == "" {
				fieldName = fmt.Sprintf("音频输出%d", j+1)
			}
			info.WriteString(fmt.Sprintf("        %d. %s - %s\n", j+1, fieldName, field.Description))
		}
	}

	// 视频输出
	if len(plugin.Output.VideoList) > 0 {
		info.WriteString("      videoList (视频输出):\n")
		for j, field := range plugin.Output.VideoList {
			fieldName := field.Content
			if fieldName == "" {
				fieldName = fmt.Sprintf("视频输出%d", j+1)
			}
			info.WriteString(fmt.Sprintf("        %d. %s - %s\n", j+1, fieldName, field.Description))
		}
	}

	// 文档输出
	if len(plugin.Output.DocumentList) > 0 {
		info.WriteString("      documentList (文档输出):\n")
		for j, field := range plugin.Output.DocumentList {
			fieldName := field.Content
			if fieldName == "" {
				fieldName = fmt.Sprintf("文档输出%d", j+1)
			}
			info.WriteString(fmt.Sprintf("        %d. %s - %s\n", j+1, fieldName, field.Description))
		}
	}

	// 其他输出
	if len(plugin.Output.OtherList) > 0 {
		info.WriteString("      otherList (其他输出):\n")
		for j, field := range plugin.Output.OtherList {
			fieldName := field.Content
			if fieldName == "" {
				fieldName = fmt.Sprintf("其他输出%d", j+1)
			}
			info.WriteString(fmt.Sprintf("        %d. %s - %s\n", j+1, fieldName, field.Description))
		}
	}

	if plugin.Output.PluginError != "" {
		info.WriteString(fmt.Sprintf("      pluginError: %s\n", plugin.Output.PluginError))
	}
}

// getContentTypeName 获取内容类型的中文名称
func (s *AIService) getContentTypeName(contentType string) string {
	switch contentType {
	case "text":
		return "文本"
	case "number":
		return "数字"
	case "email":
		return "邮箱"
	case "url":
		return "链接"
	case "phone":
		return "电话"
	case "dateTime":
		return "日期时间"
	case "boolean":
		return "布尔值"
	case "select":
		return "选择项"
	case "radio":
		return "单选"
	case "checkbox":
		return "多选"
	case "textarea":
		return "多行文本"
	default:
		return "未知类型"
	}
}

// makeAPIRequest 发送HTTP请求到OpenAI API
func (s *AIService) makeAPIRequest(jsonData []byte) (*http.Response, error) {
	apiURL := s.getConfigValue("OPENAI_API_URL")
	if apiURL == "" {
		apiURL = "https://api.openai.com/v1/chat/completions"
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	client := &http.Client{}
	return client.Do(req)
}

// validateAndFixCode 验证和修复生成的代码
func (s *AIService) validateAndFixCode(code string, inputDesc models.PluginInputDesc, outputDesc models.PluginOutputDesc, selectedPlugins []models.Plugin, genConfig *PluginGenerationConfig, dependencies []models.PluginDependency) (string, error) {
	const maxRetries = 3
	currentCode := code

	for attempt := 0; attempt < maxRetries; attempt++ {
		log.Printf("代码语法检查尝试 #%d", attempt+1)

		// 检查是否有明显的语法错误
		hasErrors, syntaxErrors := s.checkBasicSyntax(currentCode)
		if !hasErrors {
			log.Printf("代码语法检查通过")
			return currentCode, nil
		}

		log.Printf("发现语法错误: %v", syntaxErrors)

		// 如果是最后一次尝试，返回修复后的代码（可能仍有错误）
		if attempt == maxRetries-1 {
			log.Printf("已达到最大重试次数，返回最后修复的代码")
			return currentCode, nil // 返回最后一次修复的代码，不报错
		}

		// 使用AI修复代码
		log.Printf("正在使用AI修复代码...")
		fixedCode, err := s.fixCodeWithAI(currentCode, syntaxErrors, inputDesc, outputDesc, selectedPlugins, genConfig)
		if err != nil {
			log.Printf("AI修复代码失败: %v", err)
			continue // 继续下一次尝试
		}

		currentCode = fixedCode
		log.Printf("AI修复完成，开始下一次检查")
	}

	return currentCode, nil
}

// checkBasicSyntax 检查基本的语法错误
func (s *AIService) checkBasicSyntax(code string) (bool, []string) {
	var errors []string

	// 检查基本的语法问题
	lines := strings.Split(code, "\n")

	// 检查括号匹配
	openBraces := 0
	openParens := 0
	openBrackets := 0

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// 跳过空行和注释
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			continue
		}

		// 检查括号匹配
		for _, char := range line {
			switch char {
			case '{':
				openBraces++
			case '}':
				openBraces--
				if openBraces < 0 {
					errors = append(errors, fmt.Sprintf("第%d行: 多余的右大括号", lineNum))
				}
			case '(':
				openParens++
			case ')':
				openParens--
				if openParens < 0 {
					errors = append(errors, fmt.Sprintf("第%d行: 多余的右圆括号", lineNum))
				}
			case '[':
				openBrackets++
			case ']':
				openBrackets--
				if openBrackets < 0 {
					errors = append(errors, fmt.Sprintf("第%d行: 多余的右方括号", lineNum))
				}
			}
		}

		// 检查控制语句的语法
		if strings.Contains(line, "if ") && strings.Contains(line, ")") && !strings.Contains(line, "{") && !strings.HasSuffix(trimmed, "{") {
			if !strings.Contains(line, ") {") {
				errors = append(errors, fmt.Sprintf("第%d行: if语句缺少左大括号", lineNum))
			}
		}

		if strings.Contains(line, "for ") && strings.Contains(line, ")") && !strings.Contains(line, "{") && !strings.HasSuffix(trimmed, "{") {
			if !strings.Contains(line, ") {") {
				errors = append(errors, fmt.Sprintf("第%d行: for语句缺少左大括号", lineNum))
			}
		}

		// 检查字符串字面量是否完整
		quoteCount := strings.Count(line, `"`) - strings.Count(line, `\"`)
		if quoteCount%2 != 0 {
			errors = append(errors, fmt.Sprintf("第%d行: 字符串字面量不完整", lineNum))
		}
	}

	// 检查整体括号匹配
	if openBraces > 0 {
		errors = append(errors, fmt.Sprintf("缺少 %d 个右大括号", openBraces))
	}
	if openParens > 0 {
		errors = append(errors, fmt.Sprintf("缺少 %d 个右圆括号", openParens))
	}
	if openBrackets > 0 {
		errors = append(errors, fmt.Sprintf("缺少 %d 个右方括号", openBrackets))
	}

	// 检查是否有明显的语法错误模式
	if strings.Contains(code, "func(") && !strings.Contains(code, ") {") && !strings.Contains(code, "){") {
		errors = append(errors, "函数定义语法错误")
	}

	// 检查变量声明是否完整
	if strings.Contains(code, "var ") && !strings.Contains(code, "=") && !strings.Contains(code, ";") && !strings.Contains(code, "{") {
		// 允许没有初始化的变量声明
	}

	return len(errors) > 0, errors
}

// fixCodeWithAI 使用AI修复代码
func (s *AIService) fixCodeWithAI(code string, syntaxErrors []string, inputDesc models.PluginInputDesc, outputDesc models.PluginOutputDesc, selectedPlugins []models.Plugin, genConfig *PluginGenerationConfig) (string, error) {
	// 构造输入输出规范的JSON字符串
	inputIO := models.PluginIO{}
	for _, item := range inputDesc.TextList {
		inputIO.TextList = append(inputIO.TextList, item.Content)
	}
	for _, item := range inputDesc.ImageList {
		inputIO.ImageList = append(inputIO.ImageList, item.Content)
	}
	for _, item := range inputDesc.FileList {
		inputIO.FileList = append(inputIO.FileList, item.Content)
	}
	for _, item := range inputDesc.AudioList {
		inputIO.AudioList = append(inputIO.AudioList, item.Content)
	}
	for _, item := range inputDesc.DocumentList {
		inputIO.DocumentList = append(inputIO.DocumentList, item.Content)
	}
	for _, item := range inputDesc.OtherList {
		inputIO.OtherList = append(inputIO.OtherList, item.Content)
	}
	outputIO := models.PluginIO{}
	for _, item := range outputDesc.TextList {
		outputIO.TextList = append(outputIO.TextList, item.Content)
	}
	for _, item := range outputDesc.ImageList {
		outputIO.ImageList = append(outputIO.ImageList, item.Content)
	}
	for _, item := range outputDesc.FileList {
		outputIO.FileList = append(outputIO.FileList, item.Content)
	}
	for _, item := range outputDesc.AudioList {
		outputIO.AudioList = append(outputIO.AudioList, item.Content)
	}
	for _, item := range outputDesc.DocumentList {
		outputIO.DocumentList = append(outputIO.DocumentList, item.Content)
	}
	for _, item := range outputDesc.OtherList {
		outputIO.OtherList = append(outputIO.OtherList, item.Content)
	}

	ioSpec := map[string]interface{}{
		"input":  inputIO,
		"output": outputIO,
	}

	ioSpecJson, err := json.MarshalIndent(ioSpec, "", "  ")
	if err != nil {
		return "", err
	}

	// 构造修复prompt
	errorMsg := strings.Join(syntaxErrors, "; ")
	prompt := fmt.Sprintf(`请修复以下Go代码中的语法错误。

原始代码：
%s

发现的语法错误：
%s

输入输出规范：
%s

修复要求：
1. 修复所有列出的语法错误
2. 保持代码的逻辑功能不变
3. 确保代码结构正确
4. 只返回修复后的代码，不要包含任何解释
5. 代码应该适合Yaegi解释器执行，包含必要的import语句，但不要有package声明和main函数

请直接返回修复后的完整代码：`, code, errorMsg, string(ioSpecJson))

	modelType := s.getConfigValue("OPENAI_MODEL_TYPE")
	requestBody := map[string]interface{}{
		"model": modelType,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "你是一个专业的Go代码修复专家，擅长修复Yaegi解释器中的代码语法错误。",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.3, // 降低温度以获得更稳定的修复结果
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", s.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("AI API error: %s", string(body))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	content := response.Choices[0].Message.Content
	log.Printf("AI修复响应: %s", content)

	// 提取修复后的代码
	fixedCode := extractCode(content)

	return fixedCode, nil
}
