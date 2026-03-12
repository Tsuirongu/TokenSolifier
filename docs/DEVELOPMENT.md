# loji App 开发指南

## 目录

- [开发环境设置](#开发环境设置)
- [项目结构](#项目结构)
- [开发流程](#开发流程)
- [代码规范](#代码规范)
- [测试](#测试)
- [调试](#调试)
- [部署](#部署)

---

## 开发环境设置

### 前置要求

- **Go**: 1.21 或更高版本
- **Node.js**: 16.x 或更高版本
- **Wails CLI**: v2.8.0 或更高版本
- **操作系统**: macOS, Windows, Linux

### macOS 依赖

```bash
# 安装 Xcode Command Line Tools
xcode-select --install
```

### Windows 依赖

- 安装 Visual Studio 2019 或更高版本
- 安装 WebView2 运行时

### Linux 依赖

```bash
# Ubuntu/Debian
sudo apt-get install build-essential libgtk-3-dev libwebkit2gtk-4.0-dev

# Fedora
sudo dnf install gtk3-devel webkit2gtk3-devel

# Arch
sudo pacman -S gtk3 webkit2gtk
```

### 安装 Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 初始化项目

```bash
# 克隆项目
git clone <repository-url>
cd loji-app

# 初始化开发环境
make init

# 或者手动安装
cd frontend && npm install
cd .. && go mod download
```

---

## 项目结构

```
loji-app/
├── backend/                      # 后端代码
│   ├── app/                      # 应用层
│   │   └── app.go               # 主应用逻辑，导出给前端的API
│   ├── models/                   # 数据模型
│   │   └── plugin.go            # 插件数据模型
│   └── services/                 # 服务层
│       ├── database.go          # 数据库初始化和表创建
│       ├── plugin_service.go    # 插件CRUD操作
│       ├── ai_service.go        # AI代码生成服务
│       └── wasm_service.go      # WASM编译和执行
│
├── frontend/                     # 前端代码
│   ├── src/                     # 源代码
│   │   ├── index.html          # 主HTML文件
│   │   ├── style.css           # 样式文件
│   │   └── app.js              # 应用逻辑
│   ├── dist/                    # 构建产物（自动生成）
│   ├── wailsjs/                 # Wails生成的绑定（自动生成）
│   ├── package.json             # 前端依赖
│   └── vite.config.js          # Vite配置
│
├── prompts/                      # AI Prompt模板
│   └── plugin_generator.txt     # 插件代码生成Prompt
│
├── docs/                         # 文档
│   ├── API.md                   # API文档
│   └── DEVELOPMENT.md           # 开发指南（本文件）
│
├── main.go                       # 程序入口
├── go.mod                        # Go依赖管理
├── wails.json                    # Wails配置
├── Makefile                      # 构建脚本
├── build.sh                      # 构建脚本（Bash）
├── dev.sh                        # 开发脚本（Bash）
├── .env.example                 # 环境变量示例
├── .gitignore                   # Git忽略文件
└── README.md                    # 项目说明
```

---

## 开发流程

### 1. 启动开发服务器

```bash
# 使用 Makefile
make dev

# 或直接使用脚本
./dev.sh

# 或使用 Wails CLI
wails dev
```

开发服务器特性：
- ✅ 热重载
- ✅ 自动编译
- ✅ 实时错误提示
- ✅ 前端开发工具

### 2. 修改代码

#### 后端开发

1. 在 `backend/` 目录下修改 Go 代码
2. 保存后自动重新编译
3. 查看终端输出的错误信息

**示例：添加新的API方法**

```go
// backend/app/app.go
func (a *App) GetPluginStats() (map[string]int, error) {
    plugins, err := a.pluginService.GetAllPlugins()
    if err != nil {
        return nil, err
    }
    
    stats := map[string]int{
        "total": len(plugins),
        "active": 0,
    }
    
    for _, p := range plugins {
        if p.IsActive {
            stats["active"]++
        }
    }
    
    return stats, nil
}
```

#### 前端开发

1. 在 `frontend/src/` 目录下修改文件
2. 保存后浏览器自动刷新
3. 使用浏览器开发者工具调试

**示例：调用新的API**

```javascript
// frontend/src/app.js
import { GetPluginStats } from '../wailsjs/go/app/App.js';

async function loadStats() {
    try {
        const stats = await GetPluginStats();
        console.log('Plugin stats:', stats);
    } catch (error) {
        console.error('Failed to load stats:', error);
    }
}
```

### 3. 构建生产版本

```bash
# 使用 Makefile
make build

# 或直接使用脚本
./build.sh

# 或使用 Wails CLI
wails build
```

构建产物位置：
- **macOS**: `build/bin/loji-app.app`
- **Windows**: `build/bin/loji-app.exe`
- **Linux**: `build/bin/loji-app`

---

## 代码规范

### Go 代码规范

遵循 [Effective Go](https://golang.org/doc/effective_go) 和 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)。

**命名规范：**
- 包名：小写，简短，无下划线
- 函数名：驼峰命名，导出函数首字母大写
- 变量名：驼峰命名
- 常量名：驼峰命名或全大写+下划线

**示例：**
```go
package services

// 常量
const DefaultTimeout = 30

// 结构体
type PluginService struct {
    db *sql.DB
}

// 导出函数
func NewPluginService(db *sql.DB) *PluginService {
    return &PluginService{db: db}
}

// 私有函数
func (s *PluginService) validatePlugin(plugin *models.Plugin) error {
    // ...
}
```

### JavaScript 代码规范

**命名规范：**
- 变量名：驼峰命名
- 常量名：全大写+下划线
- 函数名：驼峰命名
- 类名：帕斯卡命名

**示例：**
```javascript
// 常量
const MAX_PLUGINS = 100;

// 变量
let currentPlugin = null;

// 函数
async function loadPlugins() {
    // ...
}

// 类
class PluginManager {
    constructor() {
        // ...
    }
}
```

### CSS 规范

使用 BEM 命名方式：

```css
/* Block */
.plugin-item {
    /* ... */
}

/* Element */
.plugin-item__name {
    /* ... */
}

/* Modifier */
.plugin-item--active {
    /* ... */
}
```

### 注释规范

**Go:**
```go
// GetAllPlugins 获取所有插件
// 返回按创建时间降序排列的插件列表
func (s *PluginService) GetAllPlugins() ([]models.Plugin, error) {
    // 实现...
}
```

**JavaScript:**
```javascript
/**
 * 加载所有插件
 * @returns {Promise<void>}
 */
async function loadPlugins() {
    // 实现...
}
```

---

## 测试

### 单元测试

**Go 测试：**

```go
// backend/services/plugin_service_test.go
package services

import "testing"

func TestPluginService_GetAllPlugins(t *testing.T) {
    // 测试实现
}
```

运行测试：
```bash
go test ./backend/...
```

**JavaScript 测试：**

```javascript
// frontend/src/app.test.js
import { describe, it, expect } from 'vitest';

describe('App', () => {
    it('should load plugins', () => {
        // 测试实现
    });
});
```

### 集成测试

创建完整的测试场景：

```go
func TestPluginWorkflow(t *testing.T) {
    // 1. 创建插件
    // 2. 执行插件
    // 3. 删除插件
}
```

---

## 调试

### 后端调试

**使用日志：**
```go
import "log"

log.Printf("Plugin ID: %d", pluginID)
log.Println("Executing plugin...")
```

**使用调试器：**
```bash
# 使用 Delve
dlv debug
```

### 前端调试

**使用 Console：**
```javascript
console.log('Plugin data:', plugin);
console.error('Error:', error);
console.table(plugins);
```

**使用开发者工具：**
1. 打开浏览器开发者工具（F12）
2. 查看 Console、Network、Sources 等标签
3. 设置断点进行调试

### 数据库调试

**查看数据库内容：**
```bash
sqlite3 ~/.loji-app/loji.db

# SQL命令
.tables
SELECT * FROM plugins;
.schema plugins
```

---

## 常见问题

### 1. Wails dev 启动失败

**问题：** 端口被占用

**解决：**
```bash
# 查找占用端口的进程
lsof -i :34115

# 杀掉进程
kill -9 <PID>
```

### 2. 前端无法调用后端API

**问题：** 绑定未生成

**解决：**
```bash
# 重新生成绑定
wails generate module
```

### 3. WASM 编译失败

**问题：** Go 版本不兼容

**解决：**
```bash
# 检查 Go 版本
go version

# 更新到 1.21+
```

### 4. 样式不生效

**问题：** 缓存问题

**解决：**
```bash
# 清理缓存
rm -rf frontend/dist
make build
```

---

## 部署

### macOS

```bash
# 构建
wails build -platform darwin

# 产物
build/bin/loji-app.app

# 签名（可选）
codesign --deep --force --sign "Developer ID" loji-app.app
```

### Windows

```bash
# 构建
wails build -platform windows

# 产物
build/bin/loji-app.exe

# 创建安装程序（可选）
# 使用 NSIS 或 WiX
```

### Linux

```bash
# 构建
wails build -platform linux

# 产物
build/bin/loji-app

# 创建 AppImage（可选）
# 使用 appimagetool
```

---

## 性能优化

### 1. 前端优化

- 使用虚拟列表渲染大量插件
- 懒加载图片和资源
- 压缩 CSS 和 JS
- 使用 Web Workers 处理耗时任务

### 2. 后端优化

- 使用缓存减少数据库查询
- 批量操作数据库
- 异步执行耗时任务
- 使用连接池

### 3. 数据库优化

- 添加索引
- 定期清理无用数据
- 优化查询语句
- 使用事务

---

## 扩展开发

### 添加新服务

1. 在 `backend/services/` 创建新文件
2. 定义服务结构和方法
3. 在 `app.go` 中注入服务
4. 导出API方法给前端

### 添加新页面

1. 在 `frontend/src/` 创建新HTML文件
2. 添加对应的CSS和JS
3. 在 `vite.config.js` 中配置入口
4. 更新路由

### 自定义插件类型

1. 修改 `models/plugin.go` 添加新字段
2. 更新数据库 schema
3. 修改服务层逻辑
4. 更新前端界面

---

## 贡献指南

### 提交代码

1. Fork 项目
2. 创建特性分支：`git checkout -b feature/new-feature`
3. 提交更改：`git commit -am 'Add new feature'`
4. 推送分支：`git push origin feature/new-feature`
5. 创建 Pull Request

### 代码审查

- 确保代码符合规范
- 添加必要的测试
- 更新相关文档
- 描述清楚更改内容

---

## 资源链接

- [Wails 官方文档](https://wails.io/docs/)
- [Go 官方文档](https://golang.org/doc/)
- [MDN Web Docs](https://developer.mozilla.org/)
- [SQLite 文档](https://www.sqlite.org/docs.html)
- [WebAssembly 规范](https://webassembly.github.io/spec/)

---

## 许可证

MIT License - 详见 LICENSE 文件

