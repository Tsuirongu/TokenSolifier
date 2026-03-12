# Wails 自动生成的绑定文件

这个目录包含 Wails 自动生成的前端绑定文件。

## 自动生成

当您运行以下命令时，这些文件会自动生成：

```bash
wails dev
# 或
wails build
```

## 目录结构

生成后的结构如下：

```
wailsjs/
├── go/
│   └── app/
│       └── App.js          # App 结构体的方法绑定
├── runtime/
│   └── runtime.js          # Wails 运行时 API
└── README.md               # 本文件
```

## 使用方法

### 调用后端方法

```javascript
import { GetAllPlugins, GeneratePlugin } from './wailsjs/go/app/App.js';

// 获取所有插件
const plugins = await GetAllPlugins();

// 生成新插件
const plugin = await GeneratePlugin("体脂率计算器");
```

### 使用运行时 API

```javascript
import { EventsOn, EventsEmit } from './wailsjs/runtime/runtime.js';

// 监听事件
EventsOn('plugin:added', (plugin) => {
    console.log('New plugin:', plugin);
});

// 发送事件
EventsEmit('custom:event', { data: 'value' });
```

## 可用的后端方法

以下方法会自动生成对应的 JavaScript 绑定：

### App.js

- `GetAllPlugins()` - 获取所有插件
- `GeneratePlugin(requirement)` - 生成新插件
- `ExecutePlugin(pluginID, input)` - 执行插件
- `DeletePlugin(pluginID)` - 删除插件
- `TogglePluginStatus(pluginID)` - 切换插件状态
- `GetPluginInfo(pluginID)` - 获取插件信息

## 运行时 API

### 事件系统

```javascript
// 监听事件
EventsOn(eventName, callback)

// 监听一次
EventsOnce(eventName, callback)

// 取消监听
EventsOff(eventName)

// 发送事件
EventsEmit(eventName, ...data)
```

### 窗口控制

```javascript
// 最小化窗口
WindowMinimise()

// 最大化窗口
WindowMaximise()

// 关闭窗口
Quit()

// 隐藏窗口
WindowHide()

// 显示窗口
WindowShow()
```

### 日志

```javascript
// 输出日志
LogPrint(message)

// 调试日志
LogDebug(message)

// 信息日志
LogInfo(message)

// 警告日志
LogWarning(message)

// 错误日志
LogError(message)
```

## 注意事项

1. **不要手动编辑**：这些文件由 Wails 自动生成，手动修改会在下次生成时丢失

2. **版本控制**：建议将此目录添加到 `.gitignore`，因为它可以自动生成

3. **类型提示**：如果使用 TypeScript，可以运行 `wails generate module` 生成类型定义

4. **重新生成**：如果后端方法有变化，运行 `wails dev` 或 `wails build` 会自动更新绑定

## 手动生成

如果需要手动生成绑定（不启动应用）：

```bash
wails generate module
```

## TypeScript 支持

生成 TypeScript 类型定义：

```bash
wails generate module -ts
```

这会创建 `.d.ts` 文件，提供完整的类型支持。

## 故障排除

### 绑定未更新

如果修改了后端代码但绑定没有更新：

```bash
# 清理并重新构建
wails build -clean

# 或手动生成
wails generate module
```

### 导入错误

确保导入路径正确：

```javascript
// ✅ 正确
import { GetAllPlugins } from '../wailsjs/go/app/App.js';

// ❌ 错误
import { GetAllPlugins } from './go/app/App.js';
```

### 方法不存在

确保：
1. 后端方法是 `App` 结构体的公开方法
2. 方法名首字母大写
3. 已经重新编译应用

## 更多信息

- [Wails 官方文档](https://wails.io/docs/)
- [运行时 API 文档](https://wails.io/docs/reference/runtime/)
- [前端开发指南](https://wails.io/docs/guides/frontend/)

