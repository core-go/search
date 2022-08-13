package template

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"time"

	set "github.com/core-go/search/template"
)

const (
	t0 = "2006-01-02 15:04:05"
	t1 = "2006-01-02T15:04:05Z"
	t2 = "2006-01-02T15:04:05-0700"
	t3 = "2006-01-02T15:04:05.0000000-0700"

	l1 = len(t1)
	l2 = len(t2)
	l3 = len(t3)
)
func Merge(obj map[string]interface{}, format set.StringFormat, j int, skipArray bool, separator string, prefix string, suffix string) set.TStatement {
	results := make([]string, 0)
	parameters := format.Parameters
	k := j
	params := make([]interface{}, 0)
	if len(separator) > 0 && len(parameters) == 1 {
		p := set.ValueOf(obj, parameters[0].Name)
		vo := reflect.Indirect(reflect.ValueOf(p))
		if vo.Kind() == reflect.Slice {
			l := vo.Len()
			if l > 0 {
				strs := make([]string, 0)
				for i := 0; i < l; i++ {
					ts := Merge(obj, format, k, true, "", "", "")
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
							vx, _ :=GetDBValue(p, 2, "")
							results = append(results, vx)
							params = append(params, p)
							k = k + 1
						} else {
							sa := make([]string, 0)
							for i := 0; i < l; i++ {
								model := vo.Index(i).Addr()
								vx, _ :=GetDBValue(model.Interface(), 4, "")
								params = append(params, model.Interface())
								sa = append(sa, vx)
								k = k + 1
							}
							results = append(results, strings.Join(sa, ","))
						}
					}
				} else {
					vx, _ :=GetDBValue(p, 2, "")
					results = append(results, vx)
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
func Build(obj map[string]interface{}, template set.Template) string {
	results := make([]string, 0)
	params := make([]interface{}, 0)
	i := 1
	renderNodes := set.RenderTemplateNodes(obj, template.Templates)
	for _, sub := range renderNodes {
		skipArray := sub.Array == "skip"
		s := Merge(obj, sub.Format, i, skipArray, sub.Separator, sub.Prefix, sub.Suffix)
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
	return strings.Join(results, "")
}
type QueryBuilder struct {
	Template  set.Template
	ModelType *reflect.Type
	Map       func(interface{}, *reflect.Type) map[string]interface{}
	Q         func(string) string
}
type Builder interface {
	BuildQuery(f interface{}) string
}

func UseQuery(isTemplate bool, query func(interface{}) string, id string, m map[string]*set.Template, modelType *reflect.Type, mp func(interface{}, *reflect.Type) map[string]interface{}, opts ...func(string) string) (func(interface{}) string, error) {
	if !isTemplate {
		return query, nil
	}
	b, err := NewQueryBuilder(id, m, modelType, mp, opts...)
	if err != nil {
		return nil, err
	}
	return b.BuildQuery, nil
}
func UseQueryBuilder(isTemplate bool, builder Builder, id string, m map[string]*set.Template, modelType *reflect.Type, mp func(interface{}, *reflect.Type) map[string]interface{}, opts ...func(string) string) (Builder, error) {
	if !isTemplate {
		return builder, nil
	}
	return NewQueryBuilder(id, m, modelType, mp, opts...)
}
func NewQueryBuilder(id string, m map[string]*set.Template, modelType *reflect.Type, mp func(interface{}, *reflect.Type) map[string]interface{}, opts ...func(string) string) (*QueryBuilder, error) {
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
	return &QueryBuilder{Template: *t, ModelType: modelType, Map: mp, Q: q}, nil
}
func (b *QueryBuilder) BuildQuery(f interface{}) string {
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
	return Build(m, b.Template)
}


func join(strs ...string) string {
	var sb strings.Builder
	for _, str := range strs {
		sb.WriteString(str)
	}
	return sb.String()
}

func WrapString(v string) string {
	return join(`'`, v, `'`)
}

func GetDBValue(v interface{}, scale int8, layoutTime string) (string, bool) {
	switch v.(type) {
	case string:
		s0 := v.(string)
		if len(s0) == 0 {
			return "''", true
		}

		return WrapString(s0), true
	case bool:
		b0 := v.(bool)
		if b0 {
			return "true", true
		} else {
			return "false", true
		}
	case int:
		return strconv.Itoa(v.(int)), true
	case int64:
		return strconv.FormatInt(v.(int64), 10), true
	case int32:
		return strconv.FormatInt(int64(v.(int32)), 10), true
	case big.Int:
		var z1 big.Int
		z1 = v.(big.Int)
		return z1.String(), true
	case float64:
		if scale >= 0 {
			mt := "%." + strconv.Itoa(int(scale)) + "f"
			return fmt.Sprintf(mt, v), true
		}
		return fmt.Sprintf("'%f'", v), true
	case time.Time:
		tf := v.(time.Time)
		if len(layoutTime) > 0 {
			f := tf.Format(layoutTime)
			return WrapString(f), true
		}
		f := tf.Format(t0)
		return WrapString(f), true
	case big.Float:
		n1 := v.(big.Float)
		if scale >= 0 {
			n2 := Round(n1, int(scale))
			return fmt.Sprintf("%v", &n2), true
		} else {
			return fmt.Sprintf("%v", &n1), true
		}
	case big.Rat:
		n1 := v.(big.Rat)
		if scale >= 0 {
			return RoundRat(n1, scale), true
		} else {
			return n1.String(), true
		}
	case float32:
		if scale >= 0 {
			mt := "%." + strconv.Itoa(int(scale)) + "f"
			return fmt.Sprintf(mt, v), true
		}
		return fmt.Sprintf("'%f'", v), true
	default:
		if scale >= 0 {
			v2 := reflect.ValueOf(v)
			if v2.Kind() == reflect.Ptr {
				v2 = v2.Elem()
			}
			if v2.NumField() == 1 {
				f := v2.Field(0)
				fv := f.Interface()
				k := f.Kind()
				if k == reflect.Ptr {
					if f.IsNil() {
						return "null", true
					} else {
						fv = reflect.Indirect(reflect.ValueOf(fv)).Interface()
						sv, ok := fv.(big.Float)
						if ok {
							return sv.Text('f', int(scale)), true
						} else {
							return "", false
						}
					}
				} else {
					sv, ok := fv.(big.Float)
					if ok {
						return sv.Text('f', int(scale)), true
					} else {
						return "", false
					}
				}
			} else {
				return "", false
			}
		} else {
			return "", false
		}
	}
	return "", false
}
func ParseDates(args []interface{}, dates []int) []interface{} {
	if args == nil || len(args) == 0 {
		return nil
	}
	if dates == nil || len(dates) == 0 {
		return args
	}
	res := append([]interface{}{}, args...)
	for _, d := range dates {
		if d >= len(args) {
			break
		}
		a := args[d]
		if s, ok := a.(string); ok {
			switch len(s) {
			case l1:
				t, err := time.Parse(t1, s)
				if err == nil {
					res[d] = t
				}
			case l2:
				t, err := time.Parse(t2, s)
				if err == nil {
					res[d] = t
				}
			case l3:
				t, err := time.Parse(t3, s)
				if err == nil {
					res[d] = t
				}
			}
		}
	}
	return res
}
func Round(num big.Float, scale int) big.Float {
	marshal, _ := num.MarshalText()
	var dot int
	for i, v := range marshal {
		if v == 46 {
			dot = i + 1
			break
		}
	}
	a := marshal[:dot]
	b := marshal[dot : dot+scale+1]
	c := b[:len(b)-1]

	if b[len(b)-1] >= 53 {
		c[len(c)-1] += 1
	}
	var r []byte
	r = append(r, a...)
	r = append(r, c...)
	num.UnmarshalText(r)
	return num
}
func RoundRat(rat big.Rat, scale int8) string {
	digits := int(math.Pow(float64(10), float64(scale)))
	floatNumString := rat.RatString()
	sl := strings.Split(floatNumString, "/")
	a := sl[0]
	b := sl[1]
	c, _ := strconv.Atoi(a)
	d, _ := strconv.Atoi(b)
	intNum := c / d
	surplus := c - d*intNum
	e := surplus * digits / d
	r := surplus * digits % d
	if r >= d/2 {
		e += 1
	}
	res := strconv.Itoa(intNum) + "." + strconv.Itoa(e)
	return res
}
