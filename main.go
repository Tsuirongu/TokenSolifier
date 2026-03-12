package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"

	"loji-app/backend/app"
	"loji-app/backend/services"

	"github.com/joho/godotenv"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 加载环境变量
	loadEnvFiles()

	// 初始化数据库
	db, err := services.InitDatabase()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// 创建应用实例
	application := app.NewApp(db)

	// 创建应用窗口配置
	err = wails.Run(&options.App{
		Title:            "Loji",
		Width:            1200,
		Height:           800,
		MinWidth:         1000,
		MinHeight:        700,
		MaxWidth:         2000,
		MaxHeight:        1400,
		DisableResize:    false,
		Frameless:        true,
		AlwaysOnTop:      true,
		StartHidden:      false,
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 0},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: application.Startup,
		Bind: []interface{}{
			application,
		},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  true,
				FullSizeContent:            true,
				HideTitleBar:               true,
			},
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			Appearance:           mac.NSAppearanceNameDarkAqua,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			BackdropType:         windows.Mica,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}

// loadEnvFiles 加载环境变量文件
func loadEnvFiles() {
	// 尝试从当前工作目录加载 .env
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found in current directory, trying home directory...")

		// 尝试从用户主目录的 .loji-app 加载
		homeDir, err := os.UserHomeDir()
		if err == nil {
			envPath := filepath.Join(homeDir, ".loji-app", ".env")
			if err := godotenv.Load(envPath); err != nil {
				log.Println("Warning: .env file not found in ~/.loji-app/, using system environment variables")
			} else {
				log.Println("Loaded environment from:", envPath)
			}
		}
	} else {
		log.Println("Loaded environment from: .env")
	}
}
