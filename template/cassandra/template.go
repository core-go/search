package template

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	set "github.com/core-go/search/template"
)


func Merge(obj map[string]interface{}, format set.StringFormat, param func(int) string, j int, skipArray bool, separator string, prefix string, suffix string) set.TStatement {
	results := make([]string, 0)
	parameters := format.Parameters
	k := j
	params := make([]interface{}, 0)
	if len(separator) > 0 && len(parameters) == 1 {
		p := valueOf(obj, parameters[0].Name)
		vo := reflect.Indirect(reflect.ValueOf(p))
		if vo.Kind() == reflect.Slice {
			l := vo.Len()
			if l > 0 {
				strs := make([]string, 0)
				for i := 0; i < l; i++ {
					ts := Merge(obj, format, param, k, true, "", "", "")
					strs = append(strs, ts.Query)
					model := vo.Index(i).Addr()
					params = append(params, model.Interface())
					k = k + 1
				}
				results = append(results, strings.Join(strs, separator))
				return set.TStatement{Query: prefix + strings.Join(results, "") + suffix, Params: params, Index: k}
			}
		}
	}
	texts := format.Texts
	length := len(parameters)
	for i := 0; i < length; i++ {
		results = append(results, texts[i])
		p := valueOf(obj, parameters[i].Name)
		if p != nil {
			if parameters[i].Type == set.ParamText {
				results = append(results, fmt.Sprintf("%v", p))
			} else {
				vo := reflect.Indirect(reflect.ValueOf(p))
				if vo.Kind() == reflect.Slice {
					l := vo.Len()
					if l > 0 {
						if skipArray {
							results = append(results, param(k))
							params = append(params, p)
							k = k + 1
						} else {
							sa := make([]string, 0)
							for i := 0; i < l; i++ {
								model := vo.Index(i).Addr()
								params = append(params, model.Interface())
								sa = append(sa, param(k))
								k = k + 1
							}
							results = append(results, strings.Join(sa, ","))
						}
					}
				} else {
					results = append(results, param(k))
					params = append(params, p)
					k = k + 1
				}
			}
		}
	}
	if len(texts[length]) > 0 {
		results = append(results, texts[length])
	}
	return set.TStatement{Query: prefix + strings.Join(results, "") + suffix, Params: params, Index: k}
}
func Build(obj map[string]interface{}, template set.Template, param func(int) string) (string, []interface{}) {
	results := make([]string, 0)
	params := make([]interface{}, 0)
	i := 1
	renderNodes := set.RenderTemplateNodes(obj, template.Templates)
	for _, sub := range renderNodes {
		skipArray := sub.Array == "skip"
		s := Merge(obj, sub.Format, param, i, skipArray, sub.Separator, sub.Prefix, sub.Suffix)
		i = s.Index
		if len(s.Query) > 0 {
			results = append(results, s.Query)
			if len(s.Params) > 0 {
				for _, p := range s.Params {
					params = append(params, p)
				}
			}
		}
	}
	return strings.Join(results, ""), params
}
func valueOf(m interface{}, path string) interface{} {
	arr := strings.Split(path, ".")
	i := 0
	var c interface{}
	c = m
	l1 := len(arr) - 1
	for i < len(arr) {
		key := arr[i]
		m2, ok := c.(map[string]interface{})
		if ok {
			c = m2[key]
		}
		if !ok || i >= l1 {
			return c
		}
		i++
	}
	return c
}
type QueryBuilder struct {
	Template  set.Template
	ModelType *reflect.Type
	Map       func(interface{}, *reflect.Type) map[string]interface{}
	Param     func(int) string
	Q         func(string) string
}
type Builder interface {
	BuildQuery(f interface{}) (string, []interface{})
}

func UseQuery(isTemplate bool, query func(interface{}) (string, []interface{}), id string, m map[string]*set.Template, modelType *reflect.Type, mp func(interface{}, *reflect.Type) map[string]interface{}, param func(i int) string, opts ...func(string) string) (func(interface{}) (string, []interface{}), error) {
	if !isTemplate {
		return query, nil
	}
	b, err := NewQueryBuilder(id, m, modelType, mp, param, opts...)
	if err != nil {
		return nil, err
	}
	return b.BuildQuery, nil
}
func UseQueryBuilder(isTemplate bool, builder Builder, id string, m map[string]*set.Template, modelType *reflect.Type, mp func(interface{}, *reflect.Type) map[string]interface{}, param func(i int) string, opts ...func(string) string) (Builder, error) {
	if !isTemplate {
		return builder, nil
	}
	return NewQueryBuilder(id, m, modelType, mp, param, opts...)
}
func NewQueryBuilder(id string, m map[string]*set.Template, modelType *reflect.Type, mp func(interface{}, *reflect.Type) map[string]interface{}, param func(i int) string, opts ...func(string) string) (*QueryBuilder, error) {
	t, ok := m[id]
	if !ok || t == nil {
		return nil, errors.New("cannot get the template with id " + id)
	}
	var q func(string) string
	if len(opts) > 0 {
		q = opts[0]
	} else {
		q = Q
	}
	return &QueryBuilder{Template: *t, ModelType: modelType, Map: mp, Param: param, Q: q}, nil
}
func (b *QueryBuilder) BuildQuery(f interface{}) (string, []interface{}) {
	m := b.Map(f, b.ModelType)
	if b.Q != nil {
		q, ok := m["q"]
		if ok {
			s, ok := q.(string)
			if ok {
				m["q"] = b.Q(s)
			}
		}
	}
	return Build(m, b.Template, b.Param)
}
func ReadFile(filename string) (string, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	text := string(content)
	return text, nil
}
func Q(s string) string {
	if !(strings.HasPrefix(s, "%") && strings.HasSuffix(s, "%")) {
		return "%" + s + "%"
	} else if strings.HasPrefix(s, "%") {
		return s + "%"
	} else if strings.HasSuffix(s, "%") {
		return "%" + s
	}
	return s
}
func Prefix(s string) string {
	if strings.HasSuffix(s, "%") {
		return s
	} else {
		return s + "%"
	}
}