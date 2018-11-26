package templates

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

type Template struct {
	template *template.Template
}

func replaceTemplateShorthand(str string) string {
	re := regexp.MustCompile(`\[\[([^\]]+?)\]\]`)
	return re.ReplaceAllString(str, "{{tpl . $1}}")
}

func NewTemplate(id, tpl string, subTemplates map[string]string) (*Template, error) {
	t := template.New(id)
	funcMap := template.FuncMap{
		"tpl": func(data interface{}, values ...string) (interface{}, error) {
			buf := bytes.NewBuffer([]byte{})
			for _, v := range values {
				x := t.Lookup(v)
				if x == nil {
					continue
				}
				err := x.Execute(buf, data)
				return buf.String(), err
			}
			return "", errors.New("No matching template found")
		},
	}
	x, err := t.Funcs(funcMap).Parse(replaceTemplateShorthand(tpl))
	if err != nil {
		return nil, err
	}

	for name, subTemplate := range subTemplates {
		_, err := x.Parse(fmt.Sprintf("{{define \"%s\"}}%s{{end}}", name, replaceTemplateShorthand(subTemplate)))
		if err != nil {
			return nil, err
		}
	}
	return &Template{
		template: x,
	}, nil
}

func (t *Template) Render(data interface{}) (string, error) {
	var res bytes.Buffer
	err := t.template.Execute(&res, data)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(res.String()), nil
}

func FilterResultToBool(res string) (bool, error) {
	if res == "" {
		return false, nil
	}
	return strconv.ParseBool(res)
}
