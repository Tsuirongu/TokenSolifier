package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"

	"loji-app/backend/models"
	"loji-app/backend/services/external_functions"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// pluginExecutor 独立的插件执行器，确保数据隔离
type pluginExecutor struct {
	pluginService *PluginService
	configService *ConfigService
	logFunc       func(string)
}

// YaegiService Yaegi执行服务
type YaegiService struct {
	interpreter   *interp.Interpreter
	ctx           context.Context
	logFunc       func(string)   // 日志回调函数
	pluginService *PluginService // 插件服务依赖
	configService *ConfigService // 配置服务依赖
}

// NewYaegiService 创建Yaegi服务实例
func NewYaegiService(pluginService *PluginService) *YaegiService {
	return &YaegiService{
		pluginService: pluginService,
		configService: NewConfigService(),
	}
}

// SetLogFunc 设置日志回调函数
func (s *YaegiService) SetLogFunc(logFunc func(string)) {
	s.logFunc = logFunc
}

// getConfigValue 获取配置值，如果配置为空则返回环境变量作为后备
func (s *YaegiService) getConfigValue(key string) string {
	if configItem, err := s.configService.GetConfig(key); err == nil && configItem.Value != "" {
		return configItem.Value
	}
	return os.Getenv(key)
}

// getConfigValueForExecutor 获取配置值（为executor使用）
func (e *pluginExecutor) getConfigValue(key string) string {
	if configItem, err := e.configService.GetConfig(key); err == nil && configItem.Value != "" {
		return configItem.Value
	}
	return os.Getenv(key)
}

// Initialize 初始化Yaegi解释器
func (s *YaegiService) Initialize(ctx context.Context) error {
	s.ctx = ctx

	// 初始化时不注册插件，只注册LLM函数
	if err := s.createNewInterpreter([]models.PluginRelatedPlugin{}); err != nil {
		return fmt.Errorf("failed to register InvokeLLM: %w", err)
	}

	log.Println("Yaegi interpreter initialized successfully")
	return nil
}

// ExecutePlugin 执行插件代码
func (s *YaegiService) ExecutePlugin(code string, input string, dependencies []string, relatedPlugins []models.PluginRelatedPlugin) (string, error) {
	// 为每次执行重新初始化解释器，避免重复声明错误
	if err := s.reinitializeInterpreter(relatedPlugins); err != nil {
		return "", fmt.Errorf("failed to reinitialize interpreter: %w", err)
	}

	// 为这次执行创建一个新的代码片段（包含日志函数定义和依赖）
	executionCode := s.wrapPluginCode(code, input, dependencies, relatedPlugins)

	// 执行代码
	_, err := s.interpreter.Eval(executionCode)
	if err != nil {
		return "", fmt.Errorf("failed to execute plugin code: %w", err)
	}

	// 获取PluginOutput的结果
	pluginOutput := s.getPluginOutput()

	// 转换为JSON字符串
	jsonResult, err := json.Marshal(pluginOutput)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonResult), nil
}

// GetWrappedPluginCode 获取包装后的插件执行代码（不执行）
func (s *YaegiService) GetWrappedPluginCode(originalCode string, dependencies []string, relatedPlugins []models.PluginRelatedPlugin) string {
	// 使用空的input来组装代码
	return s.wrapPluginCode(originalCode, "", dependencies, relatedPlugins)
}

func getFinalImports(imports []string) []string {

	// 我期望去除imporst里重复的包，并返回一个去重后的包列表
	uniqueImports := make(map[string]bool)
	for _, importStr := range imports {
		uniqueImports[importStr] = true
	}
	for _, importStr := range models.GetDefaultImports() {
		uniqueImports[importStr] = true
	}
	uniqueImportsList := make([]string, 0, len(uniqueImports))
	for importStr := range uniqueImports {
		uniqueImportsList = append(uniqueImportsList, importStr)
	}
	return uniqueImportsList
}

// wrapPluginCode 包装插件代码，使其能够在Yaegi中执行
func (s *YaegiService) wrapPluginCode(originalCode string, input string, dependencies []string, relatedPlugins []models.PluginRelatedPlugin) string {
	// 从用户代码中提取逻辑代码（不提取import，因为依赖由外部提供）
	_, logicCode := s.separateImportsAndLogic(originalCode)

	// 调试输出
	log.Printf("外部依赖: %v", dependencies)
	// log.Printf("逻辑代码: %s", logicCode)

	// 构建完整的import语句
	allImports := models.GetDefaultImports()
	allImports = append(allImports, dependencies...)

	// 动态检测：如果代码中包含loji引用，添加loji/loji包
	if s.containsLojiReference(logicCode) {
		allImports = append(allImports, `"loji/loji"`)
	}

	importBlock := strings.Join(getFinalImports(allImports), "\n\t")

	return fmt.Sprintf(`
package main

import (
	%s
)

// PluginInput 插件输入结构体
type PluginInput struct {
	TextList     []string `+"`json:\"textList\"`"+`
	ImageList    []string `+"`json:\"imageList\"`"+`
	FileList     []string `+"`json:\"fileList\"`"+`
	AudioList    []string `+"`json:\"audioList\"`"+`
	VideoList    []string `+"`json:\"videoList\"`"+`
	DocumentList []string `+"`json:\"documentList\"`"+`
	OtherList    []string `+"`json:\"otherList\"`"+`
}

// PluginOutput 插件输出结构体
type PluginOutput struct {
	TextList     []string `+"`json:\"textList\"`"+`
	ImageList    []string `+"`json:\"imageList\"`"+`
	FileList     []string `+"`json:\"fileList\"`"+`
	AudioList    []string `+"`json:\"audioList\"`"+`
	VideoList    []string `+"`json:\"videoList\"`"+`
	DocumentList []string `+"`json:\"documentList\"`"+`
	OtherList    []string `+"`json:\"otherList\"`"+`
	PluginError  string   `+"`json:\"pluginError\"`"+`
}


// PluginInputJSON 插件输入数据的JSON字符串
const PluginInputJSON = %s

// 全局变量声明（使用小写名称避免与类型名冲突）
var (
	pluginInput  PluginInput
	pluginOutput PluginOutput
)

// lojiLog 自定义日志函数
func lojiLog(message string) {
	%s
}
	
type OpenapiResponse struct {
	Choices []struct {
		Message struct {
			Content string `+"`json:\"content,omitempty\"`"+`
		} `+"`json:\"message\"`"+`
	} `+"`json:\"choices\"`"+`
}

// lojiLogf 带格式化的自定义日志函数
func lojiLogf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	%s
}

// initPlugin 初始化插件数据
func initPlugin() {
	if err := json.Unmarshal([]byte(PluginInputJSON), &pluginInput); err != nil {
		pluginInput = PluginInput{} // 初始化为空结构体
		pluginOutput = PluginOutput{PluginError: "输入JSON解析失败: " + err.Error()}
	}
}

// executePlugin 执行插件逻辑
func executePlugin() {
	%s
}

// 主函数（Yaegi需要）
func main() {
	initPlugin()
	executePlugin()
}
`, importBlock, s.formatInputJSON(input), s.getLogCall(), s.getLogfCall(), strings.TrimSpace(logicCode))
}

// formatInputJSON 格式化输入JSON
func (s *YaegiService) formatInputJSON(input string) string {
	// 由于使用反引号字符串，只需要处理反引号
	// 反斜杠不需要转义，因为JSON中的转义序列应该保持原样
	escaped := strings.ReplaceAll(input, "`", "` + \"`\" + `")

	return "`" + escaped + "`"
}

// getLogCall 生成日志函数调用代码
func (s *YaegiService) getLogCall() string {
	if s.logFunc != nil {
		return `log.Printf("[PLUGIN LOG] %s", message)`
	}
	return `// 日志功能未启用`
}

// getLogfCall 生成格式化日志函数调用代码
func (s *YaegiService) getLogfCall() string {
	if s.logFunc != nil {
		return `log.Printf("[PLUGIN LOG] %s", message)`
	}
	return `// 日志功能未启用`
}

// getPluginOutput 从Yaegi解释器中获取PluginOutput变量的值
func (s *YaegiService) getPluginOutput() interface{} {
	// 使用反射获取pluginOutput变量
	globals := s.interpreter.Globals()
	if pluginOutput, exists := globals["pluginOutput"]; exists {
		// 获取变量的当前值
		if pluginOutput.CanInterface() {
			return pluginOutput.Interface()
		}
	}

	// 返回nil，表示获取失败
	return nil
}

// extractResult 从反射值中提取结果
func (s *YaegiService) extractResult(result reflect.Value) string {
	if !result.IsValid() {
		return `{"error": "无效的执行结果"}`
	}

	// 如果是字符串，直接返回
	if result.Kind() == reflect.String {
		return result.String()
	}

	// 如果是interface{}，尝试转换为字符串
	if result.Kind() == reflect.Interface {
		if result.IsNil() {
			return `{"error": "空的执行结果"}`
		}

		// 尝试获取底层值
		underlying := result.Elem()
		if underlying.Kind() == reflect.String {
			return underlying.String()
		}

		// 尝试JSON序列化
		jsonBytes, err := json.Marshal(underlying.Interface())
		if err != nil {
			return fmt.Sprintf(`{"error": "结果序列化失败: %s"}`, err.Error())
		}

		return string(jsonBytes)
	}

	// 默认情况：JSON序列化
	jsonBytes, err := json.Marshal(result.Interface())
	if err != nil {
		return fmt.Sprintf(`{"error": "结果序列化失败: %s"}`, err.Error())
	}

	return string(jsonBytes)
}

// registerInvokeLLM 注册InvokeLLM外部函数到Yaegi解释器
func (s *YaegiService) createNewInterpreter(relatedPlugins []models.PluginRelatedPlugin) error {
	// 创建InvokeLLM函数的包装器，保持与原函数相同的签名和逻辑
	wrapper := func(prompt string) (string, error) {
		apiKey := s.getConfigValue("OPENAI_API_KEY")
		apiURL := s.getConfigValue("OPENAI_API_URL")
		modelType := s.getConfigValue("OPENAI_MODEL_TYPE")
		requestBody := map[string]interface{}{
			"model": modelType,
			"messages": []map[string]string{
				{
					"role":    "system",
					"content": "你是一个全知全能的AI助手，能够帮助用户完成解决任何难题。",
				},
				{
					"role":    "user",
					"content": prompt,
				},
			},
		}
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			return "", err
		}
		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)
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
			return "", fmt.Errorf("API error: %s", string(body))
		}
		var openapiResponse struct {
			Choices []struct {
				Message struct {
					Content string `json:"content,omitempty"`
				} `json:"message"`
			} `json:"choices"`
		}
		if err := json.Unmarshal(body, &openapiResponse); err != nil {
			return "", err
		}
		if len(openapiResponse.Choices) == 0 {
			return "", fmt.Errorf("no response from API")
		}
		return openapiResponse.Choices[0].Message.Content, nil
	}

	// 注册外置函数到Yaegi解释器
	exports := make(map[string]map[string]reflect.Value)

	// 拷贝标准库符号到 exports 映射中
	for path, pkg := range stdlib.Symbols {
		exports[path] = pkg
	}

	// 2. **构造自定义符号并添加到 exports**
	const pkgPath = "llm/llm"
	const funcName = "InvokeLLM"

	if _, ok := exports[pkgPath]; !ok {
		exports[pkgPath] = make(map[string]reflect.Value)
	}
	// 3. **注册 InvokeLLM 函数**
	exports[pkgPath][funcName] = reflect.ValueOf(wrapper)

	// 4. **注册文件处理函数到lojifile/lojifile包**
	const filePkgPath = "lojifile/lojifile"
	if _, ok := exports[filePkgPath]; !ok {
		exports[filePkgPath] = make(map[string]reflect.Value)
	}
	exports[filePkgPath]["ReadFile"] = reflect.ValueOf(external_functions.ReadFile)
	exports[filePkgPath]["ReadFileAsString"] = reflect.ValueOf(external_functions.ReadFileAsString)
	exports[filePkgPath]["GetFileExtension"] = reflect.ValueOf(external_functions.GetFileExtension)
	exports[filePkgPath]["GetFileSize"] = reflect.ValueOf(external_functions.GetFileSize)
	exports[filePkgPath]["WriteOutputFile"] = reflect.ValueOf(external_functions.WriteOutputFile)
	exports[filePkgPath]["WriteOutputFileFromBase64"] = reflect.ValueOf(external_functions.WriteOutputFileFromBase64)

	// 5. **注册相关插件函数到loji包**
	if len(relatedPlugins) > 0 {
		err := s.registerRelatedPluginsAsExternalFunctions(exports, relatedPlugins)
		if err != nil {
			return fmt.Errorf("failed to register related plugins: %w", err)
		}
	}

	// 创建 Yaegi 解释器，并将所有符号一次性传入
	i := interp.New(interp.Options{})
	if err := i.Use(exports); err != nil {
		return fmt.Errorf("无法注册符号表（标准库+自定义）：%v", err)
	}
	s.interpreter = i

	return nil
}

// reinitializeInterpreter 重新初始化Yaegi解释器
func (s *YaegiService) reinitializeInterpreter(relatedPlugins []models.PluginRelatedPlugin) error {
	// 创建新的Yaegi解释器实例

	// 先设置新的解释器实例，再注册外部函数

	// 重新注册InvokeLLM外部函数和相关插件
	if err := s.createNewInterpreter(relatedPlugins); err != nil {
		return fmt.Errorf("failed to register functions: %w", err)
	}

	return nil
}

// separateImportsAndLogic 从用户代码中分离import语句和逻辑代码
func (s *YaegiService) separateImportsAndLogic(code string) ([]string, string) {
	var imports []string
	var logicLines []string

	lines := strings.Split(code, "\n")
	inImportBlock := false

	for _, line := range lines {
		originalLine := line
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
		if inImportBlock && line != "" && !strings.HasPrefix(line, "//") {
			// 提取包名（去掉引号）
			if strings.Contains(line, `"`) {
				parts := strings.Fields(line)
				if len(parts) > 0 {
					pkg := strings.Trim(parts[0], `"`)
					if pkg != "" {
						imports = append(imports, `"`+pkg+`"`)
					}
				}
			}
		} else if !inImportBlock {
			// 检查单行import
			if strings.HasPrefix(line, `import "`) && strings.HasSuffix(line, `"`) {
				pkg := strings.TrimPrefix(line, `import "`)
				pkg = strings.TrimSuffix(pkg, `"`)
				if pkg != "" {
					imports = append(imports, `"`+pkg+`"`)
					continue // 跳过这个import行，不加入逻辑代码
				}
			}

			// 不在import块内且不是单行import，收集逻辑代码
			if line != "" {
				logicLines = append(logicLines, originalLine)
			}
		}
	}

	return imports, strings.Join(logicLines, "\n")
}

// createPluginWrapper 创建插件包装函数
func (s *YaegiService) createPluginWrapper(pluginID int64, pluginName string) reflect.Value {
	return reflect.ValueOf(func(input models.PluginInput) models.PluginOutput {
		// 简化版递归检测（可以后续扩展）
		// TODO: 实现完整的循环依赖检测

		// 执行插件
		plugin, err := s.pluginService.GetPluginByID(pluginID)
		if err != nil {
			return models.PluginOutput{PluginError: err.Error()}
		}

		// 序列化输入
		inputJSON, err := json.Marshal(input)
		if err != nil {
			return models.PluginOutput{PluginError: err.Error()}
		}

		// 获取依赖
		dependencies, _ := s.pluginService.GetPluginDependencies(pluginID)
		depStrings := make([]string, len(dependencies))
		for i, dep := range dependencies {
			depStrings[i] = dep.Package
		}

		// 获取相关插件（被调用的插件可能也需要调用其他插件）
		relatedPlugins, _ := s.pluginService.GetPluginRelatedPlugins(pluginID)

		// 创建独立的解释器实例来执行插件，确保数据隔离
		executor := &pluginExecutor{
			pluginService: s.pluginService,
			configService: s.configService,
			logFunc:       s.logFunc,
		}

		// 使用独立的执行器执行插件
		result, err := executor.executePlugin(pluginID, plugin.Code, string(inputJSON), depStrings, relatedPlugins)
		if err != nil {
			return models.PluginOutput{PluginError: err.Error()}
		}

		// 解析结果
		var output models.PluginOutput
		if err := json.Unmarshal([]byte(result), &output); err != nil {
			return models.PluginOutput{PluginError: err.Error()}
		}

		return output
	})
}

// registerRelatedPluginsAsExternalFunctions 注册相关插件作为外部函数到loji包
func (s *YaegiService) registerRelatedPluginsAsExternalFunctions(exports map[string]map[string]reflect.Value, relatedPlugins []models.PluginRelatedPlugin) error {
	const lojiPackageName = "loji/loji"

	if _, ok := exports[lojiPackageName]; !ok {
		exports[lojiPackageName] = make(map[string]reflect.Value)
	}

	// 为每个相关插件创建包装函数并注册
	for _, relatedPlugin := range relatedPlugins {
		// 获取插件详细信息
		plugin, err := s.pluginService.GetPluginByID(relatedPlugin.RelatedPluginID)
		if err != nil {
			log.Printf("Warning: failed to get related plugin %d: %v", relatedPlugin.RelatedPluginID, err)
			continue
		}

		wrapper := s.createPluginWrapper(plugin.ID, plugin.Name)
		exports[lojiPackageName][plugin.Name] = wrapper
	}

	log.Printf("Registered %d related plugins to loji/loji package", len(relatedPlugins))
	return nil
}

// containsLojiReference 检测代码是否包含loji引用
func (s *YaegiService) containsLojiReference(code string) bool {
	// 简单检测是否包含"loji."调用
	return strings.Contains(code, "loji.")
}

// executePlugin 使用独立的解释器实例执行插件，确保数据隔离
func (e *pluginExecutor) executePlugin(pluginID int64, code string, input string, dependencies []string, relatedPlugins []models.PluginRelatedPlugin) (string, error) {
	// 创建独立的解释器实例
	exports := make(map[string]map[string]reflect.Value)

	// 1. 注册标准库
	for path, pkg := range stdlib.Symbols {
		exports[path] = pkg
	}

	// 2. 注册LLM函数
	e.registerLLMFunctionForExecutor(exports)

	// 3. 注册相关插件（被调用的插件可能还需要调用其他插件）
	if len(relatedPlugins) > 0 {
		err := e.registerRelatedPluginsForExecutor(exports, relatedPlugins)
		if err != nil {
			return "", fmt.Errorf("failed to register related plugins: %w", err)
		}
	}

	// 4. 创建解释器实例
	interpreter := interp.New(interp.Options{})
	if err := interpreter.Use(exports); err != nil {
		return "", fmt.Errorf("无法注册符号表: %v", err)
	}

	// 5. 包装插件代码
	executionCode := e.wrapPluginCodeForExecutor(pluginID, code, input, dependencies, relatedPlugins)

	// 6. 执行代码
	_, err := interpreter.Eval(executionCode)
	if err != nil {
		return "", fmt.Errorf("failed to execute plugin code: %w", err)
	}
	log.Printf("dependency executionCode: %s", executionCode)

	// 7. 获取结果
	// outputVarName := fmt.Sprintf("pluginOutput", pluginID)
	globals := interpreter.Globals()
	if pluginOutput, exists := globals["pluginOutput"]; exists {
		if pluginOutput.CanInterface() {
			output := pluginOutput.Interface()
			jsonResult, err := json.Marshal(output)
			if err != nil {
				return "", fmt.Errorf("failed to marshal result: %w", err)
			}
			return string(jsonResult), nil
		}
	}

	return "", fmt.Errorf("failed to get plugin output")
}

// registerLLMFunctionForExecutor 为executor注册LLM函数
func (e *pluginExecutor) registerLLMFunctionForExecutor(exports map[string]map[string]reflect.Value) {
	const pkgPath = "llm/llm"
	const funcName = "InvokeLLM"

	if _, ok := exports[pkgPath]; !ok {
		exports[pkgPath] = make(map[string]reflect.Value)
	}

	// 创建LLM包装函数
	wrapper := func(prompt string) (string, error) {
		apiKey := e.getConfigValue("OPENAI_API_KEY")
		apiURL := e.getConfigValue("OPENAI_API_URL")
		modelType := e.getConfigValue("OPENAI_MODEL_TYPE")
		requestBody := map[string]interface{}{
			"model": modelType,
			"messages": []map[string]string{
				{
					"role":    "system",
					"content": "你是一个全知全能的AI助手，能够帮助用户完成解决任何难题。",
				},
				{
					"role":    "user",
					"content": prompt,
				},
			},
		}
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			return "", err
		}
		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)
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
			return "", fmt.Errorf("API error: %s", string(body))
		}
		var openapiResponse struct {
			Choices []struct {
				Message struct {
					Content string `json:"content,omitempty"`
				} `json:"message"`
			} `json:"choices"`
		}
		if err := json.Unmarshal(body, &openapiResponse); err != nil {
			return "", err
		}
		if len(openapiResponse.Choices) == 0 {
			return "", fmt.Errorf("no response from API")
		}
		return openapiResponse.Choices[0].Message.Content, nil
	}

	exports[pkgPath][funcName] = reflect.ValueOf(wrapper)
}

// registerRelatedPluginsForExecutor 为executor注册相关插件
func (e *pluginExecutor) registerRelatedPluginsForExecutor(exports map[string]map[string]reflect.Value, relatedPlugins []models.PluginRelatedPlugin) error {
	const lojiPackageName = "loji/loji"

	if _, ok := exports[lojiPackageName]; !ok {
		exports[lojiPackageName] = make(map[string]reflect.Value)
	}

	// 为每个相关插件创建包装函数并注册
	for _, relatedPlugin := range relatedPlugins {
		// 获取插件详细信息
		plugin, err := e.pluginService.GetPluginByID(relatedPlugin.RelatedPluginID)
		if err != nil {
			log.Printf("Warning: failed to get related plugin %d: %v", relatedPlugin.RelatedPluginID, err)
			continue
		}

		// 递归创建执行器来处理嵌套调用
		wrapper := e.createPluginWrapperForExecutor(plugin.ID, plugin.Name)
		exports[lojiPackageName][plugin.Name] = wrapper
	}

	return nil
}

// createPluginWrapperForExecutor 为executor创建插件包装函数
func (e *pluginExecutor) createPluginWrapperForExecutor(pluginID int64, pluginName string) reflect.Value {
	return reflect.ValueOf(func(input models.PluginInput) models.PluginOutput {
		// 创建新的执行器实例，确保递归调用也是隔离的
		subExecutor := &pluginExecutor{
			pluginService: e.pluginService,
			configService: e.configService,
			logFunc:       e.logFunc,
		}

		// 获取插件信息
		plugin, err := e.pluginService.GetPluginByID(pluginID)
		if err != nil {
			return models.PluginOutput{PluginError: err.Error()}
		}

		// 序列化输入
		inputJSON, err := json.Marshal(input)
		if err != nil {
			return models.PluginOutput{PluginError: err.Error()}
		}

		// 获取依赖
		dependencies, _ := e.pluginService.GetPluginDependencies(pluginID)
		depStrings := make([]string, len(dependencies))
		for i, dep := range dependencies {
			depStrings[i] = dep.Package
		}

		// 获取相关插件
		relatedPlugins, _ := e.pluginService.GetPluginRelatedPlugins(pluginID)

		// 使用子执行器执行
		result, err := subExecutor.executePlugin(pluginID, plugin.Code, string(inputJSON), depStrings, relatedPlugins)
		if err != nil {
			return models.PluginOutput{PluginError: err.Error()}
		}

		// 解析结果
		var output models.PluginOutput
		if err := json.Unmarshal([]byte(result), &output); err != nil {
			return models.PluginOutput{PluginError: err.Error()}
		}

		return output
	})
}

// wrapPluginCodeForExecutor 为executor包装插件代码
func (e *pluginExecutor) wrapPluginCodeForExecutor(pluginID int64, code string, input string, dependencies []string, relatedPlugins []models.PluginRelatedPlugin) string {
	// 从用户代码中提取逻辑代码
	_, logicCode := e.separateImportsAndLogic(code)

	// 调试输出
	log.Printf("执行器依赖: %v", dependencies)

	// 构建import语句
	allImports := models.GetDefaultImports()
	allImports = append(allImports, dependencies...)

	// 动态检测loji引用
	if e.containsLojiReference(logicCode) {
		allImports = append(allImports, `"loji/loji"`)
	}

	importBlock := strings.Join(allImports, "\n\t")

	// 生成唯一的变量名
	inputVarName := "pluginInput"
	outputVarName := "pluginOutput"

	return fmt.Sprintf(`
package main

import (
	%s
)

// PluginInput 插件输入结构体
type PluginInput struct {
	TextList     []string `+"`json:\"textList\"`"+`
	ImageList    []string `+"`json:\"imageList\"`"+`
	FileList     []string `+"`json:\"fileList\"`"+`
	AudioList    []string `+"`json:\"audioList\"`"+`
	VideoList    []string `+"`json:\"videoList\"`"+`
	DocumentList []string `+"`json:\"documentList\"`"+`
	OtherList    []string `+"`json:\"otherList\"`"+`
}

// PluginOutput 插件输出结构体
type PluginOutput struct {
	TextList     []string `+"`json:\"textList\"`"+`
	ImageList    []string `+"`json:\"imageList\"`"+`
	FileList     []string `+"`json:\"fileList\"`"+`
	AudioList    []string `+"`json:\"audioList\"`"+`
	VideoList    []string `+"`json:\"videoList\"`"+`
	DocumentList []string `+"`json:\"documentList\"`"+`
	OtherList    []string `+"`json:\"otherList\"`"+`
	PluginError  string   `+"`json:\"pluginError\"`"+`
}

const PluginInputJSON = %s

// 全局变量声明（使用唯一名称）
var (
	pluginInput PluginInput
	pluginOutput PluginOutput
)

// lojiLog 自定义日志函数
func lojiLog(message string) {
	%s
}

type OpenapiResponse struct {
	Choices []struct {
		Message struct {
			Content string `+"`json:\"content,omitempty\"`"+`
		} `+"`json:\"message\"`"+`
	} `+"`json:\"choices\"`"+`
}

// lojiLogf 带格式化的自定义日志函数
func lojiLogf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args)
	%s
}

// initPlugin 初始化插件数据
func initPlugin() {
	if err := json.Unmarshal([]byte(PluginInputJSON), &%s); err != nil {
		%s = PluginInput{} // 初始化为空结构体
		%s = PluginOutput{PluginError: "输入JSON解析失败: " + err.Error()}
	}
}

// executePlugin 执行插件逻辑
func executePlugin() {
	%s
}

// 主函数
func main() {
	initPlugin()
	executePlugin()
}
`, importBlock, e.formatInputJSON(input), e.getLogCall(), e.getLogfCall(), inputVarName, inputVarName, outputVarName, logicCode)
}

// separateImportsAndLogic 从用户代码中分离import语句和逻辑代码
func (e *pluginExecutor) separateImportsAndLogic(code string) ([]string, string) {
	var imports []string
	var logicLines []string

	lines := strings.Split(code, "\n")
	inImportBlock := false

	for _, line := range lines {
		originalLine := line
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
		if inImportBlock && line != "" && !strings.HasPrefix(line, "//") {
			// 提取包名（去掉引号）
			if strings.Contains(line, `"`) {
				parts := strings.Fields(line)
				if len(parts) > 0 {
					pkg := strings.Trim(parts[0], `"`)
					if pkg != "" {
						imports = append(imports, `"`+pkg+`"`)
					}
				}
			}
		} else if !inImportBlock {
			// 检查单行import
			if strings.HasPrefix(line, `import "`) && strings.HasSuffix(line, `"`) {
				pkg := strings.TrimPrefix(line, `import "`)
				pkg = strings.TrimSuffix(pkg, `"`)
				if pkg != "" {
					imports = append(imports, `"`+pkg+`"`)
					continue // 跳过这个import行，不加入逻辑代码
				}
			}

			// 不在import块内且不是单行import，收集逻辑代码
			if line != "" {
				logicLines = append(logicLines, originalLine)
			}
		}
	}

	return imports, strings.Join(logicLines, "\n")
}

// formatInputJSON 格式化输入JSON
func (e *pluginExecutor) formatInputJSON(input string) string {
	// 由于使用反引号字符串，只需要处理反引号
	escaped := strings.ReplaceAll(input, "`", "` + \"`\" + `")
	return "`" + escaped + "`"
}

// getLogCall 生成日志函数调用代码
func (e *pluginExecutor) getLogCall() string {
	if e.logFunc != nil {
		return `log.Printf("[PLUGIN LOG] %s", message)`
	}
	return `// 日志功能未启用`
}

// getLogfCall 生成格式化日志函数调用代码
func (e *pluginExecutor) getLogfCall() string {
	if e.logFunc != nil {
		return `log.Printf("[PLUGIN LOG] %s", message)`
	}
	return `// 日志功能未启用`
}

// containsLojiReference 检测代码是否包含loji引用
func (e *pluginExecutor) containsLojiReference(code string) bool {
	return strings.Contains(code, "loji.")
}

// Cleanup 清理资源
func (s *YaegiService) Cleanup() error {
	if s.interpreter != nil {
		// Yaegi解释器没有明确的清理方法，但可以重置
		s.interpreter = nil
	}
	return nil
}
