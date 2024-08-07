package template

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	set "github.com/core-go/search/template"
)

func Merge(obj map[string]interface{}, format set.StringFormat, skipArray bool, separator string, prefix string, suffix string) set.TStatement {
	results := make([]string, 0)
	parameters := format.Parameters
	params := make([]interface{}, 0)
	if len(separator) > 0 && len(parameters) == 1 {
		p := set.ValueOf(obj, parameters[0].Name)
		vo := reflect.Indirect(reflect.ValueOf(p))
		if vo.Kind() == reflect.Slice {
			l := vo.Len()
			if l > 0 {
				strs := make([]string, 0)
				for i := 0; i < l; i++ {
					ts := Merge(obj, format, true, "", "", "")
					strs = append(strs, ts.Query)
					model := vo.Index(i).Addr()
					params = append(params, model.Interface())
				}
				results = append(results, strings.Join(strs, separator))
				return set.TStatement{Query: prefix + strings.Join(results, "") + suffix, Params: params}
			}
		}
	}
	texts := format.Texts
	length := len(parameters)
	for i := 0; i < length; i++ {
		results = append(results, texts[i])
		p := set.ValueOf(obj, parameters[i].Name)
		if p != nil {
			if parameters[i].Type == set.ParamText {
				results = append(results, fmt.Sprintf("%v", p))
			} else {
				vo := reflect.Indirect(reflect.ValueOf(p))
				if vo.Kind() == reflect.Slice {
					l := vo.Len()
					if l > 0 {
						if skipArray {
							results = append(results, "?")
							params = append(params, p)
						} else {
							sa := make([]string, 0)
							for i := 0; i < l; i++ {
								model := vo.Index(i).Addr()
								params = append(params, model.Interface())
								sa = append(sa, "?")
							}
							results = append(results, strings.Join(sa, ","))
						}
					}
				} else {
					results = append(results, "?")
					params = append(params, p)
				}
			}
		}
	}
	if len(texts[length]) > 0 {
		results = append(results, texts[length])
	}
	return set.TStatement{Query: prefix + strings.Join(results, "") + suffix, Params: params}
}
func Build(obj map[string]interface{}, template set.Template) (string, []interface{}) {
	results := make([]string, 0)
	params := make([]interface{}, 0)
	renderNodes := set.RenderTemplateNodes(obj, template.Templates)
	for _, sub := range renderNodes {
		skipArray := sub.Array == "skip"
		s := Merge(obj, sub.Format, skipArray, sub.Separator, sub.Prefix, sub.Suffix)
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

type QueryBuilder struct {
	Template  set.Template
	ModelType *reflect.Type
	Map       func(interface{}, *reflect.Type, ...func(string, reflect.Type) string) map[string]interface{}
	BuildSort func(string, reflect.Type) string
	Q         func(string) string
}
type Builder interface {
	BuildQuery(f interface{}) (string, []interface{})
}

func UseQuery(id string, m map[string]*set.Template, modelType *reflect.Type, mp func(interface{}, *reflect.Type, ...func(string, reflect.Type) string) map[string]interface{}, buildSort func(string, reflect.Type) string, opts ...func(string) string) (func(interface{}) (string, []interface{}), error) {
	b, err := NewQueryBuilder(id, m, modelType, mp, buildSort, opts...)
	if err != nil {
		return nil, err
	}
	return b.BuildQuery, nil
}
func GetQuery(isTemplate bool, query func(interface{}) (string, []interface{}), id string, m map[string]*set.Template, modelType *reflect.Type, mp func(interface{}, *reflect.Type, ...func(string, reflect.Type) string) map[string]interface{}, buildSort func(string, reflect.Type) string, opts ...func(string) string) (func(interface{}) (string, []interface{}), error) {
	if !isTemplate {
		return query, nil
	}
	b, err := NewQueryBuilder(id, m, modelType, mp, buildSort, opts...)
	if err != nil {
		return nil, err
	}
	return b.BuildQuery, nil
}
func UseQueryBuilder(id string, m map[string]*set.Template, modelType *reflect.Type, mp func(interface{}, *reflect.Type, ...func(string, reflect.Type) string) map[string]interface{}, buildSort func(string, reflect.Type) string, opts ...func(string) string) (Builder, error) {
	return NewQueryBuilder(id, m, modelType, mp, buildSort, opts...)
}
func GetQueryBuilder(isTemplate bool, builder Builder, id string, m map[string]*set.Template, modelType *reflect.Type, mp func(interface{}, *reflect.Type, ...func(string, reflect.Type) string) map[string]interface{}, buildSort func(string, reflect.Type) string, opts ...func(string) string) (Builder, error) {
	if !isTemplate {
		return builder, nil
	}
	return NewQueryBuilder(id, m, modelType, mp, buildSort, opts...)
}
func NewQueryBuilder(id string, m map[string]*set.Template, modelType *reflect.Type, mp func(interface{}, *reflect.Type, ...func(string, reflect.Type) string) map[string]interface{}, buildSort func(string, reflect.Type) string, opts ...func(string) string) (Builder, error) {
	t, ok := m[id]
	if !ok || t == nil {
		return nil, errors.New("cannot get the template with id " + id)
	}
	var q func(string) string
	if len(opts) > 0 {
		q = opts[0]
	} else {
		q = set.Q
	}
	return &QueryBuilder{Template: *t, ModelType: modelType, Map: mp, BuildSort: buildSort, Q: q}, nil
}
func (b *QueryBuilder) BuildQuery(f interface{}) (string, []interface{}) {
	m := b.Map(f, b.ModelType, b.BuildSort)
	if b.Q != nil {
		q, ok := m["q"]
		if ok {
			s, ok := q.(string)
			if ok {
				m["q"] = b.Q(s)
			}
		}
	}
	return Build(m, b.Template)
}
