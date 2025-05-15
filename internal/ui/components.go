package ui

import (
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type LabeledInput struct {
	Label       string
	Editor      *widget.Editor
	Hint        string
	Theme       *material.Theme
	LabelInset  layout.Inset
	EditorInset layout.Inset
}

func NewLabeledInput(th *material.Theme, label, hint string, editor *widget.Editor) LabeledInput {
	return LabeledInput{
		Label:       label,
		Editor:      editor,
		Hint:        hint,
		Theme:       th,
		LabelInset:  layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(8), Left: unit.Dp(20), Right: unit.Dp(20)},
		EditorInset: layout.Inset{Bottom: unit.Dp(16), Left: unit.Dp(20), Right: unit.Dp(20)},
	}
}

func (l LabeledInput) Layout(gtx layout.Context) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return l.LabelInset.Layout(gtx,
				func(gtx layout.Context) layout.Dimensions {
					label := material.Body1(l.Theme, l.Label)
					return label.Layout(gtx)
				},
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return l.EditorInset.Layout(gtx,
				func(gtx layout.Context) layout.Dimensions {
					ed := material.Editor(l.Theme, l.Editor, l.Hint)
					return ed.Layout(gtx)
				},
			)
		}),
	)
}

func CenteredButton(th *material.Theme, button *widget.Clickable, text string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(th, button, text)
			return btn.Layout(gtx)
		})
	}
}

func Title(th *material.Theme, txt string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: unit.Dp(20), Bottom: unit.Dp(20), Left: unit.Dp(20), Right: unit.Dp(20)}.Layout(gtx,
			func(gtx layout.Context) layout.Dimensions {
				title := material.H4(th, txt)
				title.Alignment = text.Middle
				return title.Layout(gtx)
			},
		)
	}
}

func InfoText(th *material.Theme, text string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: unit.Dp(5), Bottom: unit.Dp(5), Left: unit.Dp(20), Right: unit.Dp(20)}.Layout(gtx,
			func(gtx layout.Context) layout.Dimensions {
				label := material.Body1(th, text)
				return label.Layout(gtx)
			},
		)
	}
}

func StatusText(th *material.Theme, txt string) layout.Widget {
	if txt == "" {
		return func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{}
		}
	}

	return func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: unit.Dp(20), Left: unit.Dp(20), Right: unit.Dp(20)}.Layout(gtx,
			func(gtx layout.Context) layout.Dimensions {
				label := material.Body1(th, txt)
				label.Alignment = text.Middle
				return label.Layout(gtx)
			},
		)
	}
}

func Spacer(height unit.Dp) layout.Widget {
	return layout.Spacer{Height: height}.Layout
}
