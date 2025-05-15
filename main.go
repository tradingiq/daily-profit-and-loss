package main

import (
	"context"
	"daily-profit-and-loss/pkg/logger"
	"gioui.org/app"
	"github.com/getlantern/systray"
	"os"
)

var (
	log     *logger.Logger
	cfg     *Config
	mStatus *systray.MenuItem
)

func init() {
	log = logger.GetInstance()
	cfg = loadConfig()

	log.SetLevel(logger.Debug)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go runPnl(ctx, cfg)

	systray.Run(onReady, onExit)

	app.Main()
}

func onReady() {
	systray.SetIcon(icon)

	systray.SetTitle("TradingIQ's Daily Crypto Profit And Loss Tracker")
	systray.SetTooltip("TradingIQ's Daily Crypto Profit And Loss Tracker")

	log.Info("application started")

	mStatus = systray.AddMenuItem("Inactive", "Running")
	systray.AddSeparator()
	mShowConfig := systray.AddMenuItem("Configuration", "Show Configuration")
	mLogs := systray.AddMenuItem("Logs", "Show application logs")
	mInfo := systray.AddMenuItem("Info", "Show application info")
	mQuit := systray.AddMenuItem("Quit", "Quit the application")

	go func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				log.Info("application shutdown requested")
				systray.Quit()
				os.Exit(0)
				return
			case <-mInfo.ClickedCh:
				log.Info("info menu item clicked")

				w := new(app.Window)
				w.Option(app.Title("Info"))
				w.Option(app.Size(350, 450))

				go func() {
					if err := runInfoWindow(w); err != nil {
						log.Error(err.Error())
					}
				}()
			case <-mLogs.ClickedCh:
				log.Info("logs menu item clicked")
				openLogFile()
			case <-mShowConfig.ClickedCh:
				log.Info("show UI menu item clicked")
				w := new(app.Window)
				w.Option(app.Title("Configuration"))
				w.Option(app.Size(350, 500))

				go func() {
					if err := runConfigWindow(w); err != nil {
						log.Error(err.Error())
					}
				}()
			}
		}
	}()
}

func onExit() {
	log.Info("Application shutting down")
	os.Exit(0)
}
