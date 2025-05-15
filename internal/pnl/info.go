package pnl

import (
	app2 "daily-profit-and-loss/internal/app"
	"daily-profit-and-loss/internal/logger"
	"daily-profit-and-loss/internal/ui"
	"fmt"
	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image/color"
)

func RunInfoWindow(w *app.Window, log *logger.Logger) error {

	th := material.NewTheme()

	var closeButton widget.Clickable

	infoHandler := func(gtx layout.Context, theme interface{}, closeRequested chan bool) layout.Dimensions {
		th := theme.(*material.Theme)

		ui.CloseButtonHandler(&closeButton, gtx, closeRequested)

		titleWidget := func(gtx layout.Context) layout.Dimensions {
			return ui.CommonInsets.Title.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				title := material.H5(th, "Daily Profit and Loss Tracker")
				title.Font.Weight = 400
				title.Color = color.NRGBA{R: 0, G: 0, B: 150, A: 255}
				return title.Layout(gtx)
			})
		}

		versionInfo := ui.InfoText(th, "Version: 0.1.8")
		copyRightInfo := ui.InfoText(th, "(c) by Victor J. C. Geyer")
		logFileInfo := ui.InfoText(th, fmt.Sprintf("Log file: %s", log.GetLogFilePath()))
		configFileInfo := ui.InfoText(th, fmt.Sprintf("Configuration file: %s", app2.GetConfigPath()))

		return ui.VerticalLayout(gtx,
			titleWidget,
			versionInfo,
			copyRightInfo,
			logFileInfo,
			configFileInfo,
			ui.Spacer(unit.Dp(15)),
			ui.CenteredButton(th, &closeButton, "Close"),
		)
	}

	return ui.RunWindow(w, infoHandler, th)
}
