package ui

import (
	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
)

type WindowHandler func(gtx layout.Context, th interface{}, closeRequested chan bool) layout.Dimensions

func RunWindow(w *app.Window, handler WindowHandler, theme interface{}) error {
	var ops op.Ops
	closeRequested := make(chan bool, 1)

	for {
		e := w.Event()

		switch e := e.(type) {
		case app.DestroyEvent:
			return e.Err

		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			select {
			case <-closeRequested:
				w.Perform(system.ActionClose)
			default:

			}

			handler(gtx, theme, closeRequested)

			e.Frame(gtx.Ops)
		}
	}
}

func VerticalLayout(gtx layout.Context, widgets ...layout.Widget) layout.Dimensions {
	return layout.Flex{
		Axis:      layout.Vertical,
		Spacing:   layout.SpaceAround,
		Alignment: layout.Middle,
	}.Layout(gtx, widgetsToFlexChildren(widgets)...)
}

func CloseButtonHandler(button *widget.Clickable, gtx layout.Context, closeRequested chan bool) {
	if button.Clicked(gtx) {
		select {
		case closeRequested <- true:

		default:

		}
	}
}

func widgetsToFlexChildren(widgets []layout.Widget) []layout.FlexChild {
	children := make([]layout.FlexChild, len(widgets))
	for i, w := range widgets {

		widget := w
		children[i] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return widget(gtx)
		})
	}
	return children
}
