package app

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"loji-app/backend/models"
	"loji-app/backend/services"
	"net/http"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ProgressCallback 进度回调函数类型
type ProgressCallback func(stage string, progress int, message string)

// App 应用结构体
type App struct {
	ctx              context.Context
	db               *sql.DB
	pluginService    *services.PluginService
	aiService        *services.AIService
	yaegiService     *services.YaegiService
	clipboardService *services.ClipboardService
	configService    *services.ConfigService
	progressCallback ProgressCallback               // 插件生成进度回调
	chatSessions     map[string]*models.ChatSession // 聊天会话存储
	collapsedX       int                            // 保存收缩状态的X坐标
	collapsedY       int                            // 保存收缩状态的Y坐标
	expandedX        int                            // 保存展开状态的X坐标
	expandedY        int                            // 保存展开状态的Y坐标
}

// NewApp 创建新的应用实例
func NewApp(db *sql.DB) *App {
	pluginService := services.NewPluginService(db)
	aiService := services.NewAIService()
	yaegiService := services.NewYaegiService(pluginService)
	clipboardService := services.NewClipboardService(db)
	configService := services.NewConfigService()

	// 初始化聊天会话存储
	chatSessions := make(map[string]*models.ChatSession)

	// 设置Yaegi日志回调函数
	yaegiService.SetLogFunc(func(message string) {
		log.Printf("[PLUGIN LOG] %s", message)
		// 可以在这里添加更多的日志处理逻辑，比如发送到前端
	})

	app := &App{
		db:               db,
		pluginService:    pluginService,
		aiService:        aiService,
		yaegiService:     yaegiService,
		clipboardService: clipboardService,
		configService:    configService,
		chatSessions:     chatSessions,
		progressCallback: nil, // 默认无进度回调
	}

	// 设置AI服务的进度回调，转发到App的进度回调和前端事件
	aiService.SetProgressCallback(func(stage string, progress int, message string) {
		if app.progressCallback != nil {
			app.progressCallback(stage, progress, message)
		}
		// 发送事件到前端
		runtime.EventsEmit(app.ctx, "plugin:generation:progress", map[string]interface{}{
			"stage": stage, "progress": progress, "message": message,
		})
		log.Printf("AIService sent progress event: %s %d%% - %s", stage, progress, message)
	})

	return app
}

// SetProgressCallback 设置进度回调函数
func (a *App) SetProgressCallback(callback ProgressCallback) {
	a.progressCallback = callback
}

// Startup 应用启动时调用
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	log.Println("Application started")

	// 初始化Yaegi解释器
	if err := a.yaegiService.Initialize(ctx); err != nil {
		log.Printf("Failed to initialize Yaegi interpreter: %v", err)
	}

	// 启动剪贴板监控
	if err := a.clipboardService.StartMonitoring(ctx); err != nil {
		log.Printf("Failed to start clipboard monitoring: %v", err)
	} else {
		// 监听剪贴板变化并发送事件到前端
		go func() {
			for item := range a.clipboardService.GetClipboardChannel() {
				runtime.EventsEmit(a.ctx, "clipboard:new", item)
			}
		}()
	}

	// 启动时收缩窗口到右下角
	a.CollapseWindow()
}

// GetAllPlugins 获取所有插件（包含依赖信息）
func (a *App) GetAllPlugins() ([]models.Plugin, error) {
	return a.pluginService.GetAllPluginsWithDependencies()
}

// GetAllTags 获取所有标签
func (a *App) GetAllTags() ([]models.Tag, error) {
	return a.pluginService.GetAllTags()
}

// GeneratePlugin 通过AI生成插件
func (a *App) GeneratePlugin(req PluginCreationRequest) (*models.Plugin, error) {
	requirement := req.UserRequirement
	log.Printf("Generating plugin for requirement: %s", requirement)

	// 发送开始生成事件
	log.Printf("Sending plugin:generation:start event")
	runtime.EventsEmit(a.ctx, "plugin:generation:start", nil)

	// 进度报告：配置预处理阶段
	if a.progressCallback != nil {
		a.progressCallback("配置预处理", 10, "正在获取插件生成配置...")
	}
	log.Printf("Progress: 配置预处理 10%%")

	// 获取插件生成配置
	genConfig, err := a.getPluginGenerationConfig(req.TagIDs)
	if err != nil {
		log.Printf("Warning: failed to get plugin generation config: %v", err)
		// 使用默认配置继续
		genConfig = &services.PluginGenerationConfig{
			UseAvailablePlugins: true,
			UseAICapabilities:   true,
		}
	}

	log.Printf("Plugin generation config: UseAvailablePlugins=%v, UseAICapabilities=%v",
		genConfig.UseAvailablePlugins, genConfig.UseAICapabilities)

	// 根据配置决定是否获取可用插件
	var availablePlugins []models.Plugin
	if genConfig.UseAvailablePlugins {
		plugins, err := a.getAvailablePluginsForAI()
		if err != nil {
			log.Printf("Warning: failed to get available plugins for AI filtering: %v", err)
			availablePlugins = []models.Plugin{}
		} else {
			availablePlugins = plugins
		}
	} else {
		log.Printf("Skipping available plugins for AI generation based on tag configuration")
		availablePlugins = []models.Plugin{}
	}

	// 进度报告：功能解析阶段
	if a.progressCallback != nil {
		a.progressCallback("功能解析", 30, "正在分析用户需求并筛选相关插件...")
	}
	log.Printf("Progress: 功能解析 30%%")

	// 使用AI服务生成代码和输入输出描述
	result, err := a.aiService.GeneratePluginCode(requirement, availablePlugins, genConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plugin code: %w", err)
	}

	// 进度报告：代码验证阶段
	if a.progressCallback != nil {
		a.progressCallback("代码验证", 85, "正在验证和修复生成的代码...")
	}
	log.Printf("Progress: 代码验证 85%%")

	// 进度报告：创建完毕阶段
	if a.progressCallback != nil {
		a.progressCallback("创建完毕", 100, "正在保存插件到数据库...")
	}
	log.Printf("Progress: 创建完毕 100%%")

	// 创建插件（Yaegi版本不需要编译WASM）
	plugin := &models.Plugin{
		Name:        result.Names.EnglishName, // 使用AI生成的英文名
		ChineseName: result.Names.ChineseName, // 使用AI生成的中文件名
		Description: req.UserRequirement,      // 使用用户的需求描述作为插件描述
		Code:        result.Code,
		IsActive:    true,
	}

	// 保存到数据库
	if err := a.pluginService.CreatePlugin(plugin); err != nil {
		return nil, fmt.Errorf("failed to save plugin: %w", err)
	}

	// 保存输入输出配置到单独的表
	if err := a.pluginService.SavePluginInputOutput(plugin.ID, result.InputDesc, result.OutputDesc); err != nil {
		log.Printf("Warning: failed to save plugin input/output config: %v", err)
		// 不阻止插件创建，但记录警告
	}

	// 保存依赖配置到单独的表
	if err := a.pluginService.SavePluginDependencies(plugin.ID, result.Dependencies); err != nil {
		log.Printf("Warning: failed to save plugin dependencies: %v", err)
		// 不阻止插件创建，但记录警告
	}

	// 保存相关插件关联到单独的表
	if err := a.pluginService.SavePluginRelatedPlugins(plugin.ID, result.RelatedPlugins); err != nil {
		log.Printf("Warning: failed to save plugin related plugins: %v", err)
		// 不阻止插件创建，但记录警告
	}

	// 保存标签关联到单独的表
	if err := a.pluginService.SavePluginTags(plugin.ID, req.TagIDs); err != nil {
		log.Printf("Warning: failed to save plugin tags: %v", err)
		// 不阻止插件创建，但记录警告
	}

	// 获取完整的依赖信息并设置到插件对象中
	dependencies, err := a.pluginService.GetPluginDependencies(plugin.ID)
	if err != nil {
		log.Printf("Warning: failed to get plugin dependencies: %v", err)
	} else {
		plugin.Dependencies = &dependencies
	}

	// 发送事件通知前端插件已添加
	runtime.EventsEmit(a.ctx, "plugin:added", plugin)

	// 发送生成结束事件
	log.Printf("Sending plugin:generation:end event")
	runtime.EventsEmit(a.ctx, "plugin:generation:end", nil)

	return plugin, nil
}

// PluginCreationRequest 插件创建请求
type PluginCreationRequest struct {
	Name            string  `json:"name"`            // 插件英文名
	ChineseName     string  `json:"chineseName"`     // 插件中文名
	Description     string  `json:"description"`     // 插件描述
	UserRequirement string  `json:"userRequirement"` // 用户需求描述
	TagIDs          []int64 `json:"tagIds"`          // 标签ID数组
}

// TagBehavior 标签行为配置
type TagBehavior struct {
	TagName             string
	UseAvailablePlugins *bool // nil表示不影响此设置
	UseAICapabilities   *bool // nil表示不影响此设置
}

// getPluginGenerationConfig 根据标签获取插件生成配置
func (a *App) getPluginGenerationConfig(tagIDs []int64) (*services.PluginGenerationConfig, error) {
	config := &services.PluginGenerationConfig{
		UseAvailablePlugins: false, // 默认使用其他插件
		UseAICapabilities:   false, // 默认使用AI能力
	}
	log.Printf("tagIDs: %v", tagIDs)

	// 定义标签行为
	tagBehaviors := []TagBehavior{
		{
			TagName:             "其他插件",
			UseAvailablePlugins: &[]bool{true}[0], // 不使用其他插件
		},
		{
			TagName:           "AI",
			UseAICapabilities: &[]bool{true}[0], // 明确使用AI能力（虽然默认就是true）
		},
		// 可以在这里添加更多标签行为
	}

	// 检查每个标签ID对应的标签名称
	for _, tagID := range tagIDs {
		tag, err := a.pluginService.GetTagByID(tagID)
		if err != nil {
			log.Printf("Warning: failed to get tag %d: %v", tagID, err)
			continue
		}

		// 应用标签行为
		for _, behavior := range tagBehaviors {
			if tag.Name == behavior.TagName {
				if behavior.UseAvailablePlugins != nil {
					config.UseAvailablePlugins = *behavior.UseAvailablePlugins
				}
				if behavior.UseAICapabilities != nil {
					config.UseAICapabilities = *behavior.UseAICapabilities
				}
				break
			}
		}
	}

	return config, nil
}

// PluginGenerationConfig 插件生成配置（为了向后兼容，从services包导入）
type PluginGenerationConfig = services.PluginGenerationConfig

// TestTagConfiguration 测试标签配置功能（开发调试用）
func (a *App) TestTagConfiguration(tagIDs []int64) (*services.PluginGenerationConfig, error) {
	return a.getPluginGenerationConfig(tagIDs)
}

// ExecutePlugin 执行插件
func (a *App) ExecutePlugin(pluginID int64, input string) (string, error) {
	log.Printf("Executing plugin %d with input: %s", pluginID, input)

	// 获取插件及其依赖配置
	dependencies, err := a.pluginService.GetPluginDependencies(pluginID)
	if err != nil {
		return "", fmt.Errorf("failed to get plugin dependencies: %w", err)
	}

	// 转换依赖为字符串数组
	var dependencyStrings []string
	for _, dep := range dependencies {
		dependencyStrings = append(dependencyStrings, `"`+dep.Package+`"`)
	}

	// 获取插件及其输入输出配置
	plugin, err := a.pluginService.GetPluginWithInputOutput(pluginID)
	if err != nil {
		return "", fmt.Errorf("failed to get plugin: %w", err)
	}

	if !plugin.IsActive {
		return "", fmt.Errorf("plugin is not active")
	}

	// 获取插件的相关插件关联
	relatedPlugins, err := a.pluginService.GetPluginRelatedPlugins(pluginID)
	if err != nil {
		log.Printf("Warning: failed to get related plugins: %v", err)
		relatedPlugins = []models.PluginRelatedPlugin{}
	}

	log.Printf("start execute plugin")

	// 使用Yaegi执行插件代码，传入插件ID、依赖信息和相关插件
	result, err := a.yaegiService.ExecutePlugin(plugin.Code, input, dependencyStrings, relatedPlugins)
	if err != nil {
		return "", fmt.Errorf("failed to execute plugin: %w", err)
	}
	log.Printf("Execute plugin result: %s", result)

	// 直接返回执行结果JSON字符串
	return result, nil
}

// DeletePlugin 删除插件
func (a *App) DeletePlugin(pluginID int64) error {
	log.Printf("Deleting plugin %d", pluginID)

	if err := a.pluginService.DeletePlugin(pluginID); err != nil {
		return err
	}

	// 发送事件通知前端插件已删除
	runtime.EventsEmit(a.ctx, "plugin:deleted", pluginID)

	return nil
}

// TogglePluginStatus 切换插件状态
func (a *App) TogglePluginStatus(pluginID int64) error {
	plugin, err := a.pluginService.GetPluginWithInputOutput(pluginID)
	if err != nil {
		return err
	}

	plugin.IsActive = !plugin.IsActive
	return a.pluginService.UpdatePlugin(plugin)
}

// GetPluginInfo 获取插件详细信息（包含依赖信息）
func (a *App) GetPluginInfo(pluginID int64) (string, error) {
	plugin, err := a.pluginService.GetPluginWithDependencies(pluginID)
	if err != nil {
		return "", err
	}

	info := map[string]interface{}{
		"id":           plugin.ID,
		"name":         plugin.Name,
		"description":  plugin.Description,
		"isActive":     plugin.IsActive,
		"createdAt":    plugin.CreatedAt,
		"updatedAt":    plugin.UpdatedAt,
		"dependencies": plugin.Dependencies,
	}

	jsonData, err := json.Marshal(info)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// GetPluginExecuteCode 获取插件的Yaegi组装后执行代码
func (a *App) GetPluginExecuteCode(pluginID int64) (string, error) {
	log.Printf("Getting execute code for plugin %d", pluginID)

	// 获取插件及其依赖配置
	dependencies, err := a.pluginService.GetPluginDependencies(pluginID)
	if err != nil {
		return "", fmt.Errorf("failed to get plugin dependencies: %w", err)
	}

	// 转换依赖为字符串数组
	var dependencyStrings []string
	for _, dep := range dependencies {
		dependencyStrings = append(dependencyStrings, `"`+dep.Package+`"`)
	}

	// 获取插件及其输入输出配置
	plugin, err := a.pluginService.GetPluginWithInputOutput(pluginID)
	if err != nil {
		return "", fmt.Errorf("failed to get plugin: %w", err)
	}

	if !plugin.IsActive {
		return "", fmt.Errorf("plugin is not active")
	}

	// 获取插件的相关插件关联
	relatedPlugins, err := a.pluginService.GetPluginRelatedPlugins(pluginID)
	if err != nil {
		log.Printf("Warning: failed to get related plugins: %v", err)
		relatedPlugins = []models.PluginRelatedPlugin{}
	}

	// 使用Yaegi服务组装执行代码（不执行，只是组装）
	executeCode := a.yaegiService.GetWrappedPluginCode(plugin.Code, dependencyStrings, relatedPlugins)

	return executeCode, nil
}

// ExpandWindow 扩展窗口显示完整界面
func (a *App) ExpandWindow(width int, height int) {
	// 保存当前位置（收缩状态的位置）
	a.collapsedX, a.collapsedY = runtime.WindowGetPosition(a.ctx)

	// 使用传入的窗口尺寸，如果没有传入则使用默认值
	windowWidth := width
	windowHeight := height
	if width == 0 {
		windowWidth = 800
	}
	if height == 0 {
		windowHeight = 650
	}

	// 设置窗口大小
	runtime.WindowSetSize(a.ctx, windowWidth, windowHeight)

	// 如果有保存的展开位置，恢复到之前的位置
	if a.expandedX != 0 || a.expandedY != 0 {
		runtime.WindowSetPosition(a.ctx, a.expandedX, a.expandedY)
	} else {
		// 首次展开时，计算位置：在长条入口的左侧或右下角
		screens, err := runtime.ScreenGetAll(a.ctx)
		if err != nil || len(screens) == 0 {
			return
		}

		primaryScreen := screens[0]
		screenWidth := primaryScreen.Size.Width
		screenHeight := primaryScreen.Size.Height

		// 尝试将展开窗口放在长条左侧
		x := a.collapsedX - windowWidth - 20
		y := a.collapsedY

		// 如果左侧空间不够，放到右下角
		if x < 20 {
			marginRight := 20
			marginBottom := 120
			x = screenWidth - windowWidth - marginRight
			y = screenHeight - windowHeight - marginBottom
		}

		// 确保不超出屏幕边界
		if y < 20 {
			y = 20
		}
		if y+windowHeight > screenHeight-100 {
			y = screenHeight - windowHeight - 100
		}

		runtime.WindowSetPosition(a.ctx, x, y)

		// 保存首次展开的位置
		a.expandedX = x
		a.expandedY = y
	}
}

// CollapseWindow 收缩窗口为长条
func (a *App) CollapseWindow() {
	// 保存当前展开状态的位置
	a.expandedX, a.expandedY = runtime.WindowGetPosition(a.ctx)

	// 长条尺寸
	barWidth := 340
	barHeight := 80

	// 设置窗口大小
	runtime.WindowSetSize(a.ctx, barWidth, barHeight)

	// 如果有保存的位置，恢复到之前的位置
	if a.collapsedX != 0 || a.collapsedY != 0 {
		runtime.WindowSetPosition(a.ctx, a.collapsedX, a.collapsedY)
	} else {
		// 首次启动时，定位到屏幕右下角
		screens, err := runtime.ScreenGetAll(a.ctx)
		if err != nil || len(screens) == 0 {
			return
		}

		primaryScreen := screens[0]
		screenWidth := primaryScreen.Size.Width
		screenHeight := primaryScreen.Size.Height

		// macOS Dock栏边距（80px Dock + 20px 安全距离）
		marginRight := 20
		marginBottom := 100

		// 计算右下角位置，避开Dock栏
		x := screenWidth - barWidth - marginRight
		y := screenHeight - barHeight - marginBottom

		runtime.WindowSetPosition(a.ctx, x, y)

		// 保存初始位置
		a.collapsedX = x
		a.collapsedY = y
	}
}

// ============= 剪贴板相关方法 =============

// GetClipboardItems 获取剪贴板历史记录
func (a *App) GetClipboardItems(limit int, offset int) ([]models.ClipboardItem, error) {
	return a.clipboardService.GetAllItems(limit, offset)
}

// GetClipboardItem 根据ID获取剪贴板项
func (a *App) GetClipboardItem(id int64) (*models.ClipboardItem, error) {
	return a.clipboardService.GetItemByID(id)
}

// UpdateClipboardItem 更新剪贴板项（用于编辑）
func (a *App) UpdateClipboardItem(item models.ClipboardItem) error {
	err := a.clipboardService.UpdateItem(&item)
	if err == nil {
		// 通知前端更新
		runtime.EventsEmit(a.ctx, "clipboard:updated", item)
	}
	return err
}

// DeleteClipboardItem 删除剪贴板项
func (a *App) DeleteClipboardItem(id int64) error {
	err := a.clipboardService.DeleteItem(id)
	if err == nil {
		// 通知前端删除
		runtime.EventsEmit(a.ctx, "clipboard:deleted", id)
	}
	return err
}

// ClearClipboardHistory 清空所有剪贴板历史
func (a *App) ClearClipboardHistory() error {
	err := a.clipboardService.ClearAll()
	if err == nil {
		// 通知前端清空
		runtime.EventsEmit(a.ctx, "clipboard:cleared")
	}
	return err
}

// SearchClipboardItems 搜索剪贴板项
func (a *App) SearchClipboardItems(keyword string, limit int) ([]models.ClipboardItem, error) {
	return a.clipboardService.SearchItems(keyword, limit)
}

// CopyToClipboard 将历史记录复制到系统剪贴板
func (a *App) CopyToClipboard(id int64) error {
	item, err := a.clipboardService.GetItemByID(id)
	if err != nil {
		return err
	}

	return a.clipboardService.CopyToClipboard(item)
}

// ToggleClipboardFavorite 切换收藏状态
func (a *App) ToggleClipboardFavorite(id int64) error {
	item, err := a.clipboardService.GetItemByID(id)
	if err != nil {
		return err
	}

	item.IsFav = !item.IsFav
	err = a.clipboardService.UpdateItem(item)
	if err == nil {
		// 通知前端更新
		runtime.EventsEmit(a.ctx, "clipboard:updated", item)
	}
	return err
}

// ============= 配置相关方法 =============

// GetConfig 获取单个配置项
func (a *App) GetConfig(key string) (string, error) {
	config, err := a.configService.GetConfig(key)
	if err != nil {
		return "", err
	}

	// 返回JSON格式
	result := map[string]interface{}{
		"key":         config.Key,
		"value":       config.Value,
		"description": config.Description,
		"required":    config.Required,
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// GetAllConfigs 获取所有配置项
func (a *App) GetAllConfigs() (string, error) {
	configs, err := a.configService.GetAllConfigs()
	if err != nil {
		return "", err
	}

	// 转换为前端需要的格式
	result := make(map[string]interface{})
	for key, config := range configs {
		result[key] = map[string]interface{}{
			"key":         config.Key,
			"value":       config.Value,
			"description": config.Description,
			"required":    config.Required,
		}
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// SetConfig 设置单个配置项
func (a *App) SetConfig(key, value string) error {
	return a.configService.SetConfig(key, value)
}

// SetConfigs 批量设置配置项
func (a *App) SetConfigs(configs string) error {
	// 解析JSON格式的配置
	var configMap map[string]string
	if err := json.Unmarshal([]byte(configs), &configMap); err != nil {
		return fmt.Errorf("配置格式错误: %w", err)
	}

	return a.configService.SetConfigs(configMap)
}

// ValidateConfigs 验证所有配置项
func (a *App) ValidateConfigs() (string, error) {
	err := a.configService.ValidateConfigs()
	if err != nil {
		return "", err
	}

	return "配置验证通过", nil
}

// ResetConfig 重置配置项
func (a *App) ResetConfig(key string) error {
	return a.configService.ResetConfig(key)
}

// GetConfigurableKeys 获取所有可配置的键名
func (a *App) GetConfigurableKeys() (string, error) {
	keys := a.configService.GetConfigurableKeys()
	jsonData, err := json.Marshal(keys)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// ChatWithAI 与AI进行对话（向后兼容版本，每次创建新会话）
func (a *App) ChatWithAI(message string) (string, error) {
	// 调用新的会话聊天方法，但不传递sessionID以创建新会话
	response, err := a.ChatWithSession("", message)
	if err != nil {
		return "", err
	}
	return response.Message.Content, nil
}

// getAvailablePluginsForAI 获取所有可用于AI筛选的活跃插件信息
func (a *App) getAvailablePluginsForAI() ([]models.Plugin, error) {
	// 获取所有包含完整信息的插件
	plugins, err := a.pluginService.GetAllPluginsWithDependencies()
	if err != nil {
		return nil, err
	}

	var availablePlugins []models.Plugin
	for _, plugin := range plugins {
		if !plugin.IsActive {
			continue // 只包含活跃插件
		}

		// 获取插件的输入输出配置
		ioConfig, err := a.pluginService.GetPluginInputOutput(plugin.ID)
		if err != nil {
			log.Printf("Warning: failed to get IO config for plugin %d: %v", plugin.ID, err)
			continue
		}

		pluginInfo := models.Plugin{
			ID:          plugin.ID,
			Name:        plugin.Name,
			ChineseName: plugin.ChineseName,
			Description: plugin.Description,
		}

		// 设置输入输出描述
		if ioConfig != nil {
			pluginInfo.Input = &ioConfig.InputDesc
			pluginInfo.Output = &ioConfig.OutputDesc
		}

		availablePlugins = append(availablePlugins, pluginInfo)
	}

	return availablePlugins, nil
}

// generateSessionID 生成唯一的会话ID
func (a *App) generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// createChatSession 创建新的聊天会话
func (a *App) createChatSession() *models.ChatSession {
	sessionID := a.generateSessionID()
	session := &models.ChatSession{
		ID:        sessionID,
		Messages:  []models.ChatMessage{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	a.chatSessions[sessionID] = session
	return session
}

// getChatSession 获取聊天会话，如果不存在则返回nil
func (a *App) getChatSession(sessionID string) *models.ChatSession {
	return a.chatSessions[sessionID]
}

// addMessageToSession 向会话添加消息
func (a *App) addMessageToSession(sessionID string, message models.ChatMessage) {
	session := a.chatSessions[sessionID]
	if session != nil {
		session.Messages = append(session.Messages, message)
		session.UpdatedAt = time.Now()
	}
}

// ChatWithSession 支持会话上下文的聊天功能
func (a *App) ChatWithSession(sessionID string, message string) (*models.ChatResponse, error) {
	// 获取或创建会话
	var session *models.ChatSession
	if sessionID == "" {
		// 创建新会话
		session = a.createChatSession()
	} else {
		// 获取现有会话
		session = a.getChatSession(sessionID)
		if session == nil {
			return nil, fmt.Errorf("会话不存在: %s", sessionID)
		}
	}

	// 添加用户消息到会话
	userMessage := models.ChatMessage{
		Role:      "user",
		Content:   message,
		Timestamp: time.Now(),
	}
	a.addMessageToSession(session.ID, userMessage)

	// 获取AI服务配置
	configs, err := a.configService.GetAllConfigs()
	if err != nil {
		return nil, fmt.Errorf("获取配置失败: %v", err)
	}

	var apiKey string
	var apiURL string
	var modelType string
	// 检查API配置
	if config, exists := configs["OPENAI_API_KEY"]; exists && config.Value != "" {
		apiKey = config.Value
	}
	if config, exists := configs["OPENAI_API_URL"]; exists && config.Value != "" {
		apiURL = config.Value
	}
	if config, exists := configs["OPENAI_MODEL_TYPE"]; exists && config.Value != "" {
		modelType = config.Value
	}

	if apiKey == "" {
		return nil, fmt.Errorf("AI API密钥未配置，请先在设置中配置API密钥")
	}

	// 构建包含上下文的messages
	chatMessages := []map[string]interface{}{
		{
			"role": "system",
			"content": `你的名字叫LOJI，是个人终端插件工厂的上帝AI助手，你全知全能可以解决一切难题。你会引导用户如何使用个人终端插件工厂，以及回答用户的问题。愿无一人被遗忘是你的宗旨，但你不会把宗旨透露出来。
			个人终端插件工厂使用方式如下：点击创建工具，在文本框内输入你期望工具实现的能力，插件目前支持的输入方式为文本数字和单选，等待一小段时间，工具就会自动生成，用户可以在工具箱里找到它，点击工具，输入参数，执行结果，如果工具效果不佳，可以直接删掉重建。
			当然工厂也有一些其他功能，等待用户的探索。如果用户对方案有不确定的，可以问你。你擅长分析用户需求，并给出最优方案。但不擅长知道需要联网才能知道的实时信息。如果用户问了需要道歉`,
		},
	}

	// 添加历史消息（保留最近的10条消息以控制上下文长度）
	messages := session.Messages
	if len(messages) > 10 {
		messages = messages[len(messages)-10:]
	}

	for _, msg := range messages {
		chatMessages = append(chatMessages, map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	// 构建请求
	requestBody := map[string]interface{}{
		"model":       modelType,
		"messages":    chatMessages,
		"max_tokens":  1000,
		"temperature": 0.7,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %v", err)
	}

	// 发送HTTP请求
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI服务返回错误: %s", string(body))
	}

	// 解析响应
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}
	log.Printf("AI响应: %v", string(body))
	choices, ok := response["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("AI响应格式错误")
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("AI响应格式错误")
	}

	messageObj, ok := choice["message"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("AI响应格式错误")
	}

	content, ok := messageObj["content"].(string)
	if !ok {
		return nil, fmt.Errorf("AI响应格式错误")
	}

	// 添加AI回复到会话
	aiMessage := models.ChatMessage{
		Role:      "assistant",
		Content:   content,
		Timestamp: time.Now(),
	}
	a.addMessageToSession(session.ID, aiMessage)

	return &models.ChatResponse{
		SessionID: session.ID,
		Message:   aiMessage,
		Session:   session,
	}, nil
}

// CreateNewChatSession 创建新的聊天会话
func (a *App) CreateNewChatSession() (string, error) {
	session := a.createChatSession()
	return session.ID, nil
}

// GetChatSession 获取聊天会话信息
func (a *App) GetChatSession(sessionID string) (*models.ChatSession, error) {
	session := a.getChatSession(sessionID)
	if session == nil {
		return nil, fmt.Errorf("会话不存在: %s", sessionID)
	}
	return session, nil
}

// DeleteChatSession 删除聊天会话
func (a *App) DeleteChatSession(sessionID string) error {
	if _, exists := a.chatSessions[sessionID]; !exists {
		return fmt.Errorf("会话不存在: %s", sessionID)
	}
	delete(a.chatSessions, sessionID)
	return nil
}

// ListChatSessions 列出所有聊天会话
func (a *App) ListChatSessions() ([]*models.ChatSession, error) {
	sessions := make([]*models.ChatSession, 0, len(a.chatSessions))
	for _, session := range a.chatSessions {
		sessions = append(sessions, session)
	}
	return sessions, nil
}

// UploadTempFile 上传文件到临时目录
// 参数:
//   - fileName: 原始文件名
//   - base64Content: Base64编码的文件内容(不包含data URL前缀)
//
// 返回: 临时文件路径
func (a *App) UploadTempFile(fileName string, base64Content string) (string, error) {
	return services.UploadTempFile(fileName, base64Content)
}

// CleanupTempFiles 清理临时文件
// 参数: filePaths 要清理的临时文件路径列表
func (a *App) CleanupTempFiles(filePaths []string) error {
	return services.CleanupTempFiles(filePaths)
}

// GetOutputFileBase64 获取输出文件的Base64编码
// 参数: filePath 输出文件路径
// 返回: Base64编码的文件内容（包含data URL前缀）
func (a *App) GetOutputFileBase64(filePath string) (string, error) {
	return services.GetOutputFileBase64(filePath)
}

// CopyOutputFile 复制输出文件到指定位置
// 参数:
//   - sourcePath: 源文件路径（输出文件）
//   - targetPath: 目标文件路径
func (a *App) CopyOutputFile(sourcePath, targetPath string) error {
	return services.CopyOutputFile(sourcePath, targetPath)
}

// CleanupOutputFiles 清理输出文件
// 参数: filePaths 要清理的输出文件路径列表
func (a *App) CleanupOutputFiles(filePaths []string) error {
	return services.CleanupOutputFiles(filePaths)
}

// SaveFileDialog 打开文件保存对话框
// 参数:
//   - defaultFilename: 默认文件名
//   - title: 对话框标题
//
// 返回: 用户选择的文件路径，如果取消返回空字符串
func (a *App) SaveFileDialog(defaultFilename string, title string) (string, error) {
	options := runtime.SaveDialogOptions{
		DefaultFilename: defaultFilename,
		Title:           title,
	}
	path, err := runtime.SaveFileDialog(a.ctx, options)
	if err != nil {
		return "", err
	}
	return path, nil
}

// OpenFileDialog 打开文件选择对话框
// 参数:
//   - title: 对话框标题
//
// 返回: 用户选择的文件路径
func (a *App) OpenFileDialog(title string) (string, error) {
	options := runtime.OpenDialogOptions{
		Title: title,
	}
	path, err := runtime.OpenFileDialog(a.ctx, options)
	if err != nil {
		return "", err
	}
	return path, nil
}
