# loji App API 文档

## 概述

loji App 使用 Wails 框架提供前后端通信，所有后端方法都可以从前端直接调用。

## 后端API

### App 结构

所有API方法都绑定在 `App` 结构体上。

### 方法列表

#### 1. GetAllPlugins

获取所有插件列表。

**签名：**
```go
func (a *App) GetAllPlugins() ([]models.Plugin, error)
```

**返回：**
- `[]models.Plugin`: 插件列表
- `error`: 错误信息

**前端调用：**
```javascript
import { GetAllPlugins } from '../wailsjs/go/app/App.js';

const plugins = await GetAllPlugins();
```

---

#### 2. GeneratePlugin

通过AI生成新的插件。

**签名：**
```go
func (a *App) GeneratePlugin(requirement string) (*models.Plugin, error)
```

**参数：**
- `requirement`: 插件需求描述（字符串）

**返回：**
- `*models.Plugin`: 生成的插件对象
- `error`: 错误信息

**前端调用：**
```javascript
import { GeneratePlugin } from '../wailsjs/go/app/App.js';

const plugin = await GeneratePlugin("体脂率计算器");
```

---

#### 3. ExecutePlugin

执行指定的插件。

**签名：**
```go
func (a *App) ExecutePlugin(pluginID int64, input string) (string, error)
```

**参数：**
- `pluginID`: 插件ID
- `input`: 输入参数（JSON字符串）

**返回：**
- `string`: 执行结果（JSON字符串）
- `error`: 错误信息

**前端调用：**
```javascript
import { ExecutePlugin } from '../wailsjs/go/app/App.js';

const input = JSON.stringify({
    weight: 70,
    height: 175,
    age: 25,
    gender: "男"
});

const result = await ExecutePlugin(1, input);
const output = JSON.parse(result);
```

---

#### 4. DeletePlugin

删除指定的插件。

**签名：**
```go
func (a *App) DeletePlugin(pluginID int64) error
```

**参数：**
- `pluginID`: 插件ID

**返回：**
- `error`: 错误信息

**前端调用：**
```javascript
import { DeletePlugin } from '../wailsjs/go/app/App.js';

await DeletePlugin(1);
```

---

#### 5. TogglePluginStatus

切换插件的启用/禁用状态。

**签名：**
```go
func (a *App) TogglePluginStatus(pluginID int64) error
```

**参数：**
- `pluginID`: 插件ID

**返回：**
- `error`: 错误信息

**前端调用：**
```javascript
import { TogglePluginStatus } from '../wailsjs/go/app/App.js';

await TogglePluginStatus(1);
```

---

#### 6. GetPluginInfo

获取插件的详细信息。

**签名：**
```go
func (a *App) GetPluginInfo(pluginID int64) (string, error)
```

**参数：**
- `pluginID`: 插件ID

**返回：**
- `string`: 插件信息（JSON字符串）
- `error`: 错误信息

**前端调用：**
```javascript
import { GetPluginInfo } from '../wailsjs/go/app/App.js';

const infoJson = await GetPluginInfo(1);
const info = JSON.parse(infoJson);
```

---

## 事件系统

### 前端监听事件

使用 Wails Runtime 的 `EventsOn` 方法监听后端事件。

#### plugin:added

当新插件被添加时触发。

**事件数据：**
- `Plugin` 对象

**监听示例：**
```javascript
import { EventsOn } from '../wailsjs/runtime/runtime.js';

EventsOn('plugin:added', (plugin) => {
    console.log('New plugin added:', plugin);
    // 刷新插件列表
});
```

---

#### plugin:deleted

当插件被删除时触发。

**事件数据：**
- `pluginID` (number)

**监听示例：**
```javascript
import { EventsOn } from '../wailsjs/runtime/runtime.js';

EventsOn('plugin:deleted', (pluginId) => {
    console.log('Plugin deleted:', pluginId);
    // 刷新插件列表
});
```

---

## 数据模型

### Plugin

插件模型定义。

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

**字段说明：**
- `ID`: 插件唯一标识
- `Name`: 插件名称
- `Description`: 插件描述
- `Code`: Go源代码
- `WasmBinary`: 编译后的WASM二进制（不在JSON中传输）
- `IsActive`: 是否启用
- `CreatedAt`: 创建时间
- `UpdatedAt`: 更新时间

---

## 错误处理

所有API方法在发生错误时都会返回错误信息。前端应该适当处理这些错误。

**示例：**
```javascript
try {
    const plugin = await GeneratePlugin("体脂率计算器");
    console.log('Success:', plugin);
} catch (error) {
    console.error('Error:', error);
    // 显示错误提示
}
```

---

## 插件开发规范

### 插件代码结构

所有插件都必须遵循以下结构：

```go
package main

import (
    "encoding/json"
    "fmt"
)

// 输入数据结构
type Input struct {
    // 定义输入字段
}

// 输出数据结构
type Output struct {
    // 定义输出字段
}

// 必须有main函数（可以为空）
func main() {}

// 导出的执行函数
//export execute
func execute(input string) string {
    var data Input
    if err := json.Unmarshal([]byte(input), &data); err != nil {
        return fmt.Sprintf("{\"error\": \"%s\"}", err.Error())
    }

    // 实现业务逻辑
    
    result := Output{
        // 填充结果
    }
    
    jsonResult, _ := json.Marshal(result)
    return string(jsonResult)
}
```

### 输入输出规范

- **输入**: 必须是有效的JSON字符串
- **输出**: 必须是有效的JSON字符串
- **错误**: 返回包含 `error` 字段的JSON对象

### 示例：体脂率计算器

**输入：**
```json
{
    "weight": 70,
    "height": 175,
    "age": 25,
    "gender": "男"
}
```

**输出：**
```json
{
    "bmi": 22.86,
    "bodyFat": 15.2,
    "category": "正常"
}
```

**错误：**
```json
{
    "error": "Invalid input format"
}
```

---

## 最佳实践

### 1. 错误处理

始终使用 try-catch 包裹异步调用：

```javascript
try {
    const result = await ExecutePlugin(pluginId, input);
    // 处理成功结果
} catch (error) {
    // 处理错误
    showToast('执行失败: ' + error, 'error');
}
```

### 2. 数据验证

在发送请求前验证数据：

```javascript
const input = document.getElementById('input').value;
try {
    JSON.parse(input); // 验证JSON格式
} catch (e) {
    showToast('输入必须是有效的JSON', 'error');
    return;
}
```

### 3. 加载状态

为耗时操作显示加载状态：

```javascript
setLoading(true);
try {
    await GeneratePlugin(requirement);
} finally {
    setLoading(false);
}
```

### 4. 事件监听

监听后端事件以实现实时更新：

```javascript
EventsOn('plugin:added', () => {
    loadPlugins(); // 刷新列表
});
```

---

## 调试技巧

### 查看日志

后端日志会输出到控制台：

```bash
# 开发模式
wails dev

# 日志会显示所有API调用和错误信息
```

### 前端调试

使用浏览器开发者工具：

```javascript
console.log('Plugin data:', plugin);
console.error('Error:', error);
```

### 数据库查看

数据库位置：`~/.loji-app/loji.db`

```bash
# 使用sqlite3查看
sqlite3 ~/.loji-app/loji.db

# 查询所有插件
SELECT * FROM plugins;
```

---

## 安全考虑

1. **代码执行隔离**: 插件在WASM沙箱中执行，与主程序隔离
2. **输入验证**: 所有输入都应该验证格式
3. **错误处理**: 避免暴露敏感信息
4. **API密钥**: 使用环境变量管理API密钥

---

## 性能优化

1. **缓存插件列表**: 避免频繁查询数据库
2. **异步执行**: 使用异步API避免阻塞UI
3. **懒加载**: 按需加载WASM模块
4. **批量操作**: 合并多个数据库操作

---

## 扩展开发

### 添加新的API方法

1. 在 `backend/app/app.go` 中添加方法
2. 方法必须是 `App` 结构体的公开方法
3. 重新构建后，Wails 会自动生成前端绑定

### 自定义AI Prompt

编辑 `prompts/plugin_generator.txt` 文件来自定义代码生成行为。

### 添加新的服务

在 `backend/services/` 目录下创建新的服务文件。

