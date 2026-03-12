# 智能剪贴板功能文档

## 功能概述

智能剪贴板是 loji App 的核心功能之一，用于自动记录和管理用户的复制历史，支持文本和图片。该功能取代了原入口栏的第三个按钮（设置按钮）。

## 主要特性

### 1. 自动记录
- ✅ 自动监控系统剪贴板变化
- ✅ 实时记录复制的文本和图片
- ✅ 智能去重，避免重复记录相同内容

### 2. 历史管理
- 📋 查看所有剪贴板历史
- 🔍 搜索历史内容
- ⭐ 收藏重要内容
- 🗑️ 删除不需要的记录
- 🧹 一键清空所有历史

### 3. 内容操作
- 📝 编辑文本内容
- 📋 一键复制到系统剪贴板
- 🖼️ 支持图片预览和复制
- 📊 显示内容大小和创建时间

### 4. 用户体验
- 🎨 美观简洁的界面设计
- 🖱️ 支持拖动面板
- ⚡ 流畅的动画效果
- 📱 响应式设计

## 技术架构

### 后端 (Go)

#### 数据模型
- **ClipboardItem**: 剪贴板项模型
  - ID: 唯一标识
  - Type: 类型（text/image）
  - Content: 内容（文本或 base64 图片）
  - Preview: 预览文本
  - Size: 内容大小
  - IsFav: 是否收藏
  - CreatedAt/UpdatedAt: 时间戳

#### 核心服务
- **ClipboardService**: 剪贴板服务
  - 系统剪贴板监控（500ms 轮询）
  - 数据库 CRUD 操作
  - 内容搜索
  - 去重逻辑

#### API 方法
```go
GetClipboardItems(limit, offset int) - 获取历史列表
GetClipboardItem(id int64) - 获取单个项
UpdateClipboardItem(item) - 更新项
DeleteClipboardItem(id int64) - 删除项
ClearClipboardHistory() - 清空历史
SearchClipboardItems(keyword, limit) - 搜索
CopyToClipboard(id int64) - 复制到系统剪贴板
ToggleClipboardFavorite(id int64) - 切换收藏状态
```

#### 事件系统
- `clipboard:new` - 新剪贴板内容
- `clipboard:updated` - 内容更新
- `clipboard:deleted` - 内容删除
- `clipboard:cleared` - 历史清空

### 前端 (JavaScript/HTML/CSS)

#### 组件结构
- **剪贴板面板** (`clipboard-panel`)
  - 头部：标题和关闭按钮
  - 搜索栏：实时搜索
  - 工具栏：清空按钮和计数器
  - 内容列表：历史记录展示

- **编辑对话框** (`clipboard-edit-modal`)
  - 文本编辑器
  - 保存/取消按钮

#### 核心功能
```javascript
toggleClipboardPanel() - 切换面板显示
loadClipboardItems() - 加载历史
renderClipboardItems() - 渲染列表
handleSearch() - 搜索处理
handleCopyItem() - 复制处理
handleDeleteItem() - 删除处理
handleToggleFavorite() - 收藏切换
showEditModal() - 显示编辑框
```

### 数据库表结构

```sql
CREATE TABLE clipboard_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,           -- 类型：text/image
    content TEXT NOT NULL,        -- 内容
    preview TEXT NOT NULL,        -- 预览文本
    size INTEGER NOT NULL,        -- 大小（字节）
    is_fav INTEGER DEFAULT 0,     -- 是否收藏
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

-- 索引
CREATE INDEX idx_clipboard_created_at ON clipboard_items(created_at DESC);
CREATE INDEX idx_clipboard_is_fav ON clipboard_items(is_fav);
```

## 使用说明

### 打开剪贴板面板
点击悬浮栏右侧第三个按钮（📋 图标）即可打开智能剪贴板面板。

### 查看历史
面板会自动显示所有复制历史，按时间倒序排列。

### 搜索内容
在搜索框中输入关键词，实时搜索匹配的历史记录。

### 复制历史内容
- 点击内容区域直接复制
- 或点击右上角的复制按钮（📋）

### 编辑文本
点击编辑按钮（✏️），在弹出的对话框中编辑文本内容。

### 收藏/取消收藏
点击星标按钮（☆/⭐）切换收藏状态。

### 删除记录
点击删除按钮（🗑️）删除单条记录。

### 清空历史
点击左下角的"清空"按钮可清空所有历史记录（需确认）。

## 可扩展性设计

### 1. 代码组织
- **高内聚**：剪贴板功能模块化，独立于其他功能
- **低耦合**：通过事件系统与其他模块通信
- **清晰接口**：前后端 API 定义明确

### 2. 扩展方向

#### 后续可作为工具箱的输入输出入口
```javascript
// 示例：从剪贴板获取输入
function getClipboardInput() {
    return state.clipboardItems[0]?.content;
}

// 示例：输出结果到剪贴板
async function outputToClipboard(content) {
    const item = createClipboardItem(content);
    await SaveClipboardItem(item);
}
```

#### 可能的扩展功能
- 📌 固定常用内容
- 🏷️ 内容标签分类
- 🔗 内容关联链接
- 📝 富文本支持
- 🔄 同步到云端
- 🤖 AI 内容分析
- 🔐 敏感内容加密

### 3. 插件集成
剪贴板可作为插件系统的标准输入输出接口：
- 插件从剪贴板读取输入
- 插件结果自动保存到剪贴板
- 支持内容格式转换

## 性能优化

### 1. 数据库
- 使用索引加速查询
- 限制单次查询数量（默认 100 条）
- 定期清理过期数据

### 2. 前端
- 虚拟滚动（未来实现）
- 图片懒加载
- 防抖搜索
- 事件委托

### 3. 监控
- 500ms 轮询间隔（平衡性能和实时性）
- 智能去重减少数据库操作
- 异步处理避免阻塞

## 注意事项

### 隐私安全
- 所有数据存储在本地（`~/.loji-app/loji.db`）
- 不会上传到服务器
- 建议不要复制敏感信息

### 存储管理
- 定期清理不需要的历史
- 大量图片可能占用较多空间
- 可通过清空功能一键清理

### 系统兼容性
- macOS: 完全支持
- Windows: 需要测试
- Linux: 需要测试

## 开发调试

### 启动开发模式
```bash
wails dev
```

### 查看数据库
```bash
sqlite3 ~/.loji-app/loji.db
```

### 查看日志
剪贴板相关日志会输出到控制台，包括：
- 监控启动/停止
- 新内容捕获
- 错误信息

## 未来改进

- [ ] 支持更多内容类型（文件、链接等）
- [ ] 内容自动分类
- [ ] 快捷键支持
- [ ] 导出/导入历史
- [ ] 内容统计分析
- [ ] AI 智能推荐

---

**版本**: v1.0.0  
**更新时间**: 2025-10-09  
**作者**: loji App Team

