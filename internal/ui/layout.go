package ui

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

var CommonInsets = struct {
	Title  layout.Inset
	Label  layout.Inset
	Field  layout.Inset
	Button layout.Inset
	Status layout.Inset
}{
	Title:  layout.Inset{Top: unit.Dp(20), Bottom: unit.Dp(20), Left: unit.Dp(20), Right: unit.Dp(20)},
	Label:  layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(8), Left: unit.Dp(20), Right: unit.Dp(20)},
	Field:  layout.Inset{Bottom: unit.Dp(16), Left: unit.Dp(20), Right: unit.Dp(20)},
	Button: layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(8)},
	Status: layout.Inset{Top: unit.Dp(20), Left: unit.Dp(20), Right: unit.Dp(20)},
}

func PaddedWidget(inset layout.Inset, widget layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return inset.Layout(gtx, widget)
	}
}

func Section(gtx layout.Context, widgets ...layout.Widget) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx, widgetsToFlexChildren(widgets)...)
}

func FormSection(th *material.Theme, items map[string]string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		var widgets []layout.Widget

		for label, value := range items {

			l, v := label, value
			widgets = append(widgets,
				PaddedWidget(CommonInsets.Label, func(gtx layout.Context) layout.Dimensions {
					return material.Body1(th, l+": "+v).Layout(gtx)
				}),
			)
		}

		return Section(gtx, widgets...)
	}
}

func Card(gtx layout.Context, content layout.Widget) layout.Dimensions {
	return layout.UniformInset(unit.Dp(16)).Layout(gtx, content)
}
