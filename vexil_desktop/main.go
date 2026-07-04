package main

import (
	"context"
	"embed"
	"os"

	"vexil/internal/app"
	"vexil/internal/cli"
	"vexil/internal/config"
	"vexil/internal/gui"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 有命令行参数 → CLI 模式
	if len(os.Args) > 1 {
		cli.Run(os.Args)
		return
	}

	// 无参数 → GUI 模式
	cfg := config.LoadAppConfig()
	cfg.Apply()
	application := app.New(cfg)
	handler := gui.NewHandler(application)

	err := wails.Run(&options.App{
		Title:  "Vexil",
		Width:  800,
		Height: 700,
		Assets: assets,
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop: true,
		},
		Bind: []interface{}{
			handler,
		},
		OnStartup: func(ctx context.Context) {
			handler.SetContext(ctx)
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}