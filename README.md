# TokenSolifier
A project driven by LLM that solidifies tasks originally requiring burning tokens through LLM into a fast script with no token consumption.

一个轻量级、高性能、跨平台的桌面应用程序，可以通过AI自动生成并运行各种小工具。
AI驱动，一句话生成脚本。自然语言一句话生成自动生成Skill，后续调用节约Token消耗。

## 功能特性

- 🎯 **悬浮窗设计**：非侵入式的桌面悬浮窗，随时可用
- 🤖 **AI代码生成**：输入需求，AI自动生成工具代码
- 🔥 **插件热更新**：动态加载新工具，无需重启应用
- ⚡ **WASM执行**：使用WebAssembly执行插件，安全高效
- 💾 **本地存储**：SQLite数据库本地存储所有插件
- 🎨 **美观界面**：现代化无边框设计，简洁易用

## 技术栈

- **前端**：HTML + CSS + JavaScript (原生)
- **后端**：Golang
- **框架**：Wails v2
- **数据库**：SQLite
- **执行层**：Go WebAssembly

## 项目结构

```
loji-app/
├── backend/                 # 后端代码
│   ├── app/                # 应用层
│   │   └── app.go          # 主应用逻辑
│   ├── models/             # 数据模型
│   │   └── plugin.go       # 插件模型
│   └── services/           # 服务层
│       ├── database.go     # 数据库服务
│       ├── plugin_service.go # 插件管理
│       ├── ai_service.go   # AI代码生成
│       └── wasm_service.go # WASM编译执行
├── frontend/               # 前端代码
│   ├── dist/              # 构建产物
│   ├── src/               # 源代码
│   │   ├── index.html     # 主页面
│   │   ├── style.css      # 样式
│   │   └── app.js         # 应用逻辑
│   └── package.json       # 前端依赖
├── prompts/               # AI Prompt模板
│   └── plugin_generator.txt
├── main.go                # 入口文件
├── go.mod                 # Go依赖
└── wails.json             # Wails配置
```

## 快速开始

### 前置要求

- Go 1.21+
- Node.js 16+
- Wails CLI v2

### 安装依赖

```bash
# 安装Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 安装Go依赖
go mod download

# 安装前端依赖
cd frontend && npm install
```

### 开发模式

```bash
wails dev
```

### 构建应用

```bash
wails build
```

## 使用说明

1. **启动应用**：运行编译后的应用程序
2. **输入需求**：点击"+"按钮，在文本框中输入您想要的工具需求
3. **生成工具**：AI会自动生成工具代码并添加到悬浮窗
4. **使用工具**：点击工具按钮，输入参数，查看结果

## 配置

### AI服务配置

设置环境变量来配置AI服务：

```bash
export OPENAI_API_KEY="your-api-key"
export OPENAI_API_URL="https://api.openai.com/v1/chat/completions"
```

如果未配置API密钥，系统会使用内置的模板生成代码。

### 自定义Prompt

编辑 `prompts/plugin_generator.txt` 文件来自定义AI生成代码的行为。

## 架构设计

### 高内聚低耦合原则

- **服务层**：独立的服务模块，各司其职
- **数据层**：统一的数据访问接口
- **业务层**：清晰的业务逻辑分层
- **前端层**：组件化设计，模块复用

### 插件系统

插件采用WASM技术，具有以下特点：

- **安全性**：沙箱环境执行，隔离主程序
- **性能**：接近原生性能
- **可移植**：跨平台兼容
- **动态加载**：运行时加载和卸载

## 示例插件

### 体脂率计算器

```json
输入: {
  "weight": 70,
  "height": 175,
  "age": 25,
  "gender": "male"
}

输出: {
  "bmi": 22.86,
  "bodyFat": 15.2,
  "category": "正常"
}
```

## 开发路线图

- [x] 基础架构搭建
- [x] AI代码生成服务
- [x] WASM编译和执行
- [x] 插件管理系统
- [ ] 前端界面实现
- [ ] 插件市场
- [ ] 多语言支持
- [ ] 云端同步

## 贡献指南

欢迎提交Issue和Pull Request！

## 许可证

MIT License
