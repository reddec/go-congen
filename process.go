package congen

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/reddec/go-congen/internal"

	"github.com/iancoleman/strcase"
	"golang.org/x/net/html"
	"golang.org/x/tools/imports"
)

// Process HTML source, find all defined forms and generate controller stub.
// Returns rendered template.
// templateFile or htmlSource has Name() function will be used for embedding.
func Process(htmlSource io.Reader, htmlFile string, packageName string) ([]byte, error) {
	forms, err := Forms(htmlSource)
	if err != nil {
		return nil, fmt.Errorf("get forms: %w", err)
	}

	if htmlFile == "" {
		if v, ok := htmlSource.(interface{ Name() string }); ok {
			htmlFile = v.Name()
		}
	}

	var buffer bytes.Buffer

	if err := internal.Render(&buffer, &RenderEnv{
		Forms:   forms,
		Package: packageName,
		Input:   htmlFile,
	}); err != nil {
		return nil, fmt.Errorf("render: %w", err)
	}

	out, err := imports.Process("", buffer.Bytes(), nil)
	if err != nil {
		return nil, fmt.Errorf("format file: %w", err)
	}
	return out, nil
}

// ProcessFile invokes process, feeds content of htmlFile and writes result to outputFile.
func ProcessFile(htmlFile string, outputFile string, packageName string) error {
	f, err := os.Open(htmlFile)
	if err != nil {
		return fmt.Errorf("open input file: %w", err)
	}
	defer f.Close()

	relInput, err := filepath.Rel(filepath.Dir(outputFile), htmlFile)
	if err != nil {
		return fmt.Errorf("calculate relative path to input: %w", err)
	}

	data, err := Process(f, relInput, packageName)
	if err != nil {
		return fmt.Errorf("process: %w", err)
	}

	if err := ioutil.WriteFile(outputFile, data, 0755); err != nil {
		return fmt.Errorf("save result file: %w", err)
	}
	return nil
}

type Path struct {
	Name string
	Get  *Form
	Post *Form
}

type RenderEnv struct {
	Input   string
	Forms   []Form
	Package string
}

func (r *RenderEnv) Paths() []Path {
	var pathsIndex = make(map[string]Path)
	for _, f := range r.Forms {
		cp := f
		p := pathsIndex[f.Action]
		if f.Method == "post" {
			p.Post = &cp
		} else {
			p.Get = &cp
		}
		p.Name = f.Action
		pathsIndex[f.Action] = p
	}

	var list = make([]Path, 0, len(pathsIndex))
	for _, p := range pathsIndex {
		list = append(list, p)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})

	return list
}

type Field struct {
	Type  Type
	Label string
}

func (f *Field) IsString() bool {
	return f.Type == String
}

func (f *Field) IsFloat() bool {
	return f.Type == Double
}

func (f *Field) IsInteger() bool {
	return f.Type == Integer
}

func (f *Field) IsBool() bool {
	return f.Type == Boolean
}

func (f *Field) Name() string {
	return strcase.ToCamel(f.Label)
}

type Form struct {
	Action string
	Method string
	Fields []Field
}

func (f *Form) MergeFields(fields []Field) {
	for _, field := range fields {
		if !f.HasField(field.Label) {
			f.Fields = append(f.Fields, field)
		}
	}
}

func (f *Form) Name() string {
	return strcase.ToCamel(f.Action)
}

func (f *Form) HasField(label string) bool {
	for _, field := range f.Fields {
		if field.Label == label {
			return true
		}
	}
	return false
}

// Forms from parsed HTML stream.
func Forms(reader io.Reader) ([]Form, error) {
	parser, err := html.Parse(reader)
	if err != nil {
		return nil, fmt.Errorf("parse stream: %w", err)
	}

	list := findForms(parser)
	// merge duplicates
	var usedForms = map[string]*Form{}
	for _, form := range list {
		oldForm, exists := usedForms[form.Action]
		if !exists {
			cp := form
			usedForms[form.Action] = &cp
		} else {
			oldForm.MergeFields(form.Fields)
		}
	}

	// convert back to array
	var ans = make([]Form, 0, len(usedForms))
	for _, form := range usedForms {
		ans = append(ans, *form)
	}

	// sort by name
	sort.Slice(ans, func(i, j int) bool {
		return ans[i].Action < ans[j].Action
	})
	return ans, nil
}

func findForms(node *html.Node) []Form {
	var ans []Form
	if node.Type == html.ElementNode && strings.EqualFold(node.Data, "form") {
		// parse form
		return []Form{parseForm(node)}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		ans = append(ans, findForms(c)...)
	}

	return ans
}

func parseForm(node *html.Node) Form {
	fields := findFields(node)
	// deduplicate fields
	var cp = make([]Field, 0, len(fields))
	exists := map[string]bool{}
	for _, field := range fields {
		if !exists[field.Label] {
			exists[field.Label] = true
			cp = append(cp, field)
		}
	}
	fields = cp

	return Form{
		Action: findAttr(node, "action"),
		Method: strings.ToLower(findAttr(node, "method")),
		Fields: fields,
	}
}

func findFields(node *html.Node) []Field {
	if node.Type != html.ElementNode {
		return nil
	}
	name := findAttr(node, "name")
	if name != "" {
		switch strings.ToLower(node.Data) {
		case "input":
			return []Field{{
				Type:  getInputType(node),
				Label: name,
			}}
		case "textarea", "button", "select":
			return []Field{{
				Type:  String,
				Label: name,
			}}
		}
	}
	var ans []Field
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		ans = append(ans, findFields(c)...)
	}
	return ans
}

func findAttr(node *html.Node, name string) string {
	for _, a := range node.Attr {
		if strings.EqualFold(a.Key, name) {
			return a.Val
		}
	}
	return ""
}

func getInputType(node *html.Node) Type {
	var baseType = findAttr(node, "type")
	var subType = findAttr(node, "format")

	switch {
	case baseType == "number":
		switch subType {
		case "integer":
			return Integer
		default:
			return Double
		}
	case baseType == "checkbox":
		return Boolean
	case baseType == "text":
		fallthrough
	default:
		return String
	}
}

type Type int

const (
	String Type = iota
	Double
	Integer
	Boolean
)

func (t Type) String() string {
	switch t {
	case String:
		return "string"
	case Double:
		return "float64"
	case Integer:
		return "int64"
	case Boolean:
		return "bool"
	default:
		return "interface{}"
	}
}
