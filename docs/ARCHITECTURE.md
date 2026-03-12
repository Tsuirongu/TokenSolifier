# loji App 架构设计文档

## 系统概述

loji App 是一个基于 Wails 框架的跨平台桌面应用，采用前后端分离架构，使用 AI 自动生成工具插件，通过 WebAssembly 执行用户动态创建的功能模块。

### 设计目标

- ✅ **轻量级**：最小化资源占用，快速启动
- ✅ **高性能**：WASM 提供接近原生的执行速度
- ✅ **跨平台**：支持 macOS、Windows、Linux
- ✅ **可扩展**：插件化架构，支持热更新
- ✅ **安全性**：WASM 沙箱隔离，防止恶意代码
- ✅ **易用性**：简洁的 UI，智能的 AI 辅助

---

## 技术栈

### 前端
- **框架**: 原生 HTML/CSS/JavaScript
- **构建工具**: Vite 5.x
- **UI设计**: 无边框现代化设计
- **通信**: Wails Runtime API

### 后端
- **语言**: Go 1.21+
- **框架**: Wails v2.8+
- **数据库**: SQLite (modernc.org/sqlite)
- **执行引擎**: WebAssembly (GOOS=js GOARCH=wasm)

### AI服务
- **模型**: GPT-4 (可配置)
- **API**: OpenAI Compatible API
- **Prompt管理**: 文件系统

---

## 架构图

### 整体架构

```
┌─────────────────────────────────────────────────┐
│                  前端层 (Frontend)                │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐      │
│  │   UI     │  │  Event   │  │  State   │      │
│  │ Component│  │  Handler │  │  Manager │      │
│  └──────────┘  └──────────┘  └──────────┘      │
└─────────────────────────────────────────────────┘
                      ↕ Wails Bridge
┌─────────────────────────────────────────────────┐
│                 应用层 (App Layer)                │
│  ┌──────────────────────────────────────────┐   │
│  │              App Controller              │   │
│  │  - API 路由                              │   │
│  │  - 事件分发                              │   │
│  │  - 状态管理                              │   │
│  └──────────────────────────────────────────┘   │
└─────────────────────────────────────────────────┘
                      ↕
┌─────────────────────────────────────────────────┐
│                服务层 (Service Layer)            │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐      │
│  │ Plugin   │  │   AI     │  │  WASM    │      │
│  │ Service  │  │ Service  │  │ Service  │      │
│  └──────────┘  └──────────┘  └──────────┘      │
└─────────────────────────────────────────────────┘
                      ↕
┌─────────────────────────────────────────────────┐
│               数据层 (Data Layer)                │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐      │
│  │ SQLite   │  │  File    │  │  Cache   │      │
│  │ Database │  │  System  │  │  Store   │      │
│  └──────────┘  └──────────┘  └──────────┘      │
└─────────────────────────────────────────────────┘
```

### 插件生成流程

```
用户输入需求
    ↓
┌────────────────┐
│ 前端 UI        │
│ - 输入验证     │
│ - 提交请求     │
└────────────────┘
    ↓
┌────────────────┐
│ App Controller │
│ - 接收请求     │
│ - 调用服务     │
└────────────────┘
    ↓
┌────────────────┐
│ AI Service     │
│ - 加载 Prompt  │
│ - 调用 AI API  │
│ - 生成代码     │
└────────────────┘
    ↓
┌────────────────┐
│ WASM Service   │
│ - Go 代码编译  │
│ - 生成 WASM    │
└────────────────┘
    ↓
┌────────────────┐
│ Plugin Service │
│ - 保存到数据库 │
│ - 触发事件     │
└────────────────┘
    ↓
┌────────────────┐
│ 前端 UI        │
│ - 显示新插件   │
│ - 可立即使用   │
└────────────────┘
```

### 插件执行流程

```
用户点击插件
    ↓
┌────────────────┐
│ 前端 UI        │
│ - 显示输入框   │
│ - 收集参数     │
└────────────────┘
    ↓
┌────────────────┐
│ App Controller │
│ - ExecutePlugin│
└────────────────┘
    ↓
┌────────────────┐
│ Plugin Service │
│ - 查询插件     │
│ - 验证状态     │
└────────────────┘
    ↓
┌────────────────┐
│ WASM Service   │
│ - 加载 WASM    │
│ - 执行代码     │
│ - 返回结果     │
└────────────────┘
    ↓
┌────────────────┐
│ 前端 UI        │
│ - 显示结果     │
│ - 格式化输出   │
└────────────────┘
```

---

## 核心模块设计

### 1. 前端层 (Frontend Layer)

#### 1.1 UI 组件

**职责：**
- 渲染用户界面
- 处理用户交互
- 显示数据和状态

**主要组件：**
- **工具栏** (`toolbar`): 新建、刷新等操作
- **输入区域** (`input-section`): AI 需求输入
- **插件列表** (`plugins-list`): 显示所有插件
- **执行区域** (`execute-section`): 插件参数输入和结果显示

**技术细节：**
```javascript
// 组件状态管理
const state = {
    plugins: [],          // 插件列表
    currentPlugin: null,  // 当前选中的插件
    isGenerating: false   // 是否正在生成
};

// 事件处理
elements.addToolBtn.addEventListener('click', showInputSection);
elements.generateBtn.addEventListener('click', handleGenerate);
```

#### 1.2 状态管理

**职责：**
- 管理应用状态
- 同步前后端数据
- 处理状态变更

**状态流：**
```
User Action → Event Handler → State Update → UI Rerender
```

#### 1.3 事件系统

**职责：**
- 监听后端事件
- 实时更新 UI
- 实现插件热更新

**实现：**
```javascript
// 监听插件添加事件
EventsOn('plugin:added', (plugin) => {
    state.plugins.push(plugin);
    renderPlugins();
});

// 监听插件删除事件
EventsOn('plugin:deleted', (pluginId) => {
    state.plugins = state.plugins.filter(p => p.id !== pluginId);
    renderPlugins();
});
```

---

### 2. 应用层 (App Layer)

#### 2.1 App Controller

**职责：**
- 导出 API 给前端
- 协调各个服务
- 处理业务逻辑
- 管理应用生命周期

**结构：**
```go
type App struct {
    ctx            context.Context
    db             *sql.DB
    pluginService  *services.PluginService
    aiService      *services.AIService
    wasmService    *services.WasmService
}
```

**主要方法：**
- `GetAllPlugins()`: 获取插件列表
- `GeneratePlugin(requirement)`: 生成新插件
- `ExecutePlugin(id, input)`: 执行插件
- `DeletePlugin(id)`: 删除插件

**设计原则：**
- **单一职责**: 每个方法只做一件事
- **依赖注入**: 服务通过构造函数注入
- **错误处理**: 统一返回错误给前端
- **事件驱动**: 通过事件通知前端更新

---

### 3. 服务层 (Service Layer)

#### 3.1 Plugin Service

**职责：**
- 插件 CRUD 操作
- 数据库交互
- 插件验证

**接口设计：**
```go
type PluginService interface {
    GetAllPlugins() ([]models.Plugin, error)
    GetPluginByID(id int64) (*models.Plugin, error)
    CreatePlugin(plugin *models.Plugin) error
    UpdatePlugin(plugin *models.Plugin) error
    DeletePlugin(id int64) error
}
```

**实现细节：**
- 使用预编译语句防止 SQL 注入
- 实现事务支持
- 添加数据验证
- 记录操作日志

#### 3.2 AI Service

**职责：**
- 调用 AI API
- 管理 Prompt 模板
- 生成插件代码
- 提取代码内容

**工作流：**
```go
1. 加载 Prompt 模板
2. 替换模板变量 ({{REQUIREMENT}})
3. 调用 AI API (GPT-4)
4. 解析响应内容
5. 提取代码块
6. 返回 Go 代码
```

**容错机制：**
- API 密钥未配置时使用内置模板
- 网络请求超时重试
- 响应解析失败时返回默认代码
- 记录生成日志供调试

#### 3.3 WASM Service

**职责：**
- Go 代码编译为 WASM
- WASM 模块执行
- 管理临时文件
- 资源清理

**编译流程：**
```go
1. 创建临时工作目录
2. 写入 Go 源代码
3. 初始化 Go 模块 (go mod init)
4. 设置环境变量 (GOOS=js GOARCH=wasm)
5. 执行编译 (go build)
6. 读取 WASM 文件
7. 清理临时文件
8. 返回二进制数据
```

**执行流程：**
```go
1. 保存 WASM 到临时文件
2. 生成 Node.js 执行脚本
3. 实例化 WebAssembly 模块
4. 调用导出的 execute 函数
5. 传递 JSON 输入
6. 接收 JSON 输出
7. 清理临时文件
8. 返回结果
```

**安全考虑：**
- WASM 沙箱隔离执行
- 限制执行时间
- 限制内存使用
- 验证输入输出格式

---

### 4. 数据层 (Data Layer)

#### 4.1 数据库设计

**表结构：**

```sql
CREATE TABLE plugins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    code TEXT NOT NULL,
    wasm_binary BLOB,
    is_active INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_plugins_name ON plugins(name);
CREATE INDEX idx_plugins_is_active ON plugins(is_active);
```

**字段说明：**
- `id`: 主键，自增
- `name`: 插件名称
- `description`: 插件描述（用户输入的需求）
- `code`: Go 源代码
- `wasm_binary`: 编译后的 WASM 二进制
- `is_active`: 是否启用
- `created_at`: 创建时间
- `updated_at`: 更新时间

#### 4.2 数据模型

```go
type Plugin struct {
    ID          int64     `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Code        string    `json:"code"`
    WasmBinary  []byte    `json:"-"`      // 不序列化
    IsActive    bool      `json:"isActive"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}
```

#### 4.3 数据访问

**Repository 模式：**
```go
type PluginRepository interface {
    FindAll() ([]Plugin, error)
    FindByID(id int64) (*Plugin, error)
    Save(plugin *Plugin) error
    Update(plugin *Plugin) error
    Delete(id int64) error
}
```

---

## 设计模式应用

### 1. 依赖注入 (Dependency Injection)

**目的：** 解耦模块，便于测试和维护

**实现：**
```go
// 构造函数注入
func NewApp(db *sql.DB) *App {
    return &App{
        db:            db,
        pluginService: services.NewPluginService(db),
        aiService:     services.NewAIService(),
        wasmService:   services.NewWasmService(),
    }
}
```

### 2. 服务层模式 (Service Layer)

**目的：** 封装业务逻辑，提供统一接口

**实现：**
```go
type PluginService struct {
    db *sql.DB
}

func (s *PluginService) CreatePlugin(plugin *models.Plugin) error {
    // 业务逻辑：验证、保存、通知
}
```

### 3. 仓储模式 (Repository Pattern)

**目的：** 抽象数据访问，隔离数据库细节

**实现：**
```go
type PluginRepository struct {
    db *sql.DB
}

func (r *PluginRepository) Save(plugin *Plugin) error {
    // SQL 操作
}
```

### 4. 观察者模式 (Observer Pattern)

**目的：** 实现事件驱动，前后端解耦

**实现：**
```go
// 后端发送事件
runtime.EventsEmit(ctx, "plugin:added", plugin)

// 前端监听事件
EventsOn('plugin:added', (plugin) => {
    // 处理事件
});
```

### 5. 策略模式 (Strategy Pattern)

**目的：** 灵活切换 AI 服务实现

**实现：**
```go
type AIService interface {
    GenerateCode(requirement string) (string, error)
}

type OpenAIService struct {}
type MockAIService struct {}

// 根据配置选择实现
```

---

## 高内聚低耦合实现

### 高内聚

**模块职责清晰：**
- **PluginService**: 只负责插件数据管理
- **AIService**: 只负责 AI 代码生成
- **WasmService**: 只负责 WASM 编译执行

**单一职责原则：**
```go
// ✅ 好的设计
func (s *PluginService) GetPluginByID(id int64) (*Plugin, error)
func (s *PluginService) CreatePlugin(plugin *Plugin) error

// ❌ 不好的设计
func (s *PluginService) GetAndExecutePlugin(id int64) (string, error)
```

### 低耦合

**接口隔离：**
```go
// 服务之间通过接口交互
type PluginRepository interface {
    Save(*Plugin) error
}

// 不直接依赖具体实现
type App struct {
    pluginRepo PluginRepository  // 接口，而非具体类型
}
```

**事件驱动：**
```go
// 不直接调用前端方法，而是发送事件
runtime.EventsEmit(ctx, "plugin:added", plugin)
```

**配置外部化：**
```go
// 使用环境变量而非硬编码
apiKey := os.Getenv("OPENAI_API_KEY")
```

---

## 性能优化策略

### 1. 前端优化

- **虚拟滚动**: 大量插件时使用虚拟列表
- **防抖节流**: 搜索输入使用防抖
- **懒加载**: 按需加载 WASM 模块
- **缓存**: 缓存插件列表和执行结果

### 2. 后端优化

- **连接池**: 复用数据库连接
- **索引优化**: 为常查询字段添加索引
- **批量操作**: 合并多个数据库操作
- **异步执行**: 耗时操作异步处理

### 3. WASM 优化

- **编译缓存**: 缓存已编译的 WASM
- **并发执行**: 支持多个插件并发运行
- **资源限制**: 限制内存和 CPU 使用
- **预加载**: 预先编译常用插件

---

## 安全设计

### 1. 代码执行安全

- ✅ WASM 沙箱隔离
- ✅ 禁止访问文件系统
- ✅ 禁止网络请求
- ✅ 限制执行时间

### 2. 输入验证

- ✅ JSON 格式验证
- ✅ SQL 注入防护
- ✅ XSS 攻击防护
- ✅ 参数长度限制

### 3. 数据安全

- ✅ 敏感信息加密
- ✅ API 密钥环境变量管理
- ✅ 本地数据库权限控制
- ✅ 定期数据备份

---

## 可扩展性设计

### 1. 插件类型扩展

当前只支持计算类插件，未来可扩展：
- UI 组件插件
- 网络请求插件
- 文件处理插件
- 系统集成插件

### 2. AI 模型扩展

当前使用 GPT-4，未来可支持：
- Claude
- Gemini
- 本地模型
- 自定义模型

### 3. 执行引擎扩展

当前使用 Go WASM，未来可支持：
- JavaScript
- Python (Pyodide)
- Rust WASM
- 原生插件

---

## 部署架构

### 单机版

```
┌──────────────────┐
│   loji App       │
│  ┌────────────┐  │
│  │  Frontend  │  │
│  └────────────┘  │
│  ┌────────────┐  │
│  │  Backend   │  │
│  └────────────┘  │
│  ┌────────────┐  │
│  │  SQLite    │  │
│  └────────────┘  │
└──────────────────┘
```

### 未来：云同步版

```
┌──────────────┐     ┌──────────────┐
│  Client App  │────▶│  Cloud API   │
└──────────────┘     └──────────────┘
                            │
                     ┌──────┴──────┐
                     │             │
              ┌──────▼───┐  ┌──────▼───┐
              │ Database │  │  Storage │
              └──────────┘  └──────────┘
```

---

## 总结

loji App 采用现代化的架构设计，实现了：

1. ✅ **高内聚**: 模块职责清晰，单一职责
2. ✅ **低耦合**: 接口隔离，依赖注入，事件驱动
3. ✅ **可扩展**: 插件化架构，易于添加新功能
4. ✅ **高性能**: WASM 执行，数据库优化
5. ✅ **安全性**: 沙箱隔离，输入验证
6. ✅ **可维护**: 代码规范，文档完善

这种架构为未来的功能扩展和性能优化奠定了坚实的基础。

