// 公共工具函数和状态管理
import { EventsOn } from '../../../wailsjs/runtime/runtime.js';

// 全局状态管理
export const state = {
    plugins: [],
    tags: [],
    currentPlugin: null,
    isGenerating: false,
    isExpanded: false,
    clipboardItems: [],
    currentEditItem: null,
    searchKeyword: '',
    openExecutors: new Map(),
    selectedTagIds: [] // 选中的标签ID数组
};

// DOM元素引用管理
export const elements = {
    // 基础元素
    floatingBar: null,
    mainButton: null,
    quickAiChatBtn: null,
    quickConfigBtn: null,
    quickClipboardBtn: null,

    floatingWindow: null,
    closeWindowBtn: null,
    minimizeBtn: null,

    // 工具箱相关
    addToolBtn: null,
    refreshBtn: null,
    toolCreator: null,
    requirementInput: null,
    closeCreatorBtn: null,
    cancelCreatorBtn: null,
    generateBtn: null,
    toolsList: null,
    toolsCount: null,

    // 工具箱详情
    detailTitle: null,
    detailDescription: null,
    detailInput: null,
    detailOutput: null,
    detailExecuteBtn: null,
    detailDeleteBtn: null,

    // 剪贴板相关
    clipboardWindow: null,
    closeClipboardBtn: null,
    clipboardMinimizeBtn: null,
    clipboardList: null,
    clipboardCount: null,
    clipboardSearchInput: null,
    searchClearBtn: null,
    clearAllBtn: null,

    // 预览相关
    previewWindow: null,
    closePreviewBtn: null,
    previewIcon: null,
    previewTitleText: null,
    previewContent: null,
    previewLoading: null,
    previewBody: null,

    // 编辑相关
    editModal: null,
    editTextarea: null,
    editSaveBtn: null,
    editCancelBtn: null,
    modalCloseBtn: null,

    // 确认对话框
    confirmModal: null,
    confirmMessage: null,
    confirmOkBtn: null,
    confirmCancelBtn: null,

    // 进度条
    progressOverlay: null,
    progressBar: null,
    progressFill: null,
    progressPercentage: null,
    progressStage: null,
    progressMessage: null,

    // AI对话相关
    aiChatModal: null,
    aiChatCloseBtn: null,
    newSessionBtn: null,
    chatMessages: null,
    chatInput: null,
    chatSendBtn: null,

    // 通知提示
    toast: null
};

// 初始化元素引用
export function initElements() {
    elements.floatingBar = document.getElementById('floating-bar');
    elements.mainButton = document.getElementById('main-button');
    elements.quickAiChatBtn = document.getElementById('quick-ai-chat-btn');
    elements.quickConfigBtn = document.getElementById('quick-config-btn');
    elements.quickClipboardBtn = document.getElementById('quick-clipboard-btn');

    elements.floatingWindow = document.getElementById('floating-window');
    elements.closeWindowBtn = document.getElementById('close-window-btn');
    elements.minimizeBtn = document.getElementById('minimize-btn');

    elements.addToolBtn = document.getElementById('add-tool-btn');
    elements.refreshBtn = document.getElementById('refresh-btn');
    elements.toolCreator = document.getElementById('tool-creator');
    elements.requirementInput = document.getElementById('requirement-input');
    elements.closeCreatorBtn = document.getElementById('close-creator-btn');
    elements.cancelCreatorBtn = document.getElementById('cancel-creator-btn');
    elements.generateBtn = document.getElementById('generate-btn');
    elements.toolsList = document.getElementById('tools-list');
    elements.toolsCount = document.getElementById('tools-count');

    elements.detailTitle = document.getElementById('detail-title');
    elements.detailDescription = document.getElementById('detail-description');
    elements.detailInput = document.getElementById('detail-input');
    elements.detailOutput = document.getElementById('detail-output');
    elements.detailExecuteBtn = document.getElementById('detail-execute-btn');
    elements.detailDeleteBtn = document.getElementById('detail-delete-btn');

    elements.clipboardWindow = document.getElementById('clipboard-window');
    elements.closeClipboardBtn = document.getElementById('close-clipboard-btn');
    elements.clipboardMinimizeBtn = document.getElementById('clipboard-minimize-btn');
    elements.clipboardList = document.getElementById('clipboard-list');
    elements.clipboardCount = document.getElementById('clipboard-count');
    elements.clipboardSearchInput = document.getElementById('clipboard-search-input');
    elements.searchClearBtn = document.getElementById('search-clear-btn');
    elements.clearAllBtn = document.getElementById('clear-all-btn');

    elements.previewWindow = document.getElementById('preview-window');
    elements.closePreviewBtn = document.getElementById('close-preview-btn');
    elements.previewIcon = document.getElementById('preview-icon');
    elements.previewTitleText = document.getElementById('preview-title-text');
    elements.previewContent = document.getElementById('preview-content');
    elements.previewLoading = document.getElementById('preview-loading');
    elements.previewBody = document.getElementById('preview-body');

    elements.editModal = document.getElementById('edit-modal');
    elements.editTextarea = document.getElementById('edit-textarea');
    elements.editSaveBtn = document.getElementById('edit-save-btn');
    elements.editCancelBtn = document.getElementById('edit-cancel-btn');
    elements.modalCloseBtn = document.getElementById('modal-close-btn');

    elements.confirmModal = document.getElementById('confirm-modal');
    elements.confirmMessage = document.getElementById('confirm-message');
    elements.confirmOkBtn = document.getElementById('confirm-ok-btn');
    elements.confirmCancelBtn = document.getElementById('confirm-cancel-btn');

    elements.progressOverlay = document.getElementById('progress-overlay');
    elements.progressBar = document.getElementById('progress-bar');
    elements.progressFill = document.getElementById('progress-fill');
    elements.progressPercentage = document.getElementById('progress-percentage');
    elements.progressStage = document.getElementById('progress-stage');
    elements.progressMessage = document.getElementById('progress-message');

    // AI对话相关
    elements.aiChatModal = document.getElementById('ai-chat-modal');
    elements.aiChatCloseBtn = document.getElementById('ai-chat-close-btn');
    elements.newSessionBtn = document.getElementById('new-session-btn');
    elements.chatMessages = document.getElementById('chat-messages');
    elements.chatInput = document.getElementById('chat-input');
    elements.chatSendBtn = document.getElementById('chat-send-btn');

    elements.toast = document.getElementById('toast');
}

// 公共工具函数
export function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

export function setButtonLoading(button, loading) {
    if (!button) return;

    const spinner = button.querySelector('.loading-spinner');
    const text = button.querySelector('.btn-text');

    if (loading) {
        if (spinner) spinner.style.display = 'inline-block';
        if (text) text.textContent = '处理中...';
        button.disabled = true;
    } else {
        if (spinner) spinner.style.display = 'none';
        if (button === elements.generateBtn && text) {
            text.textContent = '✨ 生成工具';
        } else if (text) {
            text.textContent = '🚀 执行';
        }
        button.disabled = false;
    }
}

export function showToast(message, type = 'info') {
    if (!elements.toast) return;

    elements.toast.textContent = message;
    elements.toast.className = `toast ${type} show`;

    setTimeout(() => {
        if (elements.toast) {
            elements.toast.classList.remove('show');
        }
    }, 3000);
}

export function showConfirm(message) {
    return new Promise((resolve) => {
        if (!elements.confirmModal || !elements.confirmMessage ||
            !elements.confirmOkBtn || !elements.confirmCancelBtn) {
            resolve(false);
            return;
        }

        elements.confirmMessage.textContent = message;
        elements.confirmModal.style.display = 'flex';

        // 创建新的事件处理函数
        const handleOk = () => {
            elements.confirmModal.style.display = 'none';
            cleanup();
            resolve(true);
        };

        const handleCancel = () => {
            elements.confirmModal.style.display = 'none';
            cleanup();
            resolve(false);
        };

        const cleanup = () => {
            if (elements.confirmOkBtn) {
                elements.confirmOkBtn.removeEventListener('click', handleOk);
            }
            if (elements.confirmCancelBtn) {
                elements.confirmCancelBtn.removeEventListener('click', handleCancel);
            }
        };

        // 绑定事件
        elements.confirmOkBtn.addEventListener('click', handleOk);
        elements.confirmCancelBtn.addEventListener('click', handleCancel);
    });
}

// 监听插件事件
export function listenToPluginEvents() {
    EventsOn('plugin:added', () => {
        // 触发插件加载事件，由工具箱模块处理
        if (window.loadPlugins) {
            window.loadPlugins();
        }
    });
    EventsOn('plugin:deleted', () => {
        if (window.loadPlugins) {
            window.loadPlugins();
        }
    });
}

// 监听剪贴板事件
export function listenToClipboardEvents() {
    EventsOn('clipboard:new', (item) => {
        console.log('New clipboard item:', item);
        showToast('新的剪贴板内容', 'info');
        if (window.loadClipboardItems) {
            window.loadClipboardItems();
        }
    });

    EventsOn('clipboard:updated', (item) => {
        if (window.loadClipboardItems) {
            window.loadClipboardItems();
        }
    });

    EventsOn('clipboard:deleted', (id) => {
        if (window.loadClipboardItems) {
            window.loadClipboardItems();
        }
    });

    EventsOn('clipboard:cleared', () => {
        if (window.loadClipboardItems) {
            window.loadClipboardItems();
        }
    });
}

// 格式化时间
export function formatTime(timeStr) {
    const date = new Date(timeStr);
    const now = new Date();
    const diffMs = now - date;
    const diffMins = Math.floor(diffMs / 60000);

    if (diffMins < 1) return '刚刚';
    if (diffMins < 60) return `${diffMins}分钟前`;

    const diffHours = Math.floor(diffMins / 60);
    if (diffHours < 24) return `${diffHours}小时前`;

    const diffDays = Math.floor(diffHours / 24);
    if (diffDays < 7) return `${diffDays}天前`;

    // 超过7天显示日期
    return `${date.getMonth() + 1}/${date.getDate()}`;
}

// 格式化大小
export function formatSize(bytes) {
    if (bytes < 1024) return `${bytes}B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)}KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)}MB`;
}

// 进度条控制函数
export function showProgress() {
    if (elements.progressOverlay) {
        elements.progressOverlay.style.display = 'flex';
        state.isGenerating = true;
        updateProgress(0, '准备中...', '正在初始化...');
    }
}

export function hideProgress() {
    if (elements.progressOverlay) {
        elements.progressOverlay.style.display = 'none';
        state.isGenerating = false;
    }
}

export function updateProgress(progress, stage, message) {
    if (!elements.progressFill || !elements.progressPercentage ||
        !elements.progressStage || !elements.progressMessage) {
        return;
    }

    // 更新进度条
    elements.progressFill.style.width = `${progress}%`;

    // 更新百分比
    elements.progressPercentage.textContent = `${progress}%`;

    // 更新阶段和消息
    if (stage) {
        elements.progressStage.textContent = stage;
    }
    if (message) {
        elements.progressMessage.textContent = message;
    }
}

// 监听进度事件
export function listenToProgressEvents() {
    console.log('Setting up progress event listeners...');

    EventsOn('plugin:generation:progress', (data) => {
        console.log('Progress event received:', data);
        const { stage, progress, message } = data;
        updateProgress(progress, stage, message);
    });

    EventsOn('plugin:generation:start', () => {
        console.log('Progress start event received');
        showProgress();
    });

    EventsOn('plugin:generation:end', () => {
        console.log('Progress end event received');
        hideProgress();
    });

    console.log('Progress event listeners set up successfully');
}

// 测试进度条功能（仅用于调试）
export function testProgressBar() {
    console.log('Testing progress bar...');
    showProgress();

    setTimeout(() => updateProgress(10, '配置预处理', '正在获取插件生成配置...'), 500);
    setTimeout(() => updateProgress(30, '功能解析', '正在分析用户需求...'), 1000);
    setTimeout(() => updateProgress(50, '输入输出格式化', '正在生成输入输出描述...'), 1500);
    setTimeout(() => updateProgress(80, '代码生成', '正在生成插件代码...'), 2000);
    setTimeout(() => updateProgress(100, '创建完毕', '正在保存插件...'), 2500);
    setTimeout(() => hideProgress(), 3000);
}

// 在全局暴露测试函数（仅用于调试）
if (typeof window !== 'undefined') {
    window.testProgressBar = testProgressBar;
}
