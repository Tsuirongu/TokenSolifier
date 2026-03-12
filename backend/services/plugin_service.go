package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"loji-app/backend/models"
	"strings"
	"time"
)

// PluginService 插件服务
type PluginService struct {
	db *sql.DB
}

// NewPluginService 创建插件服务实例
func NewPluginService(db *sql.DB) *PluginService {
	service := &PluginService{db: db}

	// 执行数据库迁移
	if err := service.migrateDatabase(); err != nil {
		log.Printf("Warning: failed to migrate database: %v", err)
		// 不阻止服务启动，但记录警告
	}

	return service
}

// GetAllPlugins 获取所有插件（包含依赖信息）
func (s *PluginService) GetAllPlugins() ([]models.Plugin, error) {
	// 先获取所有插件的基本信息
	query := `SELECT id, name, chinese_name, description, code, wasm_binary, is_active, created_at, updated_at
			  FROM plugins ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plugins []models.Plugin
	for rows.Next() {
		var plugin models.Plugin
		err := rows.Scan(
			&plugin.ID,
			&plugin.Name,
			&plugin.ChineseName,
			&plugin.Description,
			&plugin.Code,
			&plugin.WasmBinary,
			&plugin.IsActive,
			&plugin.CreatedAt,
			&plugin.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// 获取该插件的依赖信息
		dependencies, err := s.GetPluginDependencies(plugin.ID)
		if err != nil {
			return nil, err
		}
		plugin.Dependencies = &dependencies

		// 获取输入输出配置
		ioConfig, err := s.GetPluginInputOutput(plugin.ID)
		if err != nil {
			return nil, err
		}

		// 如果有输入输出配置，将其设置到插件对象中
		if ioConfig != nil {
			plugin.Input = &ioConfig.InputDesc
			plugin.Output = &ioConfig.OutputDesc
		}

		// 获取相关插件关联信息
		relatedPlugins, err := s.GetPluginRelatedPlugins(plugin.ID)
		if err != nil {
			return nil, err
		}
		plugin.RelatedPlugins = &relatedPlugins

		plugins = append(plugins, plugin)
	}

	return plugins, nil
}

// GetAllActivePlugins 获取所有活跃插件
func (s *PluginService) GetAllActivePlugins() ([]models.Plugin, error) {
	query := `SELECT id, name, chinese_name, description, code, wasm_binary, is_active, created_at, updated_at
			  FROM plugins WHERE is_active = 1 ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plugins []models.Plugin
	for rows.Next() {
		var plugin models.Plugin
		err := rows.Scan(
			&plugin.ID,
			&plugin.Name,
			&plugin.ChineseName,
			&plugin.Description,
			&plugin.Code,
			&plugin.WasmBinary,
			&plugin.IsActive,
			&plugin.CreatedAt,
			&plugin.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		plugins = append(plugins, plugin)
	}

	return plugins, nil
}

// GetPluginByID 根据ID获取插件（包含依赖信息）
func (s *PluginService) GetPluginByID(id int64) (*models.Plugin, error) {
	query := `SELECT id, name, chinese_name, description, code, wasm_binary, is_active, created_at, updated_at
			  FROM plugins WHERE id = ?`

	var plugin models.Plugin
	err := s.db.QueryRow(query, id).Scan(
		&plugin.ID,
		&plugin.Name,
		&plugin.ChineseName,
		&plugin.Description,
		&plugin.Code,
		&plugin.WasmBinary,
		&plugin.IsActive,
		&plugin.CreatedAt,
		&plugin.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("plugin not found")
	}
	if err != nil {
		return nil, err
	}

	// 获取该插件的依赖信息
	dependencies, err := s.GetPluginDependencies(id)
	if err != nil {
		return nil, err
	}
	plugin.Dependencies = &dependencies

	return &plugin, nil
}

// CreatePlugin 创建新插件
func (s *PluginService) CreatePlugin(plugin *models.Plugin) error {
	query := `INSERT INTO plugins (name, chinese_name, description, code, wasm_binary, is_active, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	plugin.CreatedAt = now
	plugin.UpdatedAt = now

	result, err := s.db.Exec(query,
		plugin.Name,
		plugin.ChineseName,
		plugin.Description,
		plugin.Code,
		plugin.WasmBinary,
		plugin.IsActive,
		plugin.CreatedAt,
		plugin.UpdatedAt,
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	plugin.ID = id
	return nil
}

// UpdatePlugin 更新插件
func (s *PluginService) UpdatePlugin(plugin *models.Plugin) error {
	query := `UPDATE plugins
			  SET name = ?, chinese_name = ?, description = ?, code = ?, wasm_binary = ?, is_active = ?, updated_at = ?
			  WHERE id = ?`

	plugin.UpdatedAt = time.Now()

	_, err := s.db.Exec(query,
		plugin.Name,
		plugin.ChineseName,
		plugin.Description,
		plugin.Code,
		plugin.WasmBinary,
		plugin.IsActive,
		plugin.UpdatedAt,
		plugin.ID,
	)

	return err
}

// DeletePlugin 删除插件
func (s *PluginService) DeletePlugin(id int64) error {
	// 先删除依赖配置
	if err := s.DeletePluginDependencies(id); err != nil {
		return err
	}

	// 先删除输入输出配置
	if err := s.DeletePluginInputOutput(id); err != nil {
		return err
	}

	// 再删除插件
	query := `DELETE FROM plugins WHERE id = ?`
	_, err := s.db.Exec(query, id)
	return err
}

// InitDatabase 初始化数据库表
func (s *PluginService) InitDatabase() error {
	// 检查插件表是否存在以及是否缺少chinese_name列
	checkTableSQL := `
		SELECT sql FROM sqlite_master
		WHERE type='table' AND name='plugins';
	`

	var tableSQL string
	err := s.db.QueryRow(checkTableSQL).Scan(&tableSQL)
	if err != nil {
		if err == sql.ErrNoRows {
			// 表不存在，直接创建
			pluginTableSQL := `
			CREATE TABLE plugins (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				chinese_name TEXT,
				description TEXT NOT NULL,
				code TEXT NOT NULL,
				wasm_binary BLOB,
				is_active BOOLEAN NOT NULL DEFAULT 1,
				created_at DATETIME NOT NULL,
				updated_at DATETIME NOT NULL
			)`
			if _, err := s.db.Exec(pluginTableSQL); err != nil {
				return fmt.Errorf("failed to create plugins table: %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to check table structure: %w", err)
	}

	// 检查表是否包含chinese_name列
	hasChineseName := strings.Contains(tableSQL, "chinese_name")

	// 如果表存在但缺少chinese_name列，需要重建表
	if !hasChineseName {
		log.Println("Table plugins exists but missing chinese_name column, recreating...")

		// 备份现有数据
		backupSQL := `CREATE TABLE plugins_backup AS SELECT * FROM plugins;`
		if _, err := s.db.Exec(backupSQL); err != nil {
			return fmt.Errorf("failed to backup plugins table: %w", err)
		}

		// 删除旧表
		dropSQL := `DROP TABLE plugins;`
		if _, err := s.db.Exec(dropSQL); err != nil {
			return fmt.Errorf("failed to drop plugins table: %w", err)
		}

		// 创建新表
		createSQL := `
		CREATE TABLE plugins (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			chinese_name TEXT,
			description TEXT NOT NULL,
			code TEXT NOT NULL,
			wasm_binary BLOB,
			is_active BOOLEAN NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`
		if _, err := s.db.Exec(createSQL); err != nil {
			return fmt.Errorf("failed to create new plugins table: %w", err)
		}

		// 恢复数据
		restoreSQL := `
		INSERT INTO plugins (id, name, description, code, wasm_binary, is_active, created_at, updated_at)
		SELECT id, name, description, code, wasm_binary, is_active, created_at, updated_at
		FROM plugins_backup;
		`
		if _, err := s.db.Exec(restoreSQL); err != nil {
			return fmt.Errorf("failed to restore plugins data: %w", err)
		}

		// 删除备份表
		dropBackupSQL := `DROP TABLE plugins_backup;`
		if _, err := s.db.Exec(dropBackupSQL); err != nil {
			log.Printf("Warning: failed to drop backup table: %v", err)
		}

		log.Println("Successfully migrated plugins table with chinese_name column")
	} else {
		// 创建插件表（如果不存在）
		pluginTableSQL := `
		CREATE TABLE IF NOT EXISTS plugins (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			chinese_name TEXT,
			description TEXT NOT NULL,
			code TEXT NOT NULL,
			wasm_binary BLOB,
			is_active BOOLEAN NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`
		if _, err := s.db.Exec(pluginTableSQL); err != nil {
			return fmt.Errorf("failed to create plugins table: %w", err)
		}
	}

	// 创建插件输入输出配置表
	ioTableSQL := `
	CREATE TABLE IF NOT EXISTS plugin_input_output (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		plugin_id INTEGER NOT NULL,
		input_desc TEXT NOT NULL, -- JSON格式存储输入描述
		output_desc TEXT NOT NULL, -- JSON格式存储输出描述
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (plugin_id) REFERENCES plugins (id) ON DELETE CASCADE,
		UNIQUE(plugin_id)
	)`

	if _, err := s.db.Exec(ioTableSQL); err != nil {
		return fmt.Errorf("failed to create plugin_input_output table: %w", err)
	}

	// 创建插件依赖表
	dependencyTableSQL := `
	CREATE TABLE IF NOT EXISTS plugin_dependencies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		plugin_id INTEGER NOT NULL,
		package TEXT NOT NULL, -- 包名，如 "strconv", "strings"
		version TEXT, -- 版本信息，可选
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (plugin_id) REFERENCES plugins (id) ON DELETE CASCADE,
		UNIQUE(plugin_id, package)
	)`

	if _, err := s.db.Exec(dependencyTableSQL); err != nil {
		return fmt.Errorf("failed to create plugin_dependencies table: %w", err)
	}

	return nil
}

// SavePluginInputOutput 保存插件输入输出配置
func (s *PluginService) SavePluginInputOutput(pluginID int64, inputDesc models.PluginInputDesc, outputDesc models.PluginOutputDesc) error {
	// 序列化输入输出描述为JSON
	inputJSON, err := json.Marshal(inputDesc)
	if err != nil {
		return fmt.Errorf("failed to marshal input desc: %w", err)
	}

	outputJSON, err := json.Marshal(outputDesc)
	if err != nil {
		return fmt.Errorf("failed to marshal output desc: %w", err)
	}

	query := `
	INSERT OR REPLACE INTO plugin_input_output (plugin_id, input_desc, output_desc, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?)`

	now := time.Now()
	_, err = s.db.Exec(query, pluginID, string(inputJSON), string(outputJSON), now, now)
	return err
}

// GetPluginInputOutput 获取插件输入输出配置
func (s *PluginService) GetPluginInputOutput(pluginID int64) (*models.PluginInputOutput, error) {
	query := `
	SELECT id, plugin_id, input_desc, output_desc, created_at, updated_at
	FROM plugin_input_output WHERE plugin_id = ?`

	var io models.PluginInputOutput
	var inputJSON, outputJSON string

	err := s.db.QueryRow(query, pluginID).Scan(
		&io.ID,
		&io.PluginID,
		&inputJSON,
		&outputJSON,
		&io.CreatedAt,
		&io.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // 没有找到输入输出配置
	}
	if err != nil {
		return nil, err
	}

	// 反序列化JSON
	if err := json.Unmarshal([]byte(inputJSON), &io.InputDesc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input desc: %w", err)
	}

	if err := json.Unmarshal([]byte(outputJSON), &io.OutputDesc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal output desc: %w", err)
	}

	return &io, nil
}

// DeletePluginInputOutput 删除插件输入输出配置
func (s *PluginService) DeletePluginInputOutput(pluginID int64) error {
	query := `DELETE FROM plugin_input_output WHERE plugin_id = ?`
	_, err := s.db.Exec(query, pluginID)
	return err
}

// GetPluginWithInputOutput 获取插件及其输入输出配置
func (s *PluginService) GetPluginWithInputOutput(id int64) (*models.Plugin, error) {
	// 获取插件基本信息
	plugin, err := s.GetPluginByID(id)
	if err != nil {
		return nil, err
	}

	// 获取输入输出配置
	ioConfig, err := s.GetPluginInputOutput(id)
	if err != nil {
		return nil, err
	}

	// 如果有输入输出配置，将其设置到插件对象中
	if ioConfig != nil {
		plugin.Input = &ioConfig.InputDesc
		plugin.Output = &ioConfig.OutputDesc
	}

	return plugin, nil
}

// GetAllPluginsWithInputOutput 获取所有插件及其输入输出配置
func (s *PluginService) GetAllPluginsWithInputOutput() ([]models.Plugin, error) {
	plugins, err := s.GetAllPlugins()
	if err != nil {
		return nil, err
	}

	// 为每个插件获取输入输出配置
	for i := range plugins {
		ioConfig, err := s.GetPluginInputOutput(plugins[i].ID)
		if err != nil {
			return nil, err
		}

		if ioConfig != nil {
			plugins[i].Input = &ioConfig.InputDesc
			plugins[i].Output = &ioConfig.OutputDesc
		}
	}

	return plugins, nil
}

// SavePluginDependencies 保存插件依赖配置
func (s *PluginService) SavePluginDependencies(pluginID int64, dependencies []models.PluginDependency) error {
	// 先删除现有的依赖配置
	if err := s.DeletePluginDependencies(pluginID); err != nil {
		return err
	}

	var newDependencies []models.PluginDependency
	dependencySet := make(map[string]bool)
	for _, dep := range models.GetDefaultImports() {
		dependencySet[`"`+dep+`"`] = true
	}
	// 我期望从dependencySet中去除models.GetDefaultImports()中的包,得到一个新的[]models.PluginDependency
	for _, dep := range dependencies {
		if !dependencySet[dep.Package] {
			newDependencies = append(newDependencies, dep)
		}
	}
	b, _ := json.Marshal(newDependencies)
	log.Printf("newDependencies: %s", string(b))

	// 插入新的依赖配置
	for _, dep := range newDependencies {
		query := `
		INSERT INTO plugin_dependencies (plugin_id, package, version, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`

		now := time.Now()
		_, err := s.db.Exec(query, pluginID, dep.Package, dep.Version, now, now)
		if err != nil {
			return fmt.Errorf("failed to save dependency %s: %w", dep.Package, err)
		}
	}

	return nil
}

// GetPluginDependencies 获取插件依赖配置
func (s *PluginService) GetPluginDependencies(pluginID int64) ([]models.PluginDependency, error) {
	query := `
		SELECT id, plugin_id, package, version, created_at, updated_at
		FROM plugin_dependencies WHERE plugin_id = ? ORDER BY package`

	rows, err := s.db.Query(query, pluginID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dependencies []models.PluginDependency
	for rows.Next() {
		var dep models.PluginDependency
		err := rows.Scan(
			&dep.ID,
			&dep.PluginID,
			&dep.Package,
			&dep.Version,
			&dep.CreatedAt,
			&dep.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, dep)
	}

	return dependencies, nil
}

// DeletePluginDependencies 删除插件依赖配置
func (s *PluginService) DeletePluginDependencies(pluginID int64) error {
	query := `DELETE FROM plugin_dependencies WHERE plugin_id = ?`
	_, err := s.db.Exec(query, pluginID)
	return err
}

// GetPluginWithDependencies 获取插件及其依赖配置
func (s *PluginService) GetPluginWithDependencies(id int64) (*models.Plugin, error) {
	// 获取插件基本信息
	plugin, err := s.GetPluginByID(id)
	if err != nil {
		return nil, err
	}

	// 获取依赖配置
	dependencies, err := s.GetPluginDependencies(id)
	if err != nil {
		return nil, err
	}

	// 获取输入输出配置
	ioConfig, err := s.GetPluginInputOutput(id)
	if err != nil {
		return nil, err
	}

	// 如果有输入输出配置，将其设置到插件对象中
	if ioConfig != nil {
		plugin.Input = &ioConfig.InputDesc
		plugin.Output = &ioConfig.OutputDesc
	}

	// 设置依赖信息到插件对象中
	plugin.Dependencies = &dependencies

	return plugin, nil
}

// GetAllPluginsWithDependencies 获取所有插件及其依赖配置
func (s *PluginService) GetAllPluginsWithDependencies() ([]models.Plugin, error) {
	plugins, err := s.GetAllPlugins()
	if err != nil {
		return nil, err
	}

	// 为每个插件获取依赖和输入输出配置
	for i := range plugins {
		// 获取依赖配置
		dependencies, err := s.GetPluginDependencies(plugins[i].ID)
		if err != nil {
			return nil, err
		}

		// 获取输入输出配置
		ioConfig, err := s.GetPluginInputOutput(plugins[i].ID)
		if err != nil {
			return nil, err
		}

		if ioConfig != nil {
			plugins[i].Input = &ioConfig.InputDesc
			plugins[i].Output = &ioConfig.OutputDesc
		}

		// 设置依赖信息到插件对象中
		plugins[i].Dependencies = &dependencies

		// 获取并设置标签信息
		tags, err := s.GetPluginTags(plugins[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get plugin tags for plugin %d: %w", plugins[i].ID, err)
		}
		plugins[i].Tags = &tags
	}

	return plugins, nil
}

// migrateDatabase 执行数据库迁移
func (s *PluginService) migrateDatabase() error {
	// 检查是否需要添加chinese_name列
	checkColumnSQL := `SELECT COUNT(*) FROM pragma_table_info('plugins') WHERE name = 'chinese_name';`

	var count int
	err := s.db.QueryRow(checkColumnSQL).Scan(&count)
	if err != nil {
		log.Printf("Warning: failed to check chinese_name column: %v", err)
		return nil // 不阻止启动
	}

	// 如果没有chinese_name列，添加它
	if count == 0 {
		addColumnSQL := `ALTER TABLE plugins ADD COLUMN chinese_name TEXT;`
		if _, err := s.db.Exec(addColumnSQL); err != nil {
			log.Printf("Warning: failed to add chinese_name column: %v", err)
			return nil // 不阻止启动
		}
		log.Println("Successfully added chinese_name column to plugins table")
	} else {
		log.Println("chinese_name column already exists in plugins table")
	}

	return nil
}

// SavePluginRelatedPlugins 保存插件的相关插件关联
func (s *PluginService) SavePluginRelatedPlugins(pluginID int64, relatedPlugins []models.PluginRelatedPlugin) error {
	// 先删除现有的关联
	if err := s.DeletePluginRelatedPlugins(pluginID); err != nil {
		return err
	}

	// 插入新的关联
	for _, relatedPlugin := range relatedPlugins {
		_, err := s.db.Exec(`
			INSERT INTO plugin_related_plugins (plugin_id, related_plugin_id, related_plugin_name, created_at, updated_at)
			VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, pluginID, relatedPlugin.RelatedPluginID, relatedPlugin.RelatedPluginName)
		if err != nil {
			return fmt.Errorf("failed to save related plugin: %w", err)
		}
	}

	return nil
}

// GetPluginRelatedPlugins 获取插件的相关插件关联
func (s *PluginService) GetPluginRelatedPlugins(pluginID int64) ([]models.PluginRelatedPlugin, error) {
	rows, err := s.db.Query(`
		SELECT id, plugin_id, related_plugin_id, related_plugin_name, created_at, updated_at
		FROM plugin_related_plugins
		WHERE plugin_id = ?
		ORDER BY created_at
	`, pluginID)
	if err != nil {
		return nil, fmt.Errorf("failed to query related plugins: %w", err)
	}
	defer rows.Close()

	var relatedPlugins []models.PluginRelatedPlugin
	for rows.Next() {
		var rp models.PluginRelatedPlugin
		err := rows.Scan(&rp.ID, &rp.PluginID, &rp.RelatedPluginID, &rp.RelatedPluginName, &rp.CreatedAt, &rp.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan related plugin: %w", err)
		}
		relatedPlugins = append(relatedPlugins, rp)
	}

	return relatedPlugins, nil
}

// DeletePluginRelatedPlugins 删除插件的所有相关插件关联
func (s *PluginService) DeletePluginRelatedPlugins(pluginID int64) error {
	_, err := s.db.Exec(`DELETE FROM plugin_related_plugins WHERE plugin_id = ?`, pluginID)
	if err != nil {
		return fmt.Errorf("failed to delete related plugins: %w", err)
	}
	return nil
}

// GetPluginWithRelatedPlugins 获取插件及其相关插件关联
func (s *PluginService) GetPluginWithRelatedPlugins(id int64) (*models.Plugin, error) {
	// 获取插件基本信息
	plugin, err := s.GetPluginByID(id)
	if err != nil {
		return nil, err
	}

	// 获取相关插件关联
	relatedPlugins, err := s.GetPluginRelatedPlugins(id)
	if err != nil {
		return nil, err
	}

	// 获取依赖配置
	dependencies, err := s.GetPluginDependencies(id)
	if err != nil {
		return nil, err
	}

	// 获取输入输出配置
	ioConfig, err := s.GetPluginInputOutput(id)
	if err != nil {
		return nil, err
	}

	// 如果有输入输出配置，将其设置到插件对象中
	if ioConfig != nil {
		plugin.Input = &ioConfig.InputDesc
		plugin.Output = &ioConfig.OutputDesc
	}

	// 设置依赖信息到插件对象中
	plugin.Dependencies = &dependencies

	// 设置相关插件关联到插件对象中
	plugin.RelatedPlugins = &relatedPlugins

	// 获取并设置标签信息
	tags, err := s.GetPluginTags(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin tags for plugin %d: %w", id, err)
	}
	plugin.Tags = &tags

	return plugin, nil
}

// GetAllPluginsWithRelatedPlugins 获取所有插件及其相关插件关联
func (s *PluginService) GetAllPluginsWithRelatedPlugins() ([]models.Plugin, error) {
	plugins, err := s.GetAllPlugins()
	if err != nil {
		return nil, err
	}

	// 为每个插件获取相关插件关联、依赖和输入输出配置
	for i := range plugins {
		// 获取相关插件关联
		relatedPlugins, err := s.GetPluginRelatedPlugins(plugins[i].ID)
		if err != nil {
			return nil, err
		}

		// 获取依赖配置
		dependencies, err := s.GetPluginDependencies(plugins[i].ID)
		if err != nil {
			return nil, err
		}

		// 获取输入输出配置
		ioConfig, err := s.GetPluginInputOutput(plugins[i].ID)
		if err != nil {
			return nil, err
		}

		// 如果有输入输出配置，将其设置到插件对象中
		if ioConfig != nil {
			plugins[i].Input = &ioConfig.InputDesc
			plugins[i].Output = &ioConfig.OutputDesc
		}

		// 设置依赖信息到插件对象中
		plugins[i].Dependencies = &dependencies

		// 设置相关插件关联到插件对象中
		plugins[i].RelatedPlugins = &relatedPlugins

		// 获取并设置标签信息
		tags, err := s.GetPluginTags(plugins[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get plugin tags for plugin %d: %w", plugins[i].ID, err)
		}
		plugins[i].Tags = &tags
	}

	return plugins, nil
}

// SavePluginTags 保存插件的标签关联
func (s *PluginService) SavePluginTags(pluginID int64, tagIDs []int64) error {
	// 先删除现有的标签关联
	if err := s.DeletePluginTags(pluginID); err != nil {
		return err
	}

	// 如果没有标签要保存，直接返回
	if len(tagIDs) == 0 {
		return nil
	}

	// 获取标签名称（用于冗余存储）
	tagMap := make(map[int64]string)
	for _, tagID := range tagIDs {
		tag, err := s.GetTagByID(tagID)
		if err != nil {
			return fmt.Errorf("failed to get tag %d: %w", tagID, err)
		}
		tagMap[tagID] = tag.Name
	}

	// 插入新的标签关联
	for _, tagID := range tagIDs {
		_, err := s.db.Exec(`
			INSERT INTO plugin_tags (plugin_id, tag_id, tag_name, created_at, updated_at)
			VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, pluginID, tagID, tagMap[tagID])
		if err != nil {
			return fmt.Errorf("failed to save plugin tag: %w", err)
		}
	}

	return nil
}

// GetPluginTags 获取插件的标签
func (s *PluginService) GetPluginTags(pluginID int64) ([]models.Tag, error) {
	rows, err := s.db.Query(`
		SELECT t.id, t.name, t.color, t.created_at, t.updated_at
		FROM tags t
		INNER JOIN plugin_tags pt ON t.id = pt.tag_id
		WHERE pt.plugin_id = ?
		ORDER BY t.name
	`, pluginID)
	if err != nil {
		return nil, fmt.Errorf("failed to query plugin tags: %w", err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt, &tag.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// DeletePluginTags 删除插件的所有标签关联
func (s *PluginService) DeletePluginTags(pluginID int64) error {
	_, err := s.db.Exec(`DELETE FROM plugin_tags WHERE plugin_id = ?`, pluginID)
	if err != nil {
		return fmt.Errorf("failed to delete plugin tags: %w", err)
	}
	return nil
}

// GetTagByID 根据ID获取标签
func (s *PluginService) GetTagByID(tagID int64) (*models.Tag, error) {
	var tag models.Tag
	err := s.db.QueryRow(`
		SELECT id, name, color, created_at, updated_at
		FROM tags
		WHERE id = ?
	`, tagID).Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt, &tag.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}
	return &tag, nil
}

// GetAllTags 获取所有标签
func (s *PluginService) GetAllTags() ([]models.Tag, error) {
	rows, err := s.db.Query(`
		SELECT id, name, color, created_at, updated_at
		FROM tags
		ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt, &tag.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}
