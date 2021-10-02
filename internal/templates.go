package internal

import (
	_ "embed"
	"io"
	"text/template"
)

//go:embed controller.gotemplate
var templateContent string

func Render(out io.Writer, params interface{}) error {
	t, err := template.New("").Parse(templateContent)
	if err != nil {
		return err
	}
	return t.Execute(out, params)
}
