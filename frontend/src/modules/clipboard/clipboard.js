// 剪贴板模块 - 处理剪贴板管理和操作
import {
    GetClipboardItems, UpdateClipboardItem, DeleteClipboardItem, ClearClipboardHistory,
    SearchClipboardItems, CopyToClipboard, ToggleClipboardFavorite
} from '../../../wailsjs/go/app/App.js';
import {
    state, elements, showToast, escapeHtml, showConfirm, formatTime, formatSize
} from '../common/common.js';

// 剪贴板模块初始化
export async function initClipboard() {
    // 加载剪贴板HTML内容
    await loadClipboardContent();

    // 绑定剪贴板事件
    bindClipboardEvents();
}

// 加载剪贴板HTML内容
async function loadClipboardContent() {
    try {
        // 使用import导入HTML模板
        const clipboardTemplate = await import('./clipboard.html?raw');
        const html = clipboardTemplate.default;

        // 将HTML内容插入到剪贴板窗口中
        const clipboardContent = elements.clipboardWindow?.querySelector('.window-content');
        if (clipboardContent) {
            clipboardContent.innerHTML = html;
        }

        // 重新初始化元素引用
        initClipboardElements();
    } catch (error) {
        console.error('加载剪贴板内容失败:', error);
        showToast('加载剪贴板失败', 'error');
    }
}

// 初始化剪贴板元素引用
function initClipboardElements() {
    elements.clipboardSearchInput = document.getElementById('clipboard-search-input');
    elements.searchClearBtn = document.getElementById('search-clear-btn');
    elements.clearAllBtn = document.getElementById('clear-all-btn');
    elements.clipboardList = document.getElementById('clipboard-list');
    elements.clipboardCount = document.getElementById('clipboard-count');
}

// 绑定剪贴板事件
function bindClipboardEvents() {
    if (elements.clearAllBtn) {
        elements.clearAllBtn.addEventListener('click', handleClearAll);
    }

    if (elements.clipboardSearchInput) {
        elements.clipboardSearchInput.addEventListener('input', handleSearch);
    }

    if (elements.searchClearBtn) {
        elements.searchClearBtn.addEventListener('click', clearSearch);
    }
}

// 加载剪贴板项
export async function loadClipboardItems() {
    try {
        const items = await GetClipboardItems(100, 0);
        state.clipboardItems = items || [];
        renderClipboardItems();
    } catch (error) {
        console.error('Failed to load clipboard items:', error);
        showToast('加载剪贴板失败', 'error');
    }
}

// 渲染剪贴板项
function renderClipboardItems() {
    if (!elements.clipboardCount || !elements.clipboardList) return;

    const count = state.clipboardItems.length;
    elements.clipboardCount.textContent = count;

    if (count === 0) {
        if (elements.clipboardList) {
            elements.clipboardList.innerHTML = `
                <div class="empty-state">
                    <div class="empty-icon">📋</div>
                    <p>还没有剪贴板历史</p>
                    <p class="empty-hint">复制的内容会自动记录在这里</p>
                </div>
            `;
        }
        return;
    }

    if (elements.clipboardList) {
        elements.clipboardList.innerHTML = state.clipboardItems.map(item => {
            const timeStr = formatTime(item.created_at);
            const sizeStr = formatSize(item.size);
            const isURL = isValidURL(item.content || item.preview);

            if (item.type === 'text') {
                return `
                    <div class="clipboard-item" data-id="${item.id}">
                        <div class="item-header">
                            <span class="item-type">📝 ${isURL ? '链接' : '文本'}</span>
                            <div class="item-actions">
                                <button class="item-btn fav ${item.is_fav ? 'active' : ''}" data-action="fav" title="收藏">
                                    ${item.is_fav ? '⭐' : '☆'}
                                </button>
                                ${isURL ? '<button class="item-btn item-preview-btn" data-action="preview" title="预览">👁️</button>' : ''}
                                <button class="item-btn" data-action="copy" title="复制">📋</button>
                                <button class="item-btn" data-action="edit" title="编辑">✏️</button>
                                <button class="item-btn delete" data-action="delete" title="删除">🗑️</button>
                            </div>
                        </div>
                        <div class="item-content" data-action="copy-content">
                            <pre class="item-text">${escapeHtml(item.preview)}</pre>
                        </div>
                        <div class="item-footer">
                            <span class="item-time">${timeStr}</span>
                            <span class="item-size">${sizeStr}</span>
                        </div>
                    </div>
                `;
            } else if (item.type === 'image') {
                return `
                    <div class="clipboard-item" data-id="${item.id}">
                        <div class="item-header">
                            <span class="item-type">🖼️ 图片</span>
                            <div class="item-actions">
                                <button class="item-btn fav ${item.is_fav ? 'active' : ''}" data-action="fav" title="收藏">
                                    ${item.is_fav ? '⭐' : '☆'}
                                </button>
                                <button class="item-btn item-preview-btn" data-action="preview" title="预览">👁️</button>
                                <button class="item-btn" data-action="copy" title="复制">📋</button>
                                <button class="item-btn delete" data-action="delete" title="删除">🗑️</button>
                            </div>
                        </div>
                        <div class="item-content image-content" data-action="copy-content">
                            <img src="data:image/png;base64,${item.content}" alt="剪贴板图片" class="item-image" />
                        </div>
                        <div class="item-footer">
                            <span class="item-time">${timeStr}</span>
                            <span class="item-size">${sizeStr}</span>
                        </div>
                    </div>
                `;
            }
        }).join('');

        // 绑定事件
        bindClipboardItemEvents();
    }
}

// 绑定剪贴板项事件
function bindClipboardItemEvents() {
    if (!elements.clipboardList) return;

    document.querySelectorAll('.clipboard-item').forEach(itemEl => {
        const id = parseInt(itemEl.dataset.id);
        const item = state.clipboardItems.find(i => i.id === id);

        if (!item) return;

        // 收藏按钮
        const favBtn = itemEl.querySelector('[data-action="fav"]');
        if (favBtn) {
            favBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                handleToggleFavorite(id);
            });
        }

        // 预览按钮
        const previewBtn = itemEl.querySelector('[data-action="preview"]');
        if (previewBtn) {
            previewBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                if (window.previewModule && window.previewModule.showPreview) {
                    window.previewModule.showPreview(item);
                }
            });
        }

        // 复制按钮
        const copyBtn = itemEl.querySelector('[data-action="copy"]');
        if (copyBtn) {
            copyBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                handleCopyItem(id);
            });
        }

        // 编辑按钮（仅文本）
        const editBtn = itemEl.querySelector('[data-action="edit"]');
        if (editBtn) {
            editBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                if (window.editModule && window.editModule.showEditModal) {
                    window.editModule.showEditModal(item);
                }
            });
        }

        // 删除按钮
        const deleteBtn = itemEl.querySelector('[data-action="delete"]');
        if (deleteBtn) {
            deleteBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                handleDeleteItem(id);
            });
        }

        // 点击内容区域复制
        const contentEl = itemEl.querySelector('[data-action="copy-content"]');
        if (contentEl) {
            contentEl.addEventListener('click', () => handleCopyItem(id));
        }
    });
}

// 处理收藏切换
async function handleToggleFavorite(id) {
    try {
        await ToggleClipboardFavorite(id);
        showToast('已更新', 'success');
    } catch (error) {
        console.error('Failed to toggle favorite:', error);
        showToast('操作失败', 'error');
    }
}

// 处理复制项
async function handleCopyItem(id) {
    try {
        await CopyToClipboard(id);
        showToast('已复制到剪贴板', 'success');
    } catch (error) {
        console.error('Failed to copy item:', error);
        showToast('复制失败', 'error');
    }
}

// 处理删除项
async function handleDeleteItem(id) {
    try {
        await DeleteClipboardItem(id);
        showToast('已删除', 'success');
        await loadClipboardItems(); // 重新加载列表
    } catch (error) {
        console.error('Failed to delete item:', error);
        showToast('删除失败', 'error');
    }
}

// 处理清空全部
async function handleClearAll() {
    const confirmed = await showConfirm('确定要清空所有剪贴板历史吗？此操作不可恢复。');
    if (!confirmed) {
        return;
    }

    try {
        await ClearClipboardHistory();
        showToast('已清空', 'success');
        await loadClipboardItems(); // 重新加载列表
    } catch (error) {
        console.error('Failed to clear clipboard:', error);
        showToast('清空失败', 'error');
    }
}

// 处理搜索
async function handleSearch(e) {
    const keyword = e.target.value.trim();
    state.searchKeyword = keyword;

    if (keyword) {
        if (elements.searchClearBtn) {
            elements.searchClearBtn.style.display = 'block';
        }
        try {
            const items = await SearchClipboardItems(keyword, 100);
            state.clipboardItems = items || [];
            renderClipboardItems();
        } catch (error) {
            console.error('Failed to search clipboard:', error);
        }
    } else {
        if (elements.searchClearBtn) {
            elements.searchClearBtn.style.display = 'none';
        }
        await loadClipboardItems();
    }
}

// 清空搜索
async function clearSearch() {
    if (elements.clipboardSearchInput) {
        elements.clipboardSearchInput.value = '';
    }
    if (elements.searchClearBtn) {
        elements.searchClearBtn.style.display = 'none';
    }
    state.searchKeyword = '';
    await loadClipboardItems();
}

// 检测是否为有效URL
function isValidURL(string) {
    if (!string) return false;
    try {
        const url = new URL(string.trim());
        return url.protocol === 'http:' || url.protocol === 'https:';
    } catch (_) {
        return false;
    }
}

