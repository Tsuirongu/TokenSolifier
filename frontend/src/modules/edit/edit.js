// 编辑模块 - 处理编辑功能
import { UpdateClipboardItem } from '../../../wailsjs/go/app/App.js';
import { state, elements, showToast } from '../common/common.js';

// 编辑模块初始化
export function initEdit() {
    // 加载编辑HTML内容
    loadEditContent();

    // 绑定编辑事件
    bindEditEvents();
}

// 加载编辑HTML内容
async function loadEditContent() {
    try {
        // 使用import导入HTML模板
        const editTemplate = await import('./edit.html?raw');
        const html = editTemplate.default;

        // 创建编辑模态框容器
        const editContainer = document.createElement('div');
        editContainer.innerHTML = html;
        const editModal = editContainer.firstElementChild;

        // 将编辑模态框添加到body中
        document.body.appendChild(editModal);

        // 重新初始化元素引用
        initEditElements();
    } catch (error) {
        console.error('加载编辑内容失败:', error);
    }
}

// 初始化编辑元素引用
function initEditElements() {
    elements.editModal = document.getElementById('edit-modal');
    elements.editTextarea = document.getElementById('edit-textarea');
    elements.editSaveBtn = document.getElementById('edit-save-btn');
    elements.editCancelBtn = document.getElementById('edit-cancel-btn');
    elements.modalCloseBtn = document.getElementById('modal-close-btn');
}

// 绑定编辑事件
function bindEditEvents() {
    if (elements.editSaveBtn) {
        elements.editSaveBtn.addEventListener('click', handleSaveEdit);
    }
    if (elements.editCancelBtn) {
        elements.editCancelBtn.addEventListener('click', hideEditModal);
    }
    if (elements.modalCloseBtn) {
        elements.modalCloseBtn.addEventListener('click', hideEditModal);
    }
}

// 显示编辑对话框
export function showEditModal(item) {
    if (item.type !== 'text') {
        showToast('只能编辑文本内容', 'info');
        return;
    }

    state.currentEditItem = item;
    if (elements.editTextarea) {
        elements.editTextarea.value = item.content;
    }
    if (elements.editModal) {
        elements.editModal.style.display = 'flex';
    }
}

// 隐藏编辑对话框
function hideEditModal() {
    if (elements.editModal) {
        elements.editModal.style.display = 'none';
    }
    state.currentEditItem = null;
    if (elements.editTextarea) {
        elements.editTextarea.value = '';
    }
}

// 处理保存编辑
async function handleSaveEdit() {
    if (!state.currentEditItem) return;

    if (!elements.editTextarea) return;

    const newContent = elements.editTextarea.value;

    if (!newContent.trim()) {
        showToast('内容不能为空', 'error');
        return;
    }

    try {
        const updatedItem = {
            ...state.currentEditItem,
            content: newContent,
            preview: newContent.length > 100 ? newContent.substring(0, 100) + '...' : newContent,
            size: newContent.length
        };

        await UpdateClipboardItem(updatedItem);
        showToast('保存成功', 'success');
        hideEditModal();

        // 刷新剪贴板列表
        if (window.loadClipboardItems) {
            window.loadClipboardItems();
        }
    } catch (error) {
        console.error('Failed to save edit:', error);
        showToast('保存失败', 'error');
    }
}
