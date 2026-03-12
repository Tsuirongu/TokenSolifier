// 主应用模块 - 处理应用初始化和基础功能
import { ExpandWindow, CollapseWindow } from '../../../wailsjs/go/app/App.js';
import { state, elements, initElements, showToast, listenToPluginEvents, listenToClipboardEvents, listenToProgressEvents } from '../common/common.js';

// 模块加载管理
let toolboxModule = null;
let clipboardModule = null;
let previewModule = null;
let editModule = null;

// 聊天会话状态
let currentSessionId = null;

// 初始化应用
async function initApp() {
    initElements();
    bindEvents();
    await loadModules();
    listenToPluginEvents();
    listenToClipboardEvents();
    listenToProgressEvents();
}

// 动态加载模块
async function loadModules() {
    try {
        // 加载工具箱模块
        const toolboxModuleScript = await import('../toolbox/toolbox.js');
        toolboxModule = toolboxModuleScript;

        // 加载剪贴板模块
        const clipboardModuleScript = await import('../clipboard/clipboard.js');
        clipboardModule = clipboardModuleScript;

        // 加载预览模块
        const previewModuleScript = await import('../preview/preview.js');
        previewModule = previewModuleScript;

        // 加载编辑模块
        const editModuleScript = await import('../edit/edit.js');
        editModule = editModuleScript;

        // 初始化模块
        if (toolboxModule.initToolbox) {
            await toolboxModule.initToolbox();
        }
        if (clipboardModule.initClipboard) {
            await clipboardModule.initClipboard();
        }
        if (previewModule.initPreview) {
            await previewModule.initPreview();
        }
        if (editModule.initEdit) {
            await editModule.initEdit();
        }

    } catch (error) {
        console.error('模块加载失败:', error);
        showToast('模块加载失败', 'error');
    }
}

// 绑定基础事件
function bindEvents() {
    // 主按钮点击展开窗口
    if (elements.mainButton) {
        elements.mainButton.addEventListener('click', expandApp);
    }

    // 快捷工具按钮
    if (elements.quickAiChatBtn) {
        elements.quickAiChatBtn.addEventListener('click', async (e) => {
            e.stopPropagation();
            await openAiChatDialog();
        });
    }

    if (elements.quickConfigBtn) {
        elements.quickConfigBtn.addEventListener('click', async (e) => {
            e.stopPropagation();
            await openConfigWindow();
        });
    }

    if (elements.quickClipboardBtn) {
        elements.quickClipboardBtn.addEventListener('click', async (e) => {
            e.stopPropagation();
            await openClipboardWindow();
        });
    }

    // 窗口控制
    if (elements.closeWindowBtn) {
        elements.closeWindowBtn.addEventListener('click', collapseApp);
    }

    if (elements.minimizeBtn) {
        elements.minimizeBtn.addEventListener('click', () => {
            import('../../../wailsjs/runtime/runtime.js').then(({ WindowMinimise }) => {
                WindowMinimise();
            });
        });
    }

    // 剪贴板窗口控制
    if (elements.closeClipboardBtn) {
        elements.closeClipboardBtn.addEventListener('click', closeClipboardWindow);
    }

    if (elements.clipboardMinimizeBtn) {
        elements.clipboardMinimizeBtn.addEventListener('click', () => {
            import('../../../wailsjs/runtime/runtime.js').then(({ WindowMinimise }) => {
                WindowMinimise();
            });
        });
    }

    // AI对话框事件
    if (elements.aiChatCloseBtn) {
        elements.aiChatCloseBtn.addEventListener('click', closeAiChatDialog);
    }

    if (elements.newSessionBtn) {
        elements.newSessionBtn.addEventListener('click', createNewChatSession);
    }

    if (elements.chatSendBtn) {
        elements.chatSendBtn.addEventListener('click', sendAiMessage);
    }

    if (elements.chatInput) {
        elements.chatInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                sendAiMessage();
            }
        });
    }
}

// 扩展应用显示完整界面
async function expandApp() {
    if (state.isExpanded) return;

    try {
        await ExpandWindow(800, 600); // 工具箱模式：800x600
        document.body.classList.add('expanded');
        if (elements.floatingWindow) {
            elements.floatingWindow.style.display = 'block';
        }
        if (elements.floatingBar) {
            elements.floatingBar.style.display = 'none';
        }
        state.isExpanded = true;
    } catch (error) {
        showToast('无法扩展窗口', 'error');
    }
}

// 收缩应用为长条
async function collapseApp() {
    try {
        if (elements.floatingWindow) {
            elements.floatingWindow.style.display = 'none';
        }
        if (elements.floatingBar) {
            elements.floatingBar.style.display = 'flex';
        }
        document.body.classList.remove('expanded');

        // 隐藏所有模块内容
        if (toolboxModule && toolboxModule.hideCreator) {
            toolboxModule.hideCreator();
        }

        await CollapseWindow();
        state.isExpanded = false;
    } catch (error) {
        showToast('无法收缩窗口', 'error');
    }
}

// 打开剪贴板窗口
async function openClipboardWindow() {
    try {
        await ExpandWindow(400, 600); // 剪贴板模式：400x600，更窄的宽度
        document.body.classList.add('expanded');
        if (elements.clipboardWindow) {
            elements.clipboardWindow.style.display = 'block';
        }
        if (elements.floatingWindow) {
            elements.floatingWindow.style.display = 'none';
        }
        if (elements.floatingBar) {
            elements.floatingBar.style.display = 'none';
        }
        state.isExpanded = true;

        if (clipboardModule && clipboardModule.loadClipboardItems) {
            await clipboardModule.loadClipboardItems();
        }
    } catch (error) {
        console.error('Failed to open clipboard window:', error);
        showToast('无法打开剪贴板窗口', 'error');
    }
}

// 关闭剪贴板窗口
async function closeClipboardWindow() {
    try {
        if (elements.clipboardWindow) {
            elements.clipboardWindow.style.display = 'none';
        }
        if (elements.floatingBar) {
            elements.floatingBar.style.display = 'flex';
        }
        document.body.classList.remove('expanded');
        await CollapseWindow();
        state.isExpanded = false;
    } catch (error) {
        console.error('Failed to close clipboard window:', error);
        showToast('无法关闭剪贴板窗口', 'error');
    }
}

// 打开配置窗口
async function openConfigWindow() {
    try {
        await ExpandWindow(500, 600); // 配置窗口：500x600，更窄的宽度
        document.body.classList.add('expanded');

        // 创建配置窗口HTML
        const configWindowHTML = `
            <div class="config-window" id="config-window">
                <div class="window-wrapper">
                    <!-- 窗口头部 -->
                    <div class="window-header" data-wails-drag>
                        <div class="window-title">
                            <span class="title-icon">⚙️</span>
                            <span class="title-text">全局配置</span>
                        </div>
                        <div class="window-controls" data-wails-no-drag>
                            <button class="control-btn minimize" id="config-minimize-btn" title="最小化">─</button>
                            <button class="control-btn close" id="close-config-btn" title="关闭">✕</button>
                        </div>
                    </div>

                    <!-- 窗口内容 -->
                    <div class="window-content" data-wails-no-drag>
                        <div class="config-container">
                            <div class="config-form">
                                <div class="form-group">
                                    <label for="openai-api-url">OpenAI API地址</label>
                                    <input type="url" id="openai-api-url" class="form-input"
                                           placeholder="https://api.openai.com/v1/chat/completions">
                                </div>

                                <div class="form-group">
                                    <label for="openai-model-type">OpenAI模型类型</label>
                                    <input type="text" id="openai-model-type" class="form-input"
                                           placeholder="例如：gpt-4o, gpt-4-turbo, gpt-3.5-turbo">
                                </div>

                                <div class="form-group">
                                    <label for="openai-api-key">OpenAI API密钥 *</label>
                                    <input type="password" id="openai-api-key" class="form-input"
                                           placeholder="sk-...">
                                </div>

                                <div class="form-actions">
                                    <button class="btn-secondary" id="reset-config-btn">重置</button>
                                    <button class="btn-primary" id="save-config-btn">保存配置</button>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;

        // 添加到body中
        const existingWindow = document.getElementById('config-window');
        if (existingWindow) {
            existingWindow.remove();
        }

        document.body.insertAdjacentHTML('beforeend', configWindowHTML);

        // 隐藏其他窗口
        if (elements.floatingWindow) {
            elements.floatingWindow.style.display = 'none';
        }
        if (elements.clipboardWindow) {
            elements.clipboardWindow.style.display = 'none';
        }
        if (elements.floatingBar) {
            elements.floatingBar.style.display = 'none';
        }

        // 显示配置窗口
        const configWindow = document.getElementById('config-window');
        if (configWindow) {
            configWindow.style.display = 'block';
        }

        state.isExpanded = true;

        // 绑定配置窗口事件
        bindConfigWindowEvents();

        // 加载现有配置
        await loadExistingConfig();

    } catch (error) {
        console.error('Failed to open config window:', error);
        showToast('无法打开配置窗口', 'error');
    }
}

// 关闭配置窗口
async function closeConfigWindow() {
    try {
        const configWindow = document.getElementById('config-window');
        if (configWindow) {
            configWindow.remove();
        }

        if (elements.floatingBar) {
            elements.floatingBar.style.display = 'flex';
        }
        document.body.classList.remove('expanded');
        await CollapseWindow();
        state.isExpanded = false;
    } catch (error) {
        console.error('Failed to close config window:', error);
        showToast('无法关闭配置窗口', 'error');
    }
}

// 绑定配置窗口事件
function bindConfigWindowEvents() {
    const closeConfigBtn = document.getElementById('close-config-btn');
    const configMinimizeBtn = document.getElementById('config-minimize-btn');
    const saveConfigBtn = document.getElementById('save-config-btn');
    const resetConfigBtn = document.getElementById('reset-config-btn');

    if (closeConfigBtn) {
        closeConfigBtn.addEventListener('click', closeConfigWindow);
    }

    if (configMinimizeBtn) {
        configMinimizeBtn.addEventListener('click', () => {
            import('../../../wailsjs/runtime/runtime.js').then(({ WindowMinimise }) => {
                WindowMinimise();
            });
        });
    }

    if (saveConfigBtn) {
        saveConfigBtn.addEventListener('click', saveConfig);
    }

    if (resetConfigBtn) {
        resetConfigBtn.addEventListener('click', resetConfig);
    }
}

// 加载现有配置
async function loadExistingConfig() {
    try {
        const { GetAllConfigs } = await import('../../../wailsjs/go/app/App.js');

        const configs = await GetAllConfigs();
        const configData = JSON.parse(configs);

        // 填充表单
        const apiUrlInput = document.getElementById('openai-api-url');
        const modelTypeInput = document.getElementById('openai-model-type');
        const apiKeyInput = document.getElementById('openai-api-key');

        if (apiUrlInput && configData.OPENAI_API_URL) {
            apiUrlInput.value = configData.OPENAI_API_URL.value;
        }

        if (modelTypeInput && configData.OPENAI_MODEL_TYPE) {
            modelTypeInput.value = configData.OPENAI_MODEL_TYPE.value;
        }

        if (apiKeyInput && configData.OPENAI_API_KEY) {
            apiKeyInput.value = configData.OPENAI_API_KEY.value;
        }

    } catch (error) {
        console.error('Failed to load existing config:', error);
        showToast('加载配置失败', 'error');
    }
}

// 保存配置
async function saveConfig() {
    try {
        const { SetConfigs, ValidateConfigs } = await import('../../../wailsjs/go/app/App.js');

        // 获取表单数据
        const apiUrl = document.getElementById('openai-api-url')?.value || '';
        const modelType = document.getElementById('openai-model-type')?.value || '';
        const apiKey = document.getElementById('openai-api-key')?.value || '';

        // 构建配置数据
        const configData = {
            'OPENAI_API_URL': apiUrl,
            'OPENAI_MODEL_TYPE': modelType,
            'OPENAI_API_KEY': apiKey
        };

        // 保存配置
        await SetConfigs(JSON.stringify(configData));

        // 验证配置
        const validationResult = await ValidateConfigs();
        showToast(validationResult, 'success');

        // 保存成功后自动关闭配置窗口
        setTimeout(() => {
            closeConfigWindow();
        }, 1000); // 延迟1秒关闭，让用户看到成功提示

    } catch (error) {
        console.error('Failed to save config:', error);
        showToast('保存配置失败: ' + error.message, 'error');
    }
}

// 重置配置
async function resetConfig() {
    try {
        const { ResetConfig } = await import('../../../wailsjs/go/app/App.js');

        // 重置所有配置项
        await ResetConfig('OPENAI_API_URL');
        await ResetConfig('OPENAI_MODEL_TYPE');
        await ResetConfig('OPENAI_API_KEY');

        // 清空表单
        const apiUrlInput = document.getElementById('openai-api-url');
        const modelTypeInput = document.getElementById('openai-model-type');
        const apiKeyInput = document.getElementById('openai-api-key');

        if (apiUrlInput) apiUrlInput.value = '';
        if (modelTypeInput) modelTypeInput.value = '';
        if (apiKeyInput) apiKeyInput.value = '';

        showToast('配置已重置', 'success');

    } catch (error) {
        console.error('Failed to reset config:', error);
        showToast('重置配置失败: ' + error.message, 'error');
    }
}

// 暴露公共接口给其他模块
window.loadPlugins = async () => {
    if (toolboxModule && toolboxModule.loadPlugins) {
        return await toolboxModule.loadPlugins();
    }
};

window.loadClipboardItems = async () => {
    if (clipboardModule && clipboardModule.loadClipboardItems) {
        return await clipboardModule.loadClipboardItems();
    }
};

// 启动应用
document.addEventListener('DOMContentLoaded', () => {
    initApp();
});

// 打开AI对话框
async function openAiChatDialog() {
    try {
        await ExpandWindow(950, 600); // AI对话窗口：950x600
        document.body.classList.add('expanded');

        // 隐藏其他窗口
        if (elements.floatingWindow) {
            elements.floatingWindow.style.display = 'none';
        }
        if (elements.clipboardWindow) {
            elements.clipboardWindow.style.display = 'none';
        }
        if (elements.floatingBar) {
            elements.floatingBar.style.display = 'none';
        }

        // 显示AI对话框
        if (elements.aiChatModal) {
            elements.aiChatModal.style.display = 'flex';
        }

        state.isExpanded = true;

        // 创建新会话
        await createNewChatSession();

        // 聚焦到输入框
        setTimeout(() => {
            if (elements.chatInput) {
                elements.chatInput.focus();
            }
        }, 300);

    } catch (error) {
        console.error('Failed to open AI chat dialog:', error);
        showToast('无法打开AI对话框', 'error');
    }
}

// 关闭AI对话框
async function closeAiChatDialog() {
    try {
        if (elements.aiChatModal) {
            elements.aiChatModal.style.display = 'none';
        }

        if (elements.floatingBar) {
            elements.floatingBar.style.display = 'flex';
        }
        document.body.classList.remove('expanded');
        await CollapseWindow();
        state.isExpanded = false;
    } catch (error) {
        console.error('Failed to close AI chat dialog:', error);
        showToast('无法关闭AI对话框', 'error');
    }
}

// 发送AI消息
async function sendAiMessage() {
    if (!elements.chatInput || !elements.chatSendBtn) return;

    const message = elements.chatInput.value.trim();
    if (!message) return;

    // 清空输入框
    elements.chatInput.value = '';

    // 显示发送按钮为加载状态
    elements.chatSendBtn.disabled = true;
    elements.chatSendBtn.innerHTML = '<span class="send-icon">◐</span>';

    try {
        // 显示用户消息
        addMessageToChat(message, 'user');

        // 调用AI API（使用会话上下文）
        const { ChatWithSession } = await import('../../../wailsjs/go/app/App.js');
        const response = await ChatWithSession(currentSessionId, message);

        // 更新当前会话ID（如果是新会话）
        currentSessionId = response.sessionId;

        // 显示AI回复
        addMessageToChat(response.message.content, 'ai');

    } catch (error) {
        console.error('AI chat error:', error);
        addMessageToChat('抱歉，我现在无法回复。请检查网络连接或API配置。', 'ai', true);
    } finally {
        // 恢复发送按钮
        elements.chatSendBtn.disabled = false;
        elements.chatSendBtn.innerHTML = '<span class="send-icon">→</span>';

        // 聚焦回输入框
        if (elements.chatInput) {
            elements.chatInput.focus();
        }
    }
}

// 添加消息到聊天界面
function addMessageToChat(message, type, isError = false) {
    if (!elements.chatMessages) return;

    const wrapperDiv = document.createElement('div');
    wrapperDiv.className = `message-wrapper ${type}-message-wrapper ${isError ? 'error-message' : ''}`;

    wrapperDiv.innerHTML = `
        <div class="message-content">
            <div class="message-bubble">
                <div class="message-text">${escapeHtml(message)}</div>
            </div>
        </div>
    `;

    elements.chatMessages.appendChild(wrapperDiv);

    // 滚动到底部
    setTimeout(() => {
        if (elements.chatMessages) {
            elements.chatMessages.scrollTop = elements.chatMessages.scrollHeight;
        }
    }, 100);
}

// 转义HTML字符
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// 创建新聊天会话
async function createNewChatSession() {
    try {
        // 调用后端创建新会话
        const { CreateNewChatSession } = await import('../../../wailsjs/go/app/App.js');
        const sessionId = await CreateNewChatSession();

        // 更新当前会话ID
        currentSessionId = sessionId;

        // 清空聊天消息
        if (elements.chatMessages) {
            elements.chatMessages.innerHTML = `
                <div class="message-wrapper ai-message-wrapper">
                    <div class="message-content">
                        <div class="message-bubble">
                            <div class="message-text">你好！欢迎来到个人终端插件工厂，你有想要的工具，可以工具箱里创建，如果对方案有不确定的，可以问我。你可以叫我LOJI。</div>
                        </div>
                    </div>
                </div>
            `;
        }

        showToast('已创建新会话', 'success');

    } catch (error) {
        console.error('Failed to create new chat session:', error);
        showToast('创建新会话失败', 'error');
    }
}

// 暴露给全局，供其他模块调用
window.mainModule = {
    expandApp,
    collapseApp,
    openClipboardWindow,
    closeClipboardWindow,
    openConfigWindow,
    closeConfigWindow,
    openAiChatDialog,
    closeAiChatDialog
};
