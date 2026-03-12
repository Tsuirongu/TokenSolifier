// 工具箱模块 - 处理工具管理和执行
import {
    GetAllPlugins, GetAllTags, GeneratePlugin, ExecutePlugin, DeletePlugin, GetPluginExecuteCode, UploadTempFile
} from '../../../wailsjs/go/app/App.js';

// 标签数据（临时存储，实际应该从后端获取）
const DEFAULT_LABELS = [
    { id: 1, name: 'AI', icon: '🤖', description: '使用人工智能能力' },
    { id: 2, name: '网络', icon: '🌐', description: '网络请求和API调用' },
    { id: 3, name: '文件', icon: '📁', description: '文件操作和处理' },
    { id: 4, name: '数据库', icon: '🗄️', description: '数据库操作' },
    { id: 5, name: '工具', icon: '🔗', description: '依赖其他工具' },
    { id: 6, name: '文本处理', icon: '📝', description: '文本分析和处理' },
    { id: 7, name: '图像处理', icon: '🖼️', description: '图像处理和分析' },
    { id: 8, name: '音频处理', icon: '🎵', description: '音频处理和分析' },
    { id: 9, name: '视频处理', icon: '🎥', description: '视频处理和分析' }
];

// 分页配置常量
const PAGINATION_CONFIG = {
    ITEMS_PER_PAGE: 5, // 每页显示8个工具
};
import {
    state, elements, showToast, setButtonLoading, escapeHtml,
    showProgress, hideProgress, listenToProgressEvents
} from '../common/common.js';

// 工具箱模块初始化
export async function initToolbox() {
    // 加载工具箱HTML内容
    await loadToolboxContent();

    // 初始化分页状态
    initPaginationState();

    // 绑定工具箱事件
    bindToolboxEvents();

    // 加载插件列表
    await loadPlugins();

    // 加载标签列表
    await loadTags();

    // 初始化显示Banner图片（未选择工具时的默认状态）
    if (elements.detailBanner) {
        elements.detailBanner.style.display = 'block';
    }
}

// 工具箱HTML模板 - 从外部HTML文件加载
let toolboxHTML = null;

// 加载工具箱HTML内容
async function loadToolboxContent() {
    try {
        // 动态加载HTML模板
        if (!toolboxHTML) {
            const htmlModule = await import('./toolbox.html?raw');
            toolboxHTML = htmlModule.default;
        }
        
        // 将HTML内容插入到主窗口中
        const windowContent = elements.floatingWindow?.querySelector('.window-content');
        if (windowContent) {
            windowContent.innerHTML = toolboxHTML;
        }

        // 重新初始化元素引用
        initToolboxElements();
    } catch (error) {
        console.error('加载工具箱内容失败:', error);
        showToast('加载工具箱失败', 'error');
    }
}

// 初始化分页状态
function initPaginationState() {
    state.paginationState = {
        currentPage: 1,
        itemsPerPage: PAGINATION_CONFIG.ITEMS_PER_PAGE,
        totalItems: 0,
        totalPages: 0
    };
}

// 初始化工具箱元素引用
function initToolboxElements() {
    // 这些元素在HTML加载后才可用
    elements.addToolBtn = document.getElementById('add-tool-btn');
    elements.refreshBtn = document.getElementById('refresh-btn');
    elements.toolCreator = document.getElementById('tool-creator');
    elements.requirementInput = document.getElementById('requirement-input');
    elements.enableToolDependencies = document.getElementById('enable-tool-dependencies');
    elements.closeCreatorBtn = document.getElementById('close-creator-btn');
    elements.cancelCreatorBtn = document.getElementById('cancel-creator-btn');
    elements.generateBtn = document.getElementById('generate-btn');
    elements.toolsList = document.getElementById('tools-list');
    elements.toolsCount = document.getElementById('tools-count');
    
    // 分页控件元素
    elements.paginationContainer = document.getElementById('pagination-container');
    elements.prevPageBtn = document.getElementById('prev-page-btn');
    elements.nextPageBtn = document.getElementById('next-page-btn');
    elements.paginationInfo = document.getElementById('pagination-info');

    elements.detailTitle = document.getElementById('detail-title');
    elements.detailDescription = document.getElementById('detail-description');
    elements.detailBanner = document.getElementById('detail-banner');
    elements.detailInput = document.getElementById('detail-input');
    elements.detailOutput = document.getElementById('detail-output');
    elements.detailExecuteBtn = document.getElementById('detail-execute-btn');
    elements.detailDeleteBtn = document.getElementById('detail-delete-btn');


    // 代码预览模态框相关元素
    elements.detailCodeBtn = document.getElementById('detail-code-btn');
    elements.codeModal = document.getElementById('code-modal');
    elements.codeModalCloseBtn = document.getElementById('code-modal-close-btn');
    elements.codeCloseBtn = document.getElementById('code-close-btn');
    elements.codeCopyBtn = document.getElementById('code-copy-btn');
    elements.codeContent = document.getElementById('code-content');

    // 动态表单相关元素
    elements.inputFormContainer = document.getElementById('input-form-container');
    elements.jsonInputContainer = document.getElementById('json-input-container');
    elements.inputForm = document.getElementById('input-form');
    elements.formToggleBtn = document.getElementById('form-toggle-btn');
    elements.detailInput = document.getElementById('detail-input');

    // 依赖信息相关元素
    elements.pluginDependencies = document.getElementById('plugin-dependencies');
    elements.dependenciesList = document.getElementById('dependencies-list');
    elements.dependenciesWarning = document.getElementById('dependencies-warning');

    // 初始化输入输出相关元素的显示状态（默认隐藏）
    if (elements.inputFormContainer) {
        elements.inputFormContainer.style.display = 'none';
    }
    if (elements.jsonInputContainer) {
        elements.jsonInputContainer.style.display = 'none';
    }
    const outputGroup = document.querySelector('.output-group');
    if (outputGroup) {
        outputGroup.style.display = 'none';
    }
    if (elements.detailExecuteBtn) {
        elements.detailExecuteBtn.style.display = 'none';
    }
    if (elements.detailDeleteBtn) {
        elements.detailDeleteBtn.style.display = 'none';
    }
    if (elements.detailCodeBtn) {
        elements.detailCodeBtn.style.display = 'none';
    }
}

// 绑定工具箱事件
function bindToolboxEvents() {
    // 工具创建
    if (elements.addToolBtn) {
        elements.addToolBtn.addEventListener('click', showCreator);
    }
    if (elements.refreshBtn) {
        elements.refreshBtn.addEventListener('click', loadPlugins);
    }
    if (elements.closeCreatorBtn) {
        elements.closeCreatorBtn.addEventListener('click', hideCreator);
    }
    if (elements.cancelCreatorBtn) {
        elements.cancelCreatorBtn.addEventListener('click', hideCreator);
    }
    if (elements.generateBtn) {
        elements.generateBtn.addEventListener('click', handleGenerate);
    }
    
    // 分页按钮事件
    if (elements.prevPageBtn) {
        elements.prevPageBtn.addEventListener('click', handlePrevPage);
    }
    if (elements.nextPageBtn) {
        elements.nextPageBtn.addEventListener('click', handleNextPage);
    }

    // 右侧详情执行和删除
    if (elements.detailExecuteBtn) {
        elements.detailExecuteBtn.addEventListener('click', handleExecuteInDetail);
    }
    if (elements.detailDeleteBtn) {
        elements.detailDeleteBtn.addEventListener('click', () => {
            if (state.currentPlugin) {
                handleDelete(state.currentPlugin);
            }
        });
    }

    // 代码预览模态框事件
    if (elements.detailCodeBtn) {
        elements.detailCodeBtn.addEventListener('click', showCodeModal);
    }
    if (elements.codeModalCloseBtn) {
        elements.codeModalCloseBtn.addEventListener('click', hideCodeModal);
    }
    if (elements.codeCloseBtn) {
        elements.codeCloseBtn.addEventListener('click', hideCodeModal);
    }
    if (elements.codeCopyBtn) {
        elements.codeCopyBtn.addEventListener('click', copyCodeToClipboard);
    }

    // 表单切换事件
    if (elements.formToggleBtn) {
        elements.formToggleBtn.addEventListener('click', toggleInputMode);
    }

    // 绑定复制字段事件（在需要时动态绑定）
    bindCopyFieldEvents();

    // Tooltip事件
    bindTooltipEvents();
}

// 显示创建面板
export function showCreator() {
    if (elements.toolCreator) {
        elements.toolCreator.style.display = 'block';
        if (elements.requirementInput) {
            elements.requirementInput.focus();
        }
        // 重置选中的标签并渲染标签选择器
        state.selectedTagIds = [];
        renderTagOptions();
    }
}

// 隐藏创建面板
export function hideCreator() {
    if (elements.toolCreator) {
        elements.toolCreator.style.display = 'none';
    }
    if (elements.requirementInput) {
        elements.requirementInput.value = '';
    }
    // 清除标签选择
    clearLabelSelection();
}

// 渲染标签选择器
function renderLabelSelector() {
    const labelSelector = document.getElementById('label-selector');
    if (!labelSelector) return;

    labelSelector.innerHTML = '';

    DEFAULT_LABELS.forEach(label => {
        const labelElement = document.createElement('div');
        labelElement.className = 'label-option';
        labelElement.dataset.labelId = label.id;
        labelElement.innerHTML = `
            <span class="label-option-icon">${label.icon}</span>
            <span>${label.name}</span>
        `;

        labelElement.addEventListener('click', () => toggleLabelSelection(label.id));

        labelSelector.appendChild(labelElement);
    });
}

// 切换标签选择状态
function toggleLabelSelection(labelId) {
    const labelElement = document.querySelector(`.label-option[data-label-id="${labelId}"]`);
    if (!labelElement) return;

    labelElement.classList.toggle('selected');
}

// 清除标签选择
function clearLabelSelection() {
    const selectedLabels = document.querySelectorAll('.label-option.selected');
    selectedLabels.forEach(label => label.classList.remove('selected'));
}

// 获取选中的标签ID列表
function getSelectedLabelIds() {
    const selectedLabels = document.querySelectorAll('.label-option.selected');
    return Array.from(selectedLabels).map(label => parseInt(label.dataset.labelId));
}

// 获取示例输入（生成标准格式）
function getExampleInput(plugin) {
    // 生成标准格式的示例数据
    const exampleData = {
        textList: [],
        imageList: [],
        fileList: [],
        audioList: [],
        videoList: [],
        documentList: [],
        otherList: []
    };

    if (plugin?.input) {
        // 根据插件的输入描述生成示例数据
        if (plugin.input.textList && plugin.input.textList.length > 0) {
            plugin.input.textList.forEach(field => {
                // field现在是InputItem对象
                const fieldName = (field.content || field.name || '字段').toLowerCase();
                const fieldType = field.contentType || 'text';

                // 处理boolean类型
                if (fieldType === 'boolean') {
                    exampleData.textList.push('true'); // 默认true
                }
                // 处理select类型
                else if (fieldType === 'select' && field.options && field.options.length > 0) {
                    const normalizedOpts = normalizeOptions(field.options);
                    if (normalizedOpts.length > 0) {
                        exampleData.textList.push(normalizedOpts[0].value); // 使用value而非label
                    }
                }
                // 处理radio类型
                else if (fieldType === 'radio' && field.options && field.options.length > 0) {
                    const normalizedOpts = normalizeOptions(field.options);
                    if (normalizedOpts.length > 0) {
                        exampleData.textList.push(normalizedOpts[0].value); // 使用value而非label
                    }
                }
                // 处理checkbox类型
                else if (fieldType === 'checkbox' && field.options && field.options.length > 0) {
                    const normalizedOpts = normalizeOptions(field.options);
                    // 多选类型，选择前两个选项作为示例
                    const selectedOptions = normalizedOpts.slice(0, Math.min(2, normalizedOpts.length));
                    selectedOptions.forEach(option => {
                        exampleData.textList.push(option.value); // 使用value而非label
                    });
                }
                // 处理number类型
                else if (fieldType === 'number') {
                    if (fieldName.includes('年龄') || fieldName.includes('age')) {
                        exampleData.textList.push('25');
                    } else if (fieldName.includes('身高') || fieldName.includes('height')) {
                        exampleData.textList.push('175');
                    } else if (fieldName.includes('体重') || fieldName.includes('weight')) {
                        exampleData.textList.push('70');
                    } else if (fieldName.includes('数值') || fieldName.includes('value') || fieldName.includes('数量') || fieldName.includes('数字')) {
                        exampleData.textList.push('100');
                    } else {
                        exampleData.textList.push('0');
                    }
                }
                // 处理其他文本类型
                else {
                    if (fieldName.includes('年龄') || fieldName.includes('age')) {
                        exampleData.textList.push('25');
                    } else if (fieldName.includes('身高') || fieldName.includes('height')) {
                        exampleData.textList.push('175');
                    } else if (fieldName.includes('体重') || fieldName.includes('weight')) {
                        exampleData.textList.push('70');
                    } else if (fieldName.includes('性别') || fieldName.includes('gender')) {
                        exampleData.textList.push('男');
                    } else if (fieldName.includes('表达式') || fieldName.includes('expression')) {
                        exampleData.textList.push('10 + 20');
                    } else if (fieldName.includes('数值') || fieldName.includes('value') || fieldName.includes('数字')) {
                        exampleData.textList.push('100');
                    } else if (fieldName.includes('文本') || fieldName.includes('text') || fieldName.includes('内容')) {
                        exampleData.textList.push('示例文本');
                    } else if (fieldName.includes('名称') || fieldName.includes('name')) {
                        exampleData.textList.push('示例名称');
                    } else {
                        exampleData.textList.push(`示例${field.content || field.name || '字段'}`);
                    }
                }
            });
        }

        // 为其他类型字段添加示例数据
        if (plugin.input.imageList && plugin.input.imageList.length > 0) {
            plugin.input.imageList.forEach(() => {
                exampleData.imageList.push('https://example.com/image.jpg');
            });
        }

        if (plugin.input.fileList && plugin.input.fileList.length > 0) {
            plugin.input.fileList.forEach(() => {
                exampleData.fileList.push('/path/to/example/file.txt');
            });
        }

        if (plugin.input.audioList && plugin.input.audioList.length > 0) {
            plugin.input.audioList.forEach(() => {
                exampleData.audioList.push('https://example.com/audio.mp3');
            });
        }

        if (plugin.input.videoList && plugin.input.videoList.length > 0) {
            plugin.input.videoList.forEach(() => {
                exampleData.videoList.push('https://example.com/video.mp4');
            });
        }

        if (plugin.input.documentList && plugin.input.documentList.length > 0) {
            plugin.input.documentList.forEach(() => {
                exampleData.documentList.push('/path/to/example/document.pdf');
            });
        }

        // 如果没有任何示例数据，回退到基于描述的生成
        const hasData = Object.values(exampleData).some(arr => arr.length > 0);
        if (!hasData) {
            return generateExampleFromDescription(plugin.description);
        }

        return exampleData;
    }

    // 回退到基于描述的生成
    return generateExampleFromDescription(plugin?.description);
}

// 基于描述生成示例输入的回退函数（标准格式）
function generateExampleFromDescription(description) {
    if (!description) {
        return {
            textList: ['示例输入'],
            imageList: [],
            fileList: [],
            audioList: [],
            videoList: [],
            documentList: [],
            otherList: []
        };
    }

    const lowerDesc = description.toLowerCase();
    const result = {
        textList: [],
        imageList: [],
        fileList: [],
        audioList: [],
        videoList: [],
        documentList: [],
        otherList: []
    };

    if (lowerDesc.includes('体脂') || lowerDesc.includes('bmi')) {
        result.textList = ['70', '175', '25', '男'];
    } else if (lowerDesc.includes('计算')) {
        result.textList = ['10 + 20'];
    } else if (lowerDesc.includes('翻译')) {
        result.textList = ['Hello world', '英语', '中文'];
    } else if (lowerDesc.includes('天气')) {
        result.textList = ['北京'];
    } else {
        result.textList = ['示例输入'];
    }

    return result;
}



// 加载插件列表
export async function loadPlugins() {
    try {
        const plugins = await GetAllPlugins();
        state.plugins = plugins || [];
        renderPlugins();
    } catch (error) {
        console.log(error);

        showToast('加载工具列表失败', 'error');
    }
}

// 加载标签列表
export async function loadTags() {
    try {
        const tags = await GetAllTags();
        state.tags = tags || [];
        renderTagOptions();
    } catch (error) {
        console.log('加载标签列表失败:', error);
        showToast('加载标签列表失败', 'error');
    }
}

// 渲染标签选项
function renderTagOptions() {
    const tagOptions = document.getElementById('tag-options');
    if (!tagOptions) return;

    tagOptions.innerHTML = '';

    if (state.tags.length === 0) {
        tagOptions.innerHTML = '<div class="no-tags">暂无标签</div>';
        return;
    }

    state.tags.forEach(tag => {
        const tagElement = document.createElement('div');
        tagElement.className = 'tag-option';
        tagElement.dataset.tagId = tag.id;

        // 检查是否已选中
        if (state.selectedTagIds.includes(tag.id)) {
            tagElement.classList.add('selected');
        }

        tagElement.innerHTML = `
            <div class="tag-color" style="background-color: ${tag.color || '#ccc'}"></div>
            <span>${tag.name}</span>
        `;

        tagElement.addEventListener('click', () => toggleTagSelection(tag.id));
        tagOptions.appendChild(tagElement);
    });
}

// 切换标签选择
function toggleTagSelection(tagId) {
    const index = state.selectedTagIds.indexOf(tagId);
    if (index > -1) {
        // 已选中，取消选择
        state.selectedTagIds.splice(index, 1);
    } else {
        // 未选中，添加选择
        state.selectedTagIds.push(tagId);
    }

    // 重新渲染标签选项
    renderTagOptions();
}


// 渲染插件列表（带分页）
function renderPlugins() {
    if (!elements.toolsCount || !elements.toolsList) return;

    const count = state.plugins.length;
    elements.toolsCount.textContent = count;

    // 更新分页状态
    const pagination = state.paginationState;
    pagination.totalItems = count;
    pagination.totalPages = Math.ceil(count / pagination.itemsPerPage) || 1;
    
    // 确保当前页在有效范围内
    if (pagination.currentPage > pagination.totalPages) {
        pagination.currentPage = pagination.totalPages;
    }
    if (pagination.currentPage < 1) {
        pagination.currentPage = 1;
    }

    if (count === 0) {
        if (elements.toolsList) {
            elements.toolsList.innerHTML = `
                <div class="empty-state">
                    <div class="empty-icon">📦</div>
                    <p>还没有工具</p>
                    <p class="empty-hint">点击上方按钮创建第一个工具</p>
                </div>
            `;
        }
        // 隐藏分页控件
        if (elements.paginationContainer) {
            elements.paginationContainer.style.display = 'none';
        }
        return;
    }

    // 计算当前页的开始和结束索引
    const startIndex = (pagination.currentPage - 1) * pagination.itemsPerPage;
    const endIndex = startIndex + pagination.itemsPerPage;
    const currentPlugins = state.plugins.slice(startIndex, endIndex);

    if (elements.toolsList) {
        elements.toolsList.innerHTML = currentPlugins.map(plugin => {
            // 生成标签HTML
            let labelsHtml = '';
            if (plugin.labels && plugin.labels.length > 0) {
                const labels = plugin.labels.map(relation => {
                    const label = relation.label || DEFAULT_LABELS.find(l => l.id === relation.labelId);
                    if (label) {
                        return `<span class="plugin-label" style="background: ${label.color || '#667eea'}"><span class="plugin-label-icon">${label.icon}</span>${label.name}</span>`;
                    }
                    return '';
                }).filter(html => html !== '');

                if (labels.length > 0) {
                    labelsHtml = `<div class="plugin-labels">${labels.join('')}</div>`;
                }
            }

            return `
            <div class="tool-item" data-id="${plugin.id}">
                <div class="tool-info">
                    <div class="tool-name">${escapeHtml(plugin.chineseName || plugin.name)}</div>
                    ${labelsHtml}
                    <div class="tool-description">${escapeHtml(plugin.description)}</div>
                </div>
                <div class="tool-actions">
                    <button class="tool-action-btn delete" data-action="delete" title="删除">🗑</button>
                </div>
            </div>
            `;
        }).join('');

        // 绑定事件
        elements.toolsList.querySelectorAll('.tool-item').forEach(item => {
            const pluginId = parseInt(item.dataset.id);
            const plugin = state.plugins.find(p => p.id === pluginId);

            if (!plugin) return;

            const deleteBtn = item.querySelector('[data-action="delete"]');
            if (deleteBtn) {
                deleteBtn.addEventListener('click', (e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    handleDelete(plugin);
                });
            }

            const infoArea = item.querySelector('.tool-info');
            if (infoArea) {
                infoArea.addEventListener('click', () => selectPlugin(plugin));
            }
        });
    }

    // 更新分页控件的显示状态和内容
    updatePaginationUI();
}

// 选择左侧工具，填充右侧详情
function selectPlugin(plugin) {
    state.currentPlugin = plugin;

    // 标记选中项
    if (elements.toolsList) {
        elements.toolsList.querySelectorAll('.tool-item').forEach(i => i.classList.remove('active'));
        const activeEl = elements.toolsList.querySelector(`.tool-item[data-id="${plugin.id}"]`);
        if (activeEl) activeEl.classList.add('active');
    }

    // 控制图片显示：未选择工具时显示图片，选择工具时隐藏图片
    if (elements.detailBanner) {
        elements.detailBanner.style.display = plugin ? 'none' : 'block';
    }

    // 控制输入输出相关元素的显示：只有选择工具时才显示
    const shouldShowToolElements = plugin ? 'block' : 'none';
    const shouldShowToolElementsFlex = plugin ? 'inline-flex' : 'none';

    // 隐藏/显示输入表单容器
    if (elements.inputFormContainer) {
        elements.inputFormContainer.style.display = shouldShowToolElements;
    }

    // 隐藏/显示JSON输入容器
    if (elements.jsonInputContainer) {
        elements.jsonInputContainer.style.display = shouldShowToolElements;
    }

    // 隐藏/显示输出组
    const outputGroup = document.querySelector('.output-group');
    if (outputGroup) {
        outputGroup.style.display = shouldShowToolElements;
    }

    // 隐藏/显示操作按钮
    if (elements.detailExecuteBtn) {
        elements.detailExecuteBtn.style.display = shouldShowToolElementsFlex;
    }
    if (elements.detailDeleteBtn) {
        elements.detailDeleteBtn.style.display = shouldShowToolElementsFlex;
    }
    if (elements.detailCodeBtn) {
        elements.detailCodeBtn.style.display = shouldShowToolElementsFlex;
    }

    // 标题与描述
    if (elements.detailTitle) {
        elements.detailTitle.textContent = plugin ? `🔧 ${plugin.chineseName || plugin.name}` : '请选择左侧工具';
    }
    if (elements.detailDescription) {
        elements.detailDescription.textContent = plugin?.description || '在这里将展示工具描述。';
    }

    // 显示插件标签
    displayPluginLabels(plugin);

    // 显示插件依赖信息
    displayPluginDependencies(plugin);

    console.log(plugin);

    console.log('Plugin object:', plugin);
    console.log('Plugin input:', plugin?.input);
    console.log('Plugin output:', plugin?.output);

    // 生成输入表单或设置示例JSON
    if (plugin?.input) {
        generateInputForm(plugin);
        // 隐藏JSON编辑器，显示表单
        showInputForm();
    } else {
        // 如果没有输入描述，回退到JSON编辑模式
        showJsonInput();
        if (elements.detailInput) {
            const exampleInput = getExampleInput(plugin);
            elements.detailInput.value = JSON.stringify(exampleInput, null, 2);
        }
    }

    // 显示操作按钮
    if (elements.detailExecuteBtn) {
        elements.detailExecuteBtn.style.display = 'inline-flex';
        elements.detailExecuteBtn.disabled = false;
        elements.detailExecuteBtn.textContent = '▶ 执行';
        elements.detailExecuteBtn.style.opacity = '1';
    }
    if (elements.detailDeleteBtn) {
        elements.detailDeleteBtn.style.display = 'inline-flex';
    }
    if (elements.detailCodeBtn) {
        elements.detailCodeBtn.style.display = 'inline-flex';
    }

    // 重置输出
    if (elements.detailOutput) {
        elements.detailOutput.textContent = '等待执行...';
        elements.detailOutput.className = 'executor-output';
    }

    // 重新绑定复制事件
    bindCopyFieldEvents();
}

// 在右侧详情中执行工具
async function handleExecuteInDetail() {
    if (!state.currentPlugin) {
        showToast('请先选择一个工具', 'error');
        return;
    }

    let inputData;

    // 检查当前是表单模式还是JSON模式
    const isFormVisible = elements.inputFormContainer?.style.display !== 'none';

    if (isFormVisible) {
        // 从表单收集数据
        inputData = await collectFormData();
        if (!inputData) {
            showToast('请填写至少一个参数', 'error');
            return;
        }
    } else {
        // 从JSON输入框获取数据
        if (!elements.detailInput) return;

        const input = elements.detailInput.value.trim();
        if (!input) {
            showToast('请输入参数', 'error');
            return;
        }

        try {
            inputData = JSON.parse(input);
        } catch (e) {
            showToast('输入参数必须是有效的JSON格式', 'error');
            return;
        }
    }

    // 将数据转换为JSON字符串
    const inputJson = JSON.stringify(inputData);

    setButtonLoading(elements.detailExecuteBtn, true);
    if (elements.detailOutput) {
        elements.detailOutput.textContent = '执行中...';
        elements.detailOutput.className = 'executor-output';
    }

    try {
        const result = await ExecutePlugin(state.currentPlugin.id, inputJson);

        if (elements.detailOutput) {
            // 确保result是字符串类型
            const safeResult = result && typeof result === 'string' ? result : String(result || '');

            // 根据插件输出描述格式化显示结果
            const formattedResult = formatExecutionResult(safeResult, state.currentPlugin);
            elements.detailOutput.innerHTML = formattedResult;

            // 重新绑定复制事件，确保DOM更新后事件能正确工作
            bindCopyFieldEvents();

            // 加载本地图片
            loadLocalImages();

            // 绑定文件操作事件
            bindFileActionEvents();

            elements.detailOutput.classList.add('success');
        }
        showToast('执行成功', 'success');
    } catch (error) {
        if (elements.detailOutput) {
            const errorMessage = '错误: ' + error;
            elements.detailOutput.textContent = errorMessage;
            elements.detailOutput.classList.add('error');
        }

        showToast('执行失败', 'error');
    } finally {
        setButtonLoading(elements.detailExecuteBtn, false);
    }
}

// 格式化执行结果显示
function formatExecutionResult(result, plugin) {
    let parsedResult;

    // 检查result是否有效
    if (!result || typeof result !== 'string') {
        return `<div class="execution-result">
            <div class="result-item">
                <div class="result-label">执行结果</div>
                <div class="result-value">${escapeHtml(String(result || ''))}</div>
            </div>
        </div>`;
    }

    // 尝试解析JSON结果
    try {
        parsedResult = JSON.parse(result);
    } catch {
        // 如果不是JSON格式，直接显示原文
        return `<div class="execution-result">
            <div class="result-item">
                <div class="result-label">执行结果</div>
                <div class="result-value">${escapeHtml(result)}</div>
            </div>
        </div>`;
    }

    // 如果插件没有输出描述，直接显示JSON
    if (!plugin?.output) {
        return `<div class="execution-result">
            <pre>${JSON.stringify(parsedResult, null, 2)}</pre>
        </div>`;
    }

    const outputFields = [];

    // 处理文本类型输出
    if (plugin.output.textList && plugin.output.textList.length > 0) {
        const textValues = Array.isArray(parsedResult.textList) ? parsedResult.textList : [];
        plugin.output.textList.forEach((field, index) => {
            // field现在是InputItem对象
            const fieldName = field.content || field.name || `文本字段 ${index + 1}`;
            const fieldType = getDisplayTypeName(field.contentType);
            const value = textValues[index] || '无数据';
            outputFields.push({
                name: fieldName,
                type: fieldType,
                value: escapeHtml(String(value)),
                description: field.description || ''
            });
        });
    }

    // 处理图片类型输出
    if (plugin.output.imageList && plugin.output.imageList.length > 0) {
        const imageValues = Array.isArray(parsedResult.imageList) ? parsedResult.imageList : [];
        plugin.output.imageList.forEach((field, index) => {
            const fieldName = field.content || field.name || `图片字段 ${index + 1}`;
            const fieldType = getDisplayTypeName(field.contentType);
            const value = imageValues[index];
            if (value) {
                // 检查是否为本地文件路径（绝对路径或包含临时目录特征）
                const isLocalPath = value.startsWith('/') || value.startsWith('C:\\') || 
                                   value.includes('loji-outputs') || value.includes('/temp/') || 
                                   value.includes('/tmp/') || value.includes('\\Temp\\');
                outputFields.push({
                    name: fieldName,
                    type: fieldType,
                    value: isLocalPath 
                        ? `<div class="result-image-container" data-file-path="${escapeHtml(value)}">
                             <div class="loading-indicator">加载中...</div>
                           </div>` 
                        : `<img src="${escapeHtml(value)}" alt="${escapeHtml(fieldName)}" class="result-image" />`,
                    description: field.description || '',
                    isImage: true,
                    isLocalFile: isLocalPath,
                    filePath: value
                });
            }
        });
    }

    // 处理文件类型输出
    if (plugin.output.fileList && plugin.output.fileList.length > 0) {
        const fileValues = Array.isArray(parsedResult.fileList) ? parsedResult.fileList : [];
        plugin.output.fileList.forEach((field, index) => {
            const fieldName = field.content || field.name || `文件字段 ${index + 1}`;
            const fieldType = getDisplayTypeName(field.contentType);
            const value = fileValues[index];
            if (value) {
                const fileName = value.split('/').pop() || value;
                const isLocalPath = value.startsWith('/') || value.startsWith('C:\\') || 
                                   value.includes('loji-outputs') || value.includes('/temp/') || 
                                   value.includes('/tmp/') || value.includes('\\Temp\\');
                outputFields.push({
                    name: fieldName,
                    type: fieldType,
                    value: isLocalPath 
                        ? `<div class="result-file-actions">
                             <span class="file-name">${escapeHtml(fileName)}</span>
                             <button class="btn-download-file" data-file-path="${escapeHtml(value)}">下载</button>
                             <button class="btn-copy-file" data-file-path="${escapeHtml(value)}">复制到...</button>
                           </div>` 
                        : `<a href="${escapeHtml(value)}" target="_blank" class="result-file-link">${escapeHtml(fileName)}</a>`,
                    description: field.description || '',
                    isFile: true,
                    isLocalFile: isLocalPath,
                    filePath: value
                });
            }
        });
    }

    // 处理音频类型输出
    if (plugin.output.audioList && plugin.output.audioList.length > 0) {
        const audioValues = Array.isArray(parsedResult.audioList) ? parsedResult.audioList : [];
        plugin.output.audioList.forEach((field, index) => {
            const fieldName = field.content || field.name || `音频字段 ${index + 1}`;
            const fieldType = getDisplayTypeName(field.contentType);
            const value = audioValues[index];
            if (value) {
                outputFields.push({
                    name: fieldName,
                    type: fieldType,
                    value: `<audio controls class="result-audio"><source src="${escapeHtml(value)}"></audio>`,
                    description: field.description || ''
                });
            }
        });
    }

    // 处理视频类型输出
    if (plugin.output.videoList && plugin.output.videoList.length > 0) {
        const videoValues = Array.isArray(parsedResult.videoList) ? parsedResult.videoList : [];
        plugin.output.videoList.forEach((field, index) => {
            const fieldName = field.content || field.name || `视频字段 ${index + 1}`;
            const fieldType = getDisplayTypeName(field.contentType);
            const value = videoValues[index];
            if (value) {
                outputFields.push({
                    name: fieldName,
                    type: fieldType,
                    value: `<video controls class="result-video"><source src="${escapeHtml(value)}"></video>`,
                    description: field.description || ''
                });
            }
        });
    }

    // 处理文档类型输出
    if (plugin.output.documentList && plugin.output.documentList.length > 0) {
        const documentValues = Array.isArray(parsedResult.documentList) ? parsedResult.documentList : [];
        plugin.output.documentList.forEach((field, index) => {
            const fieldName = field.content || field.name || `文档字段 ${index + 1}`;
            const fieldType = getDisplayTypeName(field.contentType);
            const value = documentValues[index];
            if (value) {
                const fileName = value.split('/').pop() || value;
                const isLocalPath = value.startsWith('/') || value.startsWith('C:\\') || 
                                   value.includes('loji-outputs') || value.includes('/temp/') || 
                                   value.includes('/tmp/') || value.includes('\\Temp\\');
                outputFields.push({
                    name: fieldName,
                    type: fieldType,
                    value: isLocalPath 
                        ? `<div class="result-file-actions">
                             <span class="file-name">${escapeHtml(fileName)}</span>
                             <button class="btn-download-file" data-file-path="${escapeHtml(value)}">下载</button>
                             <button class="btn-copy-file" data-file-path="${escapeHtml(value)}">复制到...</button>
                           </div>` 
                        : `<a href="${escapeHtml(value)}" target="_blank" class="result-file-link">${escapeHtml(fileName)}</a>`,
                    description: field.description || '',
                    isFile: true,
                    isLocalFile: isLocalPath,
                    filePath: value
                });
            }
        });
    }

    // 处理其他类型输出
    if (plugin.output.otherList && plugin.output.otherList.length > 0) {
        const otherValues = Array.isArray(parsedResult.otherList) ? parsedResult.otherList : [];
        plugin.output.otherList.forEach((field, index) => {
            const fieldName = field.content || field.name || `其他字段 ${index + 1}`;
            const fieldType = getDisplayTypeName(field.contentType);
            const value = otherValues[index] || '无数据';
            outputFields.push({
                name: fieldName,
                type: fieldType,
                value: escapeHtml(String(value)),
                description: field.description || ''
            });
        });
    }

    // 处理插件错误信息
    if (parsedResult.pluginError && parsedResult.pluginError.trim()) {
        outputFields.unshift({
            name: '执行错误',
            type: '错误',
            value: `<span class="result-error">${escapeHtml(parsedResult.pluginError)}</span>`,
            description: '插件执行过程中发生错误'
        });
    }

    // 生成格式化的HTML
    if (outputFields.length === 0) {
        return `<div class="execution-result">
            <div class="no-results">暂无输出数据</div>
        </div>`;
    }

    return `<div class="execution-result">
        ${outputFields.map((field, index) => `
            <div class="result-item" data-field-index="${index}">
                <div class="result-header">
                    <span class="result-type-badge ${getResultTypeClass(field.type)}">${field.type}</span>
                    <span class="result-label">${escapeHtml(field.name)}</span>
                    <button class="copy-field-btn" data-field-index="${index}" title="复制此结果">
                        <span class="copy-icon">□</span>
                    </button>
                </div>
                ${field.description ? `<div class="result-description">${escapeHtml(field.description)}</div>` : ''}
                <div class="result-value">${field.value}</div>
            </div>
        `).join('')}
    </div>`;
}

// 获取结果类型对应的CSS类名
function getResultTypeClass(type) {
    switch(type) {
        case '文本': return 'type-text';
        case '数字': return 'type-number';
        case '图片': return 'type-image';
        case '文件': return 'type-file';
        case '音频': return 'type-audio';
        case '视频': return 'type-video';
        case '文档': return 'type-document';
        case '其他': return 'type-other';
        case '错误': return 'type-error';
        default: return 'type-default';
    }
}

// 处理生成
async function handleGenerate() {
    if (!elements.requirementInput) return;

    const requirement = elements.requirementInput.value.trim();

    if (!requirement) {
        showToast('请输入工具需求描述', 'error');
        return;
    }

    if (state.isGenerating) return;

    state.isGenerating = true;
    setButtonLoading(elements.generateBtn, true);

    // 显示进度条
    showProgress();

    try {
        // 构造PluginCreationRequest对象
        const pluginRequest = {
            name: "", // AI会生成英文名
            chineseName: "", // AI会生成中文名
            description: "", // AI会生成描述
            userRequirement: requirement,
            tagIds: [...state.selectedTagIds] // 添加选中的标签ID
        };

        await GeneratePlugin(pluginRequest);
        showToast('工具生成成功！', 'success');
        hideCreator();
        
        // 重置分页到第一页
        if (state.paginationState) {
            state.paginationState.currentPage = 1;
        }
        
        await loadPlugins();
    } catch (error) {
        showToast('生成失败: ' + error, 'error');
        // 隐藏进度条
        hideProgress();
    } finally {
        state.isGenerating = false;
        setButtonLoading(elements.generateBtn, false);
    }
}

// 处理删除
async function handleDelete(plugin) {
    if (!plugin || !plugin.id) {
        showToast('无效的工具', 'error');
        return;
    }

    try {
        const pluginId = parseInt(plugin.id, 10);
        await DeletePlugin(pluginId);
        showToast('工具已删除', 'success');

        // 清除当前选中状态
        state.currentPlugin = null;
        if (elements.detailTitle) {
            elements.detailTitle.textContent = '请选择左侧工具';
        }
        if (elements.detailDescription) {
            elements.detailDescription.textContent = '';
        }
        if (elements.detailDeleteBtn) {
            elements.detailDeleteBtn.style.display = 'none';
        }
        if (elements.detailCodeBtn) {
            elements.detailCodeBtn.style.display = 'none';
        }

        // 重置输入模式到JSON编辑
        showJsonInput();

        // 重置分页到第一页
        if (state.paginationState) {
            state.paginationState.currentPage = 1;
        }

        await loadPlugins();

        // 如果当前有选中的插件，检查其依赖状态是否需要更新
        if (state.currentPlugin) {
            // 检查当前插件是否依赖于被删除的插件
            const currentPlugin = state.plugins.find(p => p.id === state.currentPlugin.id);
            if (currentPlugin) {
                // 重新显示依赖信息
                displayPluginDependencies(currentPlugin);
            }
        }
    } catch (error) {
        showToast('删除失败', 'error');
    }
}

// 显示代码预览模态框
async function showCodeModal() {
    if (!state.currentPlugin) {
        showToast('未选择插件', 'error');
        return;
    }

    try {
        // 调用后端函数获取yaegi组装后的执行代码
        const executeCode = await GetPluginExecuteCode(state.currentPlugin.id);

        console.log('显示代码预览:', {
            pluginId: state.currentPlugin.id,
            pluginName: state.currentPlugin.name,
            executeCodeLength: executeCode?.length,
            executeCodePreview: executeCode?.substring(0, 200) + '...'
        });

        if (elements.codeModal && elements.codeContent) {
            // 显示yaegi组装后的执行代码
            elements.codeContent.textContent = executeCode;

            // 显示模态框
            elements.codeModal.style.display = 'flex';

            // 聚焦到代码内容区域
            setTimeout(() => {
                elements.codeContent.scrollTop = 0;
            }, 100);
        }
    } catch (error) {
        console.error('获取插件执行代码失败:', error);
        showToast('获取插件代码失败: ' + error.message, 'error');
    }
}

// 隐藏代码预览模态框
function hideCodeModal() {
    if (elements.codeModal) {
        elements.codeModal.style.display = 'none';
    }
}

// 复制代码到剪贴板
async function copyCodeToClipboard() {
    if (!state.currentPlugin) {
        showToast('未选择插件', 'error');
        return;
    }

    try {
        // 获取页面上显示的yaegi包装后的执行代码
        const executeCode = await GetPluginExecuteCode(state.currentPlugin.id);
        const code = executeCode;

        // 使用现代的 Clipboard API
        if (navigator.clipboard && navigator.clipboard.writeText) {
            navigator.clipboard.writeText(code).then(() => {
                showToast('Go代码已复制到剪贴板', 'success');
            }).catch(err => {
                console.error('复制失败:', err);
                fallbackCopyTextToClipboard(code);
            });
        } else {
            // 回退到传统方法
            fallbackCopyTextToClipboard(code);
        }
    } catch (error) {
        console.error('获取插件执行代码失败:', error);
        showToast('复制代码失败: ' + error.message, 'error');
    }
}

// 复制代码到剪贴板（保留原有同步版本作为备用）
function copyCodeToClipboardSync() {
    if (!state.currentPlugin || !state.currentPlugin.code) {
        showToast('没有可复制的代码', 'error');
        return;
    }

    const code = state.currentPlugin.code;

    // 使用现代的 Clipboard API
    if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(code).then(() => {
            showToast('Go代码已复制到剪贴板', 'success');
        }).catch(err => {
            console.error('复制失败:', err);
            fallbackCopyTextToClipboard(code);
        });
    } else {
        // 回退到传统方法
        fallbackCopyTextToClipboard(code);
    }
}

// 回退的复制方法
function fallbackCopyTextToClipboard(text) {
    const textArea = document.createElement('textarea');
    textArea.value = text;
    textArea.style.position = 'fixed';
    textArea.style.left = '-999999px';
    textArea.style.top = '-999999px';
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();

    try {
        document.execCommand('copy');
        showToast('Go代码已复制到剪贴板', 'success');
    } catch (err) {
        console.error('复制失败:', err);
        showToast('复制失败，请手动复制', 'error');
    }

    document.body.removeChild(textArea);
}

// 生成输入表单
function generateInputForm(plugin) {
    if (!elements.inputForm || !plugin?.input) return;

    const fields = [];

    // 收集所有类型的输入字段
    if (plugin.input.textList && plugin.input.textList.length > 0) {
        plugin.input.textList.forEach(field => {
            // field现在是InputItem对象
            const fieldName = field.content || field.name || '未命名字段';
            const fieldType = getDisplayTypeName(field.contentType);
            fields.push(createFormField(field, fieldName, fieldType));
        });
    }

    if (plugin.input.imageList && plugin.input.imageList.length > 0) {
        plugin.input.imageList.forEach(field => {
            const fieldName = field.content || field.name || '未命名字段';
            const fieldType = getDisplayTypeName(field.contentType);
            fields.push(createFormField(field, fieldName, fieldType));
        });
    }

    if (plugin.input.fileList && plugin.input.fileList.length > 0) {
        plugin.input.fileList.forEach(field => {
            const fieldName = field.content || field.name || '未命名字段';
            const fieldType = getDisplayTypeName(field.contentType);
            fields.push(createFormField(field, fieldName, fieldType));
        });
    }

    if (plugin.input.audioList && plugin.input.audioList.length > 0) {
        plugin.input.audioList.forEach(field => {
            const fieldName = field.content || field.name || '未命名字段';
            const fieldType = getDisplayTypeName(field.contentType);
            fields.push(createFormField(field, fieldName, fieldType));
        });
    }

    if (plugin.input.videoList && plugin.input.videoList.length > 0) {
        plugin.input.videoList.forEach(field => {
            const fieldName = field.content || field.name || '未命名字段';
            const fieldType = getDisplayTypeName(field.contentType);
            fields.push(createFormField(field, fieldName, fieldType));
        });
    }

    if (plugin.input.documentList && plugin.input.documentList.length > 0) {
        plugin.input.documentList.forEach(field => {
            const fieldName = field.content || field.name || '未命名字段';
            const fieldType = getDisplayTypeName(field.contentType);
            fields.push(createFormField(field, fieldName, fieldType));
        });
    }

    if (plugin.input.otherList && plugin.input.otherList.length > 0) {
        plugin.input.otherList.forEach(field => {
            const fieldName = field.content || field.name || '未命名字段';
            const fieldType = getDisplayTypeName(field.contentType);
            fields.push(createFormField(field, fieldName, fieldType));
        });
    }

    // 渲染表单字段
    elements.inputForm.innerHTML = fields.join('');
    
    // 绑定文件拖拽事件
    bindFileDropEvents();
}

// 规范化选项格式，支持两种格式：
// 1. 对象数组: [{label: "编码", value: "encode"}]
// 2. 字符串数组: ["选项1", "选项2"] (向后兼容)
function normalizeOptions(options) {
    if (!options || options.length === 0) return [];
    
    // 检查第一个元素的类型
    if (typeof options[0] === 'object' && options[0].label !== undefined && options[0].value !== undefined) {
        // 已经是对象数组格式
        return options;
    } else if (typeof options[0] === 'string') {
        // 字符串数组，转换为对象数组（label和value相同）
        return options.map(opt => ({
            label: opt,
            value: opt
        }));
    }
    return [];
}

// 创建表单字段HTML
function createFormField(field, fieldName, fieldType) {
    const inputId = `field-${fieldName.replace(/[^a-zA-Z0-9]/g, '_')}`;
    const placeholder = getFieldPlaceholder(field.description, fieldType);

    // 将fieldType转换为标准格式的类型名
    const standardType = getStandardTypeName(fieldType);

    // 根据contentType决定UI组件类型
    const contentType = field.contentType || 'text';
    
    // 确定数据类型：只有多媒体类型才使用对应的List，其他都用textList
    const dataType = getDataTypeByContentType(contentType, standardType);
    
    // 规范化options格式
    const normalizedOptions = normalizeOptions(field.options);

    // boolean类型：开关/复选框
    if (contentType === 'boolean') {
        return `
            <div class="form-field">
                <label class="field-label checkbox-label">
                    <input
                        type="checkbox"
                        class="field-checkbox"
                        id="${inputId}"
                        name="${fieldName}"
                        data-type="${dataType}"
                        data-content-type="boolean"
                        value="true"
                    >
                    <span class="checkmark"></span>
                    ${escapeHtml(fieldName)}
                    <span class="field-type-badge">${fieldType}</span>
                    ${field.description ? `
                        <span class="field-description-icon">
                            ℹ
                            <div class="tooltip">${escapeHtml(field.description)}</div>
                        </span>
                    ` : ''}
                </label>
            </div>
        `;
    }

    // select类型：下拉选择框
    if (contentType === 'select' && normalizedOptions.length > 0) {
        const optionsHtml = normalizedOptions.map(option =>
            `<option value="${escapeHtml(option.value)}">${escapeHtml(option.label)}</option>`
        ).join('');

        return `
            <div class="form-field">
                <label class="field-label" for="${inputId}">
                    ${escapeHtml(fieldName)}
                    <span class="field-type-badge">${fieldType}</span>
                    ${field.description ? `
                        <span class="field-description-icon">
                            ℹ
                            <div class="tooltip">${escapeHtml(field.description)}</div>
                        </span>
                    ` : ''}
                </label>
                <select
                    class="field-select"
                    id="${inputId}"
                    name="${fieldName}"
                    data-type="${dataType}"
                    data-content-type="select"
                >
                    <option value="">请选择...</option>
                    ${optionsHtml}
                </select>
            </div>
        `;
    }

    // radio类型：单选按钮组
    if (contentType === 'radio' && normalizedOptions.length > 0) {
        const radioHtml = normalizedOptions.map((option, index) =>
            `<label class="radio-option">
                <input
                    type="radio"
                    name="${fieldName}"
                    value="${escapeHtml(option.value)}"
                    data-type="${dataType}"
                    data-content-type="radio"
                    id="${inputId}_${index}"
                >
                <span class="radio-checkmark"></span>
                ${escapeHtml(option.label)}
            </label>`
        ).join('');

        return `
            <div class="form-field">
                <div class="field-label">
                    ${escapeHtml(fieldName)}
                    <span class="field-type-badge">${fieldType}</span>
                    ${field.description ? `
                        <span class="field-description-icon">
                            ℹ
                            <div class="tooltip">${escapeHtml(field.description)}</div>
                        </span>
                    ` : ''}
                </div>
                <div class="radio-group">
                    ${radioHtml}
                </div>
            </div>
        `;
    }

    // checkbox类型：多选复选框组
    if (contentType === 'checkbox' && normalizedOptions.length > 0) {
        const checkboxHtml = normalizedOptions.map((option, index) =>
            `<label class="checkbox-option">
                <input
                    type="checkbox"
                    name="${fieldName}[]"
                    value="${escapeHtml(option.value)}"
                    data-type="${dataType}"
                    data-content-type="checkbox"
                    id="${inputId}_${index}"
                >
                <span class="checkbox-checkmark"></span>
                ${escapeHtml(option.label)}
            </label>`
        ).join('');

        return `
            <div class="form-field">
                <div class="field-label">
                    ${escapeHtml(fieldName)}
                    <span class="field-type-badge">${fieldType}</span>
                    ${field.description ? `
                        <span class="field-description-icon">
                            ℹ
                            <div class="tooltip">${escapeHtml(field.description)}</div>
                        </span>
                    ` : ''}
                </div>
                <div class="checkbox-group">
                    ${checkboxHtml}
                </div>
            </div>
        `;
    }

    // 文件类型：拖拽上传区域
    if (contentType === 'file' || contentType === 'image' || contentType === 'document' || 
        contentType === 'audio' || contentType === 'video') {
        const acceptAttr = getAcceptAttribute(contentType);
        const fileTypeHint = getFileTypeHint(contentType);
        
        return `
            <div class="form-field">
                <label class="field-label" for="${inputId}">
                    ${escapeHtml(fieldName)}
                    <span class="field-type-badge">${fieldType}</span>
                    ${field.description ? `
                        <span class="field-description-icon">
                            ℹ
                            <div class="tooltip">${escapeHtml(field.description)}</div>
                        </span>
                    ` : ''}
                </label>
                
                <!-- 文件拖拽区域 -->
                <div class="file-drop-zone" 
                     id="${inputId}_drop_zone"
                     data-type="${dataType}"
                     data-content-type="${contentType}"
                     data-field-name="${escapeHtml(fieldName)}">
                    <input type="file" 
                           id="${inputId}" 
                           name="${fieldName}"
                           class="file-input-hidden"
                           ${acceptAttr}
                    >
                    <div class="drop-zone-content">
                        <span class="drop-zone-icon">📁</span>
                        <p class="drop-zone-text">拖拽文件到此处或点击选择</p>
                        <p class="drop-zone-hint">${fileTypeHint}</p>
                    </div>
                    <div class="file-preview" style="display:none;">
                        <span class="file-name"></span>
                        <span class="file-size"></span>
                        <button type="button" class="file-remove-btn">×</button>
                    </div>
                </div>
            </div>
        `;
    }

    // 其他类型：默认文本输入框
    let inputType = 'text';
    if (contentType === 'number') {
        inputType = 'number';
    } else if (contentType === 'email') {
        inputType = 'email';
    } else if (contentType === 'url') {
        inputType = 'url';
    } else if (contentType === 'date') {
        inputType = 'date';
    } else if (contentType === 'time') {
        inputType = 'time';
    } else if (contentType === 'dateTime') {
        inputType = 'datetime-local';
    }

    return `
        <div class="form-field">
            <label class="field-label" for="${inputId}">
                ${escapeHtml(fieldName)}
                <span class="field-type-badge">${fieldType}</span>
                ${field.description ? `
                    <span class="field-description-icon">
                        ℹ
                        <div class="tooltip">${escapeHtml(field.description)}</div>
                    </span>
                ` : ''}
            </label>
            <input
                type="${inputType}"
                class="field-input"
                id="${inputId}"
                name="${fieldName}"
                data-type="${dataType}"
                data-content-type="${contentType}"
                placeholder="${placeholder}"
            >
        </div>
    `;
}

// 获取accept属性
function getAcceptAttribute(contentType) {
    const acceptMap = {
        'image': 'accept="image/*"',
        'audio': 'accept="audio/*"',
        'video': 'accept="video/*"',
        'document': 'accept=".pdf,.doc,.docx,.txt,.md"',
        'file': '' // 接受所有文件
    };
    return acceptMap[contentType] || '';
}

// 获取文件类型提示
function getFileTypeHint(contentType) {
    const hintMap = {
        'image': '支持 JPG, PNG, GIF 等图片格式',
        'audio': '支持 MP3, WAV, OGG 等音频格式',
        'video': '支持 MP4, AVI, MOV 等视频格式',
        'document': '支持 PDF, Word, TXT 等文档格式',
        'file': '支持所有文件类型'
    };
    return hintMap[contentType] || '支持所有文件类型';
}

// 将显示类型转换为标准类型名
function getStandardTypeName(displayType) {
    switch(displayType) {
        case '文本': return 'textList';
        case '图片': return 'imageList';
        case '文件': return 'fileList';
        case '音频': return 'audioList';
        case '视频': return 'videoList';
        case '文档': return 'documentList';
        default: return 'otherList';
    }
}

// 根据contentType获取显示类型名称
function getDisplayTypeName(contentType) {
    switch(contentType) {
        case 'text': return '文本';
        case 'number': return '数字';
        case 'image': return '图片';
        case 'file': return '文件';
        case 'audio': return '音频';
        case 'video': return '视频';
        case 'document': return '文档';
        case 'email': return '邮箱';
        case 'url': return '网址';
        case 'phone': return '电话';
        case 'date': return '日期';
        case 'time': return '时间';
        case 'dateTime': return '日期时间';
        case 'boolean': return '布尔';
        case 'select': return '选择';
        case 'radio': return '单选';
        case 'checkbox': return '多选';
        case 'textarea': return '多行文本';
        default: return '其他';
    }
}

// 根据contentType和standardType确定数据应该放在哪个List中
// 只有多媒体类型（image/audio/video/file/document）使用对应的List
// 其他所有类型（text/number/boolean/select/radio/checkbox等）都放在textList中
function getDataTypeByContentType(contentType, standardType) {
    switch(contentType) {
        case 'image': return 'imageList';
        case 'audio': return 'audioList';
        case 'video': return 'videoList';
        case 'file': return 'fileList';
        case 'document': return 'documentList';
        // 所有非多媒体类型都放在textList中
        case 'text':
        case 'number':
        case 'boolean':
        case 'select':
        case 'radio':
        case 'checkbox':
        case 'email':
        case 'url':
        case 'phone':
        case 'date':
        case 'time':
        case 'dateTime':
        case 'textarea':
        default:
            return 'textList';
    }
}

// 获取字段占位符文本
function getFieldPlaceholder(fieldName, fieldType) {
    return `请输入${fieldName}`;
}

// 显示表单模式
function showInputForm() {
    if (elements.inputFormContainer) {
        elements.inputFormContainer.style.display = 'block';
    }
    if (elements.jsonInputContainer) {
        elements.jsonInputContainer.style.display = 'none';
    }
}

// 显示JSON编辑模式
function showJsonInput() {
    if (elements.inputFormContainer) {
        elements.inputFormContainer.style.display = 'none';
    }
    if (elements.jsonInputContainer) {
        elements.jsonInputContainer.style.display = 'block';
    }
}

// 切换输入模式
async function toggleInputMode() {
    const isFormVisible = elements.inputFormContainer?.style.display !== 'none';

    if (isFormVisible) {
        // 从表单切换到JSON模式
        const formData = await collectFormData();
        if (elements.detailInput && formData) {
            elements.detailInput.value = JSON.stringify(formData, null, 2);
        }
        showJsonInput();
    } else {
        // 从JSON切换到表单模式
        try {
            if (elements.detailInput) {
                const jsonData = JSON.parse(elements.detailInput.value);
                populateFormWithData(jsonData);
            }
        } catch (e) {
            console.warn('JSON解析失败，使用空表单');
        }
        showInputForm();
    }
}

// 从表单收集数据（按标准格式组装）
async function collectFormData() {
    if (!elements.inputForm) return null;

    const categorizedData = {
        textList: [],
        imageList: [],
        fileList: [],
        audioList: [],
        videoList: [],
        documentList: [],
        otherList: []
    };

    // 处理文件输入（拖拽区域）
    const fileDropZones = elements.inputForm.querySelectorAll('.file-drop-zone');
    for (const dropZone of fileDropZones) {
        const dataType = dropZone.getAttribute('data-type');
        const file = dropZone._fileData;

        if (file) {
            try {
                // 统一上传到临时目录，使用路径方式传递
                const base64Content = await fileToBase64(file);
                const filePath = await UploadTempFile(file.name, base64Content);
                
                if (categorizedData[dataType]) {
                    categorizedData[dataType].push(filePath);
                }
                
                console.log(`文件已上传: ${file.name} -> ${filePath}`);
            } catch (error) {
                console.error('文件处理失败:', error);
                showToast(`文件上传失败: ${error.message || error}`, 'error');
                return null;
            }
        }
    }

    // 处理所有类型的输入控件（排除隐藏的文件输入框）
    const allInputs = elements.inputForm.querySelectorAll('input:not(.file-input-hidden), select, textarea');

    allInputs.forEach(input => {
        const dataType = input.getAttribute('data-type') || 'textList';
        const contentType = input.getAttribute('data-content-type') || 'text';
        const fieldName = input.name;

        // 跳过没有name属性的输入控件
        if (!fieldName) return;

        let value = null;

        // 根据contentType处理不同类型的输入
        switch (contentType) {
            case 'boolean':
                // checkbox的boolean类型
                if (input.type === 'checkbox') {
                    value = input.checked ? 'true' : 'false';
                }
                break;

            case 'checkbox':
                // 多选checkbox组，只有选中的才收集
                if (input.type === 'checkbox' && input.checked) {
                    value = input.value;
                }
                break;

            case 'radio':
                // 单选radio，只有选中的才收集
                if (input.type === 'radio' && input.checked) {
                    value = input.value;
                }
                break;

            case 'select':
            case 'number':
            case 'email':
            case 'url':
            case 'date':
            case 'time':
            case 'dateTime':
                // 这些类型只要有值就收集
                if (input.value && input.value.trim()) {
                    value = input.value.trim();
                }
                break;

            default:
                // 文本类型和其他类型
                if (input.value && input.value.trim()) {
                    value = input.value.trim();
                }
                break;
        }

        // 如果有值，添加到对应的分类中
        if (value !== null && categorizedData[dataType]) {
            categorizedData[dataType].push(String(value));
        }
    });

    // 构造标准格式的输出 - 所有字段都要包含，即使为空数组
    const result = {
        textList: categorizedData.textList,
        imageList: categorizedData.imageList,
        fileList: categorizedData.fileList,
        audioList: categorizedData.audioList,
        videoList: categorizedData.videoList,
        documentList: categorizedData.documentList,
        otherList: categorizedData.otherList
    };

    return result;
}

// 根据字段名称判断类型
function getFieldTypeFromName(filedType) {
    const name = filedType.toLowerCase();

    if (name.includes('image')) {
        return 'imageList';
    } else if (name.includes('file')) {
        return 'fileList';
    } else if (name.includes('audio')) {
        return 'audioList';
    } else if (name.includes('video')) {
        return 'videoList';
    } else if (name.includes('document')) {
        return 'documentList';
    } else {
        return 'textList'; // 默认归类为文本类型
    }
}

// 用数据填充表单
function populateFormWithData(data) {
    if (!elements.inputForm || !data) return;

    // 处理所有类型的输入控件
    const allInputs = elements.inputForm.querySelectorAll('input, select, textarea');
    
    // 用于跟踪每个字段类型已使用的值索引
    const fieldIndexes = {};

    allInputs.forEach(input => {
        const dataType = input.getAttribute('data-type') || 'textList';
        const contentType = input.getAttribute('data-content-type') || 'text';
        const fieldName = input.name;

        if (!fieldName || !data[dataType] || !Array.isArray(data[dataType])) return;

        // 初始化该类型的索引
        if (!fieldIndexes[dataType]) {
            fieldIndexes[dataType] = 0;
        }

        const valueIndex = fieldIndexes[dataType];
        
        if (valueIndex >= data[dataType].length) return;

        const value = String(data[dataType][valueIndex]);

        // 根据contentType设置不同类型的控件值
        switch (contentType) {
            case 'boolean':
                // checkbox的boolean类型
                if (input.type === 'checkbox') {
                    input.checked = (value === 'true' || value === '1');
                    fieldIndexes[dataType]++;
                }
                break;

            case 'checkbox':
                // 多选checkbox组
                if (input.type === 'checkbox') {
                    // 检查当前值是否匹配任何一个选中的值
                    if (data[dataType].includes(input.value)) {
                        input.checked = true;
                    }
                }
                break;

            case 'radio':
                // 单选radio
                if (input.type === 'radio' && input.value === value) {
                    input.checked = true;
                    fieldIndexes[dataType]++;
                }
                break;

            case 'select':
                // 下拉选择框
                if (input.tagName.toLowerCase() === 'select') {
                    input.value = value;
                    fieldIndexes[dataType]++;
                }
                break;

            default:
                // 文本、数字等其他类型
                input.value = value;
                fieldIndexes[dataType]++;
                break;
        }
    });
}

// 获取字段在同类型中的索引位置
function getFieldIndexInType(input, fieldType) {
    const inputs = elements.inputForm.querySelectorAll(`.field-input[data-type="${fieldType}"]`);
    let index = 0;
    for (const inp of inputs) {
        if (inp === input) {
            return index;
        }
        index++;
    }
    return 0;
}

// Tooltip事件绑定
function bindTooltipEvents() {
    // 使用事件委托来处理所有tooltip图标的悬停事件
    document.addEventListener('mouseenter', (e) => {
        if (e.target.classList.contains('field-description-icon')) {
            const tooltip = e.target.querySelector('.tooltip');
            if (tooltip) {
                tooltip.classList.add('show');
                // 调整位置以确保tooltip在视窗内
                adjustTooltipPosition(e.target, tooltip);
            }
        }
    }, true);

    document.addEventListener('mouseleave', (e) => {
        if (e.target.classList.contains('field-description-icon')) {
            const tooltip = e.target.querySelector('.tooltip');
            if (tooltip) {
                tooltip.classList.remove('show');
            }
        }
    }, true);

    // 添加键盘支持
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            // 隐藏所有显示中的tooltip
            document.querySelectorAll('.tooltip.show').forEach(tooltip => {
                tooltip.classList.remove('show');
            });
        }
    });
}

// 调整tooltip位置以确保在视窗内
function adjustTooltipPosition(icon, tooltip) {
    const rect = icon.getBoundingClientRect();
    const viewportHeight = window.innerHeight;
    const tooltipHeight = 100; // 估算的tooltip高度（稍微减少因为字体变小了）

    // 重置样式
    tooltip.style.transform = 'translateX(-50%)';
    tooltip.style.bottom = '125%';
    tooltip.style.top = 'auto';
    tooltip.style.left = '50%';
    icon.classList.remove('tooltip-bottom');

    // 检查是否超出上边界
    if (rect.top - tooltipHeight < 0) {
        // 如果超出上边界，将tooltip放在图标下方
        tooltip.style.bottom = 'auto';
        tooltip.style.top = '125%';
        icon.classList.add('tooltip-bottom');
    }

    // 水平位置调整 - 先获取tooltip的实际宽度
    const tooltipRect = tooltip.getBoundingClientRect();
    const viewportWidth = window.innerWidth;

    // 检查是否超出右边界
    if (rect.left + tooltipRect.width / 2 > viewportWidth) {
        tooltip.style.left = 'auto';
        tooltip.style.right = '0';
        tooltip.style.transform = 'translateX(0)';
    }
    // 检查是否超出左边界
    else if (rect.left - tooltipRect.width / 2 < 0) {
        tooltip.style.left = '0';
        tooltip.style.transform = 'translateX(0)';
    }
}


// 绑定复制字段事件
function bindCopyFieldEvents() {
    // 移除之前的事件监听器
    if (elements.detailOutput) {
        elements.detailOutput.removeEventListener('click', handleFieldCopyClick);
        // 重新添加事件监听器（使用事件委托处理冒泡）
        elements.detailOutput.addEventListener('click', handleFieldCopyClick);
    }
}

// 加载本地图片
async function loadLocalImages() {
    const imageContainers = document.querySelectorAll('.result-image-container');
    
    for (const container of imageContainers) {
        const filePath = container.dataset.filePath;
        if (!filePath) continue;
        
        try {
            // 调用后端API获取图片的Base64数据
            const { GetOutputFileBase64 } = window.go.app.App;
            const base64Data = await GetOutputFileBase64(filePath);
            
            // 创建图片元素并替换加载提示
            const img = document.createElement('img');
            img.src = base64Data;
            img.alt = '输出图片';
            img.className = 'result-image';
            
            container.innerHTML = '';
            container.appendChild(img);
        } catch (error) {
            console.error('加载图片失败:', error);
            container.innerHTML = '<div class="error-message">图片加载失败</div>';
        }
    }
}

// 绑定文件操作事件
function bindFileActionEvents() {
    // 绑定下载按钮事件
    const downloadButtons = document.querySelectorAll('.btn-download-file');
    downloadButtons.forEach(button => {
        button.addEventListener('click', handleDownloadFile);
    });
    
    // 绑定复制文件按钮事件
    const copyButtons = document.querySelectorAll('.btn-copy-file');
    copyButtons.forEach(button => {
        button.addEventListener('click', handleCopyFileToLocation);
    });
}

// 处理文件下载
async function handleDownloadFile(event) {
    const button = event.target;
    const filePath = button.dataset.filePath;
    
    if (!filePath) {
        showToast('无效的文件路径', 'error');
        return;
    }
    
    try {
        // 调用后端API打开文件保存对话框，选择下载位置
        const { SaveFileDialog } = window.go.app.App;
        if (!SaveFileDialog) {
            showToast('文件下载功能不可用', 'error');
            return;
        }
        
        const fileName = filePath.split('/').pop() || 'download';
        const targetPath = await SaveFileDialog(fileName, '选择下载位置');
        
        if (!targetPath) {
            // 用户取消了选择
            return;
        }
        
        // 调用后端API复制文件到下载位置
        const { CopyOutputFile } = window.go.app.App;
        await CopyOutputFile(filePath, targetPath);
        
        showToast('文件已下载到: ' + targetPath, 'success');
    } catch (error) {
        console.error('下载文件失败:', error);
        showToast('下载文件失败: ' + error.message, 'error');
    }
}

// 处理复制文件到指定位置
async function handleCopyFileToLocation(event) {
    const button = event.target;
    const sourcePath = button.dataset.filePath;
    
    if (!sourcePath) {
        showToast('无效的文件路径', 'error');
        return;
    }
    
    try {
        // 调用后端API打开文件保存对话框
        const { SaveFileDialog } = window.go.app.App;
        if (!SaveFileDialog) {
            showToast('文件保存功能不可用', 'error');
            return;
        }
        
        const fileName = sourcePath.split('/').pop() || 'file';
        const targetPath = await SaveFileDialog(fileName, '选择保存位置');
        
        if (!targetPath) {
            // 用户取消了选择
            return;
        }
        
        // 调用后端API复制文件
        const { CopyOutputFile } = window.go.app.App;
        await CopyOutputFile(sourcePath, targetPath);
        
        showToast('文件已复制到: ' + targetPath, 'success');
    } catch (error) {
        console.error('复制文件失败:', error);
        showToast('复制文件失败: ' + error.message, 'error');
    }
}

// 处理字段复制按钮点击事件
function handleFieldCopyClick(event) {
    // 使用 closest() 方法找到最近的复制按钮祖先元素
    // 这样可以处理事件冒泡，无论点击图标、按钮空白区域还是按钮本身都能工作
    const button = event.target.closest('.copy-field-btn');

    if (!button) {
        return; // 不是复制按钮，忽略
    }

    const fieldIndex = parseInt(button.dataset.fieldIndex);

    if (isNaN(fieldIndex)) {
        showToast('复制失败：无效的字段索引', 'error');
        return;
    }

    const resultContainer = elements.detailOutput?.querySelector('.execution-result');
    if (!resultContainer) {
        showToast('复制失败：找不到结果容器', 'error');
        return;
    }

    const resultItem = resultContainer.querySelector(`.result-item[data-field-index="${fieldIndex}"]`);
    if (!resultItem) {
        showToast('复制失败：找不到对应的结果项', 'error');
        return;
    }

    // 从结果项中提取文本内容
    const fieldLabel = resultItem.querySelector('.result-label')?.textContent || '';
    const fieldValueElement = resultItem.querySelector('.result-value');

    // 提取纯文本内容，处理可能包含HTML的情况
    let fieldValue = '';
    if (fieldValueElement) {
        // 如果是文本节点，直接获取textContent
        if (fieldValueElement.textContent) {
            fieldValue = fieldValueElement.textContent;
        } else {
            // 如果包含HTML，提取可见文本
            fieldValue = fieldValueElement.innerText || fieldValueElement.textContent || '';
        }
    }

    const fieldDescription = resultItem.querySelector('.result-description')?.textContent || '';

    // 构建要复制的文本
    let textToCopy = '';

    if (fieldLabel) {
        textToCopy += `${fieldLabel}`;
        if (fieldValue) {
            textToCopy += `: ${fieldValue}`;
        }
    } else {
        textToCopy = fieldValue;
    }
    // 清理文本内容
    textToCopy = textToCopy.trim();

    if (!textToCopy) {
        showToast('复制失败：没有可复制的内容', 'error');
        return;
    }

    // 使用现代的 Clipboard API
    if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(textToCopy).then(() => {
            showToast(`${fieldLabel || '结果'} 已复制到剪贴板`, 'success');
        }).catch(err => {
            console.error('复制失败:', err);
            fallbackCopyTextToClipboard(textToCopy);
        });
    } else {
        // 回退到传统方法
        fallbackCopyTextToClipboard(textToCopy);
    }
}

// 显示插件标签和工具依赖
function displayPluginLabels(plugin) {
    // 移除现有的标签区域
    const existingLabels = document.querySelector('.plugin-labels-display');
    if (existingLabels) {
        existingLabels.remove();
    }

    // 检查是否有标签或相关插件需要显示
    const hasTags = plugin.tags && plugin.tags.length > 0;
    const hasRelatedPlugins = plugin.relatedPlugins && plugin.relatedPlugins.length > 0;

    if (!hasTags && !hasRelatedPlugins) {
        return;
    }

    // 创建标签容器
    const labelsContainer = document.createElement('div');
    labelsContainer.className = 'plugin-labels-display';

    // 添加标签
    if (hasTags) {
        plugin.tags.forEach(tag => {
            const labelElement = document.createElement('span');
            labelElement.className = 'plugin-label';
            labelElement.style.background = tag.color || '#667eea';
            labelElement.innerHTML = `<span class="plugin-label-color" style="background-color: ${tag.color || '#ccc'}"></span>${tag.name}`;
            labelsContainer.appendChild(labelElement);
        });
    }

    // 添加相关插件关联列表
    let hasMissingDependencies = false;
    if (plugin.relatedPlugins && plugin.relatedPlugins.length > 0) {
        const toolsList = document.createElement('div');
        toolsList.className = 'plugin-related-tools-list';

        const title = document.createElement('h5');
        title.textContent = '相关工具';
        toolsList.appendChild(title);

        plugin.relatedPlugins.forEach(relatedPlugin => {
            // 根据相关插件ID找到对应的插件信息
            const relatedPluginInfo = state.plugins.find(p => p.id === relatedPlugin.relatedPluginId);

            const toolItem = document.createElement('div');
            toolItem.className = 'plugin-tool-item';

            const toolIcon = document.createElement('span');
            const toolInfo = document.createElement('div');

            if (relatedPluginInfo) {
                // 依赖工具存在
                toolIcon.textContent = '🔗';
                toolItem.classList.add('dependency-exists');

                const toolName = document.createElement('span');
                toolName.className = 'plugin-tool-name';

                // 显示中文名和英文名
                const chineseName = relatedPluginInfo.chineseName || '';
                const englishName = relatedPluginInfo.name || '';

                if (chineseName && englishName && chineseName !== englishName) {
                    toolName.innerHTML = `${chineseName} <span class="plugin-tool-english">(${englishName})</span>`;
                } else {
                    toolName.textContent = chineseName || englishName || relatedPlugin.relatedPluginName;
                }

                toolInfo.appendChild(toolName);
            } else {
                // 依赖工具不存在
                hasMissingDependencies = true;
                toolIcon.textContent = '⚠️';
                toolItem.classList.add('dependency-missing');

                const toolName = document.createElement('span');
                toolName.className = 'plugin-tool-name';
                toolName.innerHTML = `<span class="plugin-tool-error">${relatedPlugin.relatedPluginName || '未知工具'}</span> <span class="plugin-tool-status">(工具不存在)</span>`;

                toolInfo.appendChild(toolName);
            }

            toolItem.appendChild(toolIcon);
            toolItem.appendChild(toolInfo);
            toolsList.appendChild(toolItem);
        });

        labelsContainer.appendChild(toolsList);

        // 如果有缺失的依赖，禁用执行按钮
        if (hasMissingDependencies && elements.detailExecuteBtn) {
            elements.detailExecuteBtn.disabled = true;
            elements.detailExecuteBtn.textContent = '⚠️ 依赖工具缺失';
            elements.detailExecuteBtn.style.opacity = '0.6';
        }
    }

    // 如果有标签，插入到详情描述之后
    if (labelsContainer.children.length > 0 && elements.detailDescription) {
        elements.detailDescription.parentNode.insertBefore(labelsContainer, elements.detailDescription.nextSibling);
    }
}



// 显示插件依赖信息
function displayPluginDependencies(plugin) {
    if (!elements.pluginDependencies || !elements.dependenciesList || !elements.dependenciesWarning) {
        return;
    }

    // 如果没有相关插件关联，隐藏依赖区域
    if (!plugin.relatedPlugins || plugin.relatedPlugins.length === 0) {
        elements.pluginDependencies.style.display = 'none';
        return;
    }

    // 显示依赖区域
    elements.pluginDependencies.style.display = 'block';

    // 清空现有内容
    elements.dependenciesList.innerHTML = '';
    elements.dependenciesWarning.style.display = 'none';

    let hasMissingDependencies = false;
    const dependencyItems = [];

    // 检查每个依赖的插件是否存在
    plugin.relatedPlugins.forEach(relatedPlugin => {
        // 查找依赖的插件是否存在
        const dependentPlugin = state.plugins.find(p => p.id === relatedPlugin.relatedPluginId);

        const isMissing = !dependentPlugin;
        if (isMissing) {
            hasMissingDependencies = true;
        }

        const dependencyItem = document.createElement('div');
        dependencyItem.className = `dependency-item ${isMissing ? 'missing' : ''}`;

        const chineseName = dependentPlugin?.chineseName || '未知工具';
        const englishName = dependentPlugin?.name || relatedPlugin.relatedPluginName || 'Unknown Tool';

        dependencyItem.innerHTML = `
            <span class="dependency-icon">${isMissing ? '❌' : '✅'}</span>
            <div class="dependency-info">
                <div class="dependency-name">${escapeHtml(chineseName)}</div>
                <div class="dependency-english">${escapeHtml(englishName)}</div>
                <div class="dependency-desc">${isMissing ? '工具不存在或已删除' : '工具正常可用'}</div>
            </div>
        `;

        dependencyItems.push(dependencyItem);
    });

    // 添加依赖项到列表
    dependencyItems.forEach(item => {
        elements.dependenciesList.appendChild(item);
    });

    // 显示警告信息
    if (hasMissingDependencies) {
        elements.dependenciesWarning.style.display = 'flex';

        // 禁用执行按钮
        if (elements.detailExecuteBtn) {
            elements.detailExecuteBtn.disabled = true;
            elements.detailExecuteBtn.style.opacity = '0.5';
            elements.detailExecuteBtn.style.cursor = 'not-allowed';
            elements.detailExecuteBtn.textContent = '依赖缺失，无法执行';
        }
    } else {
        // 启用执行按钮
        if (elements.detailExecuteBtn) {
            elements.detailExecuteBtn.disabled = false;
            elements.detailExecuteBtn.style.opacity = '1';
            elements.detailExecuteBtn.style.cursor = 'pointer';
            elements.detailExecuteBtn.innerHTML = '<span class="loading-spinner" style="display: none;"></span><span class="btn-text">▶ 执行</span>';
        }
    }
}

// formatCode函数已移除，改用纯文本显示

// ============= 文件拖拽上传相关函数 =============

// 绑定文件拖拽事件
function bindFileDropEvents() {
    const dropZones = elements.inputForm?.querySelectorAll('.file-drop-zone');
    if (!dropZones) return;

    dropZones.forEach(dropZone => {
        const fileInput = dropZone.querySelector('.file-input-hidden');
        const dropContent = dropZone.querySelector('.drop-zone-content');
        const filePreview = dropZone.querySelector('.file-preview');
        const fileName = dropZone.querySelector('.file-name');
        const fileSize = dropZone.querySelector('.file-size');
        const removeBtn = dropZone.querySelector('.file-remove-btn');

        // 点击区域触发文件选择
        dropZone.addEventListener('click', (e) => {
            if (!e.target.classList.contains('file-remove-btn')) {
                fileInput.click();
            }
        });

        // 文件选择变化
        fileInput.addEventListener('change', (e) => {
            if (e.target.files && e.target.files.length > 0) {
                handleFileSelect(e.target.files[0], dropZone);
            }
        });

        // 拖拽进入
        dropZone.addEventListener('dragover', (e) => {
            e.preventDefault();
            dropZone.classList.add('drag-over');
        });

        // 拖拽离开
        dropZone.addEventListener('dragleave', (e) => {
            e.preventDefault();
            dropZone.classList.remove('drag-over');
        });

        // 放置文件
        dropZone.addEventListener('drop', (e) => {
            e.preventDefault();
            dropZone.classList.remove('drag-over');
            
            const files = e.dataTransfer.files;
            if (files.length > 0) {
                handleFileSelect(files[0], dropZone);
            }
        });

        // 移除文件
        removeBtn?.addEventListener('click', (e) => {
            e.stopPropagation();
            clearFile(dropZone);
        });
    });
}

// 处理文件选择
function handleFileSelect(file, dropZone) {
    if (!file) return;

    const fileInput = dropZone.querySelector('.file-input-hidden');
    const dropContent = dropZone.querySelector('.drop-zone-content');
    const filePreview = dropZone.querySelector('.file-preview');
    const fileName = dropZone.querySelector('.file-name');
    const fileSize = dropZone.querySelector('.file-size');

    // 显示文件信息
    fileName.textContent = file.name;
    fileSize.textContent = formatFileSize(file.size);
    
    dropContent.style.display = 'none';
    filePreview.style.display = 'flex';

    // 存储文件对象到DOM元素上
    dropZone._fileData = file;

    showToast(`已选择文件: ${file.name}`, 'success');
}

// 清除文件
function clearFile(dropZone) {
    const fileInput = dropZone.querySelector('.file-input-hidden');
    const dropContent = dropZone.querySelector('.drop-zone-content');
    const filePreview = dropZone.querySelector('.file-preview');

    fileInput.value = '';
    dropZone._fileData = null;
    
    dropContent.style.display = 'block';
    filePreview.style.display = 'none';
}

// 格式化文件大小
function formatFileSize(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
}

// 读取文件为Base64
function readFileAsBase64(file) {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => {
            // 返回 data URL 格式: data:image/png;base64,iVBORw0KG...
            resolve(reader.result);
        };
        reader.onerror = reject;
        reader.readAsDataURL(file);
    });
}

// 文件转Base64(用于后端传输，去掉data URL前缀)
function fileToBase64(file) {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => resolve(reader.result.split(',')[1]); // 去掉data URL前缀
        reader.onerror = reject;
        reader.readAsDataURL(file);
    });
}

// ============= 分页相关函数 =============

// 更新分页UI
function updatePaginationUI() {
    if (!elements.paginationContainer) return;
    
    const pagination = state.paginationState;
    const totalPages = pagination.totalPages;
    const currentPage = pagination.currentPage;
    
    // 如果只有一页，隐藏分页控件
    if (totalPages <= 1) {
        elements.paginationContainer.style.display = 'none';
        return;
    }
    
    // 显示分页控件
    elements.paginationContainer.style.display = 'flex';
    
    // 更新分页信息文本
    if (elements.paginationInfo) {
        elements.paginationInfo.textContent = `第 ${currentPage} 页 / 共 ${totalPages} 页`;
    }
    
    // 更新上一页按钮状态
    if (elements.prevPageBtn) {
        elements.prevPageBtn.disabled = currentPage <= 1;
    }
    
    // 更新下一页按钮状态
    if (elements.nextPageBtn) {
        elements.nextPageBtn.disabled = currentPage >= totalPages;
    }
}

// 处理上一页按钮点击
function handlePrevPage() {
    const pagination = state.paginationState;
    if (pagination.currentPage > 1) {
        pagination.currentPage--;
        renderPlugins();
        // 滚动到顶部
        if (elements.toolsList) {
            elements.toolsList.scrollTop = 0;
        }
    }
}

// 处理下一页按钮点击
function handleNextPage() {
    const pagination = state.paginationState;
    if (pagination.currentPage < pagination.totalPages) {
        pagination.currentPage++;
        renderPlugins();
        // 滚动到顶部
        if (elements.toolsList) {
            elements.toolsList.scrollTop = 0;
        }
    }
}

