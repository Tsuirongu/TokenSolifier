package models

import (
	"encoding/json"
	"time"
)

type InputType string

const (
	InputTypeNumber   InputType = "number"
	InputTypeEmail    InputType = "email"
	InputTypeUrl      InputType = "url"
	InputTypePhone    InputType = "phone"
	InputTypeDateTime InputType = "dateTime"
	InputTypeBoolean  InputType = "boolean"
	InputTypeSelect   InputType = "select"
	InputTypeRadio    InputType = "radio"
	InputTypeCheckbox InputType = "checkbox"
	InputTypeText     InputType = "text"
	InputTypeTextarea InputType = "media"
)

// SelectOption 选择项配置，支持显示值和实际值分离
type SelectOption struct {
	Label string `json:"label"` // 显示文本（前端展示）
	Value string `json:"value"` // 实际值（传给插件）
}

type InputItem struct {
	Content     string         `json:"content"`
	ContentType InputType      `json:"contentType"`
	Description string         `json:"description"`
	Options     []SelectOption `json:"options,omitempty"` // 选项列表，支持label+value格式
}

// UnmarshalJSON 自定义JSON反序列化，支持两种格式：
// 1. 简单字符串数组: ["选项1", "选项2"] (向后兼容)
// 2. 对象数组: [{"label": "编码", "value": "encode"}]
func (item *InputItem) UnmarshalJSON(data []byte) error {
	// 定义临时结构体用于解析
	type Alias InputItem
	aux := &struct {
		Options interface{} `json:"options,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(item),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 处理options字段
	if aux.Options != nil {
		switch v := aux.Options.(type) {
		case []interface{}:
			if len(v) > 0 {
				// 判断第一个元素的类型
				switch v[0].(type) {
				case string:
					// 简单字符串数组格式，转换为SelectOption
					item.Options = make([]SelectOption, len(v))
					for i, opt := range v {
						if str, ok := opt.(string); ok {
							item.Options[i] = SelectOption{
								Label: str,
								Value: str, // label和value相同
							}
						}
					}
				case map[string]interface{}:
					// 对象数组格式，直接解析
					optionsJSON, _ := json.Marshal(v)
					json.Unmarshal(optionsJSON, &item.Options)
				}
			}
		}
	}

	return nil
}

// Plugin 插件模型
type Plugin struct {
	ID             int64                  `json:"id"`
	Name           string                 `json:"name"`        // 英文函数名
	ChineseName    string                 `json:"chineseName"` // 中文名
	Description    string                 `json:"description"`
	Code           string                 `json:"code"`
	WasmBinary     []byte                 `json:"-"` // WASM二进制数据，不序列化到JSON
	IsActive       bool                   `json:"isActive"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
	Input          *PluginInputDesc       `json:"input,omitempty"`          // 可选的输入描述，用于向后兼容
	Output         *PluginOutputDesc      `json:"output,omitempty"`         // 可选的输出描述，用于向后兼容
	Dependencies   *[]PluginDependency    `json:"dependencies,omitempty"`   // 依赖信息
	RelatedPlugins *[]PluginRelatedPlugin `json:"relatedPlugins,omitempty"` // 相关插件关联
	Tags           *[]Tag                 `json:"tags,omitempty"`           // 插件标签
}

// PluginInputDesc 仅用于前端展示的输入格式，并非代码的实际输入格式
type PluginInputDesc struct {
	TextList     []*InputItem `json:"textList"`
	ImageList    []*InputItem `json:"imageList"`
	FileList     []*InputItem `json:"fileList"`
	AudioList    []*InputItem `json:"audioList"`
	VideoList    []*InputItem `json:"videoList"`
	DocumentList []*InputItem `json:"documentList"`
	OtherList    []*InputItem `json:"otherList"`
}

// PluginOutputDesc 用于前端展示
type PluginOutputDesc struct {
	TextList     []*InputItem `json:"textList"`
	ImageList    []*InputItem `json:"imageList"`
	FileList     []*InputItem `json:"fileList"`
	AudioList    []*InputItem `json:"audioList"`
	VideoList    []*InputItem `json:"videoList"`
	DocumentList []*InputItem `json:"documentList"`
	OtherList    []*InputItem `json:"otherList"`
	PluginError  string       `json:"pluginError"`
}

// PluginIO 实际传输的输入输出描述
type PluginIO struct {
	TextList     []string `json:"textList"`
	ImageList    []string `json:"imageList"`
	FileList     []string `json:"fileList"`
	AudioList    []string `json:"audioList"`
	VideoList    []string `json:"videoList"`
	DocumentList []string `json:"documentList"`
	OtherList    []string `json:"otherList"`
	PluginError  string   `json:"pluginError"`
}

// PluginInputOutput 插件输入输出配置
type PluginInputOutput struct {
	ID         int64            `json:"id"`
	PluginID   int64            `json:"pluginId"`
	InputDesc  PluginInputDesc  `json:"inputDesc"`
	OutputDesc PluginOutputDesc `json:"outputDesc"`
	CreatedAt  time.Time        `json:"createdAt"`
	UpdatedAt  time.Time        `json:"updatedAt"`
}

// PluginDependency 插件的包依赖 （注意不是关联插件）
type PluginDependency struct {
	ID        int64     `json:"id"`
	PluginID  int64     `json:"pluginId"`
	Package   string    `json:"package"` // 包名，如 "strconv", "strings"
	Version   string    `json:"version"` // 版本信息，可选
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// PluginRelatedPlugin 插件相关插件关联表
type PluginRelatedPlugin struct {
	ID                int64     `json:"id"`
	PluginID          int64     `json:"pluginId"`          // 当前插件ID
	RelatedPluginID   int64     `json:"relatedPluginId"`   // 相关插件ID
	RelatedPluginName string    `json:"relatedPluginName"` // 相关插件英文名
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

// PluginInput 插件输入结构体（用于Yaegi外部函数调用）
type PluginInput struct {
	TextList     []string `json:"textList"`
	ImageList    []string `json:"imageList"`
	FileList     []string `json:"fileList"`
	AudioList    []string `json:"audioList"`
	VideoList    []string `json:"videoList"`
	DocumentList []string `json:"documentList"`
	OtherList    []string `json:"otherList"`
}

// PluginOutput 插件输出结构体（用于Yaegi外部函数调用）
type PluginOutput struct {
	TextList     []string `json:"textList"`
	ImageList    []string `json:"imageList"`
	FileList     []string `json:"fileList"`
	AudioList    []string `json:"audioList"`
	VideoList    []string `json:"videoList"`
	DocumentList []string `json:"documentList"`
	OtherList    []string `json:"otherList"`
	PluginError  string   `json:"pluginError"`
}

// Tag 标签模型
type Tag struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`  // 标签名称
	Color     string    `json:"color"` // 标签颜色（可选）
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// PluginTag 插件标签关联模型
type PluginTag struct {
	ID        int64     `json:"id"`
	PluginID  int64     `json:"pluginId"` // 插件ID
	TagID     int64     `json:"tagId"`    // 标签ID
	TagName   string    `json:"tagName"`  // 标签名称（冗余字段，便于查询）
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// PluginExecutionResult 插件执行结果（已弃用，ExecutePlugin现在直接返回JSON字符串）
// 保留此结构体以供历史参考或未来扩展使用
type PluginExecutionResult struct {
	Result       string              `json:"result"`       // 执行结果JSON字符串
	Dependencies *[]PluginDependency `json:"dependencies"` // 依赖信息
}
