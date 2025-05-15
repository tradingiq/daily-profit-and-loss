package main

import (
	"context"
	app2 "daily-profit-and-loss/internal/app"
	"daily-profit-and-loss/internal/config"
	"daily-profit-and-loss/internal/logger"
	"daily-profit-and-loss/internal/pnl"
	"gioui.org/app"
	"github.com/getlantern/systray"
	"os"
)

var (
	cfg     *config.Config
	mStatus *systray.MenuItem
)

func init() {
	log := logger.GetInstance()
	cfg = config.LoadConfig()

	log.SetLevel(logger.Debug)
}

func main() {
	systray.Run(onReady, onExit)

	app.Main()
}

func onReady() {
	ctx := context.Background()
	log := logger.GetInstance()

	systray.SetIcon(app2.Icon)

	systray.SetTitle("TradingIQ's Daily Crypto Profit And Loss Tracker")
	systray.SetTooltip("TradingIQ's Daily Crypto Profit And Loss Tracker")

	log.Info("application started")

	mStatus = systray.AddMenuItem("Inactive", "Status")
	systray.AddSeparator()
	mShowConfig := systray.AddMenuItem("Configuration", "Show Configuration")
	mLogs := systray.AddMenuItem("Logs", "Show application logs")
	mInfo := systray.AddMenuItem("Info", "Show application info")
	mQuit := systray.AddMenuItem("Quit", "Quit the application")

	go pnl.RunPnl(ctx, cfg, mStatus)

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
					if err := pnl.RunInfoWindow(w, log); err != nil {
						log.Error(err.Error())
					}
				}()
			case <-mLogs.ClickedCh:
				log.Info("logs menu item clicked")
				pnl.OpenLogFile(log)
			case <-mShowConfig.ClickedCh:
				log.Info("show UI menu item clicked")
				w := new(app.Window)
				w.Option(app.Title("Configuration"))
				w.Option(app.Size(350, 500))

				go func() {
					if err := config.RunConfigWindow(w, cfg, log); err != nil {
						log.Error(err.Error())
					}
				}()
			}
		}
	}()
}

func onExit() {
	os.Exit(0)
}
