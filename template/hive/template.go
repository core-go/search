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

func Merge(obj map[string]interface{}, format set.StringFormat, skipArray bool, separator string, prefix string, suffix string) string {
	results := make([]string, 0)
	parameters := format.Parameters
	if len(separator) > 0 && len(parameters) == 1 {
		p := set.ValueOf(obj, parameters[0].Name)
		vo := reflect.Indirect(reflect.ValueOf(p))
		if vo.Kind() == reflect.Slice {
			l := vo.Len()
			if l > 0 {
				strs := make([]string, 0)
				for i := 0; i < l; i++ {
					ts := Merge(obj, format, true, "", "", "")
					strs = append(strs, ts)
				}
				results = append(results, strings.Join(strs, separator))
				return prefix + strings.Join(results, "") + suffix
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
							vx, _ := GetDBValue(p, 2, "")
							results = append(results, vx)
						} else {
							sa := make([]string, 0)
							for i := 0; i < l; i++ {
								model := vo.Index(i).Addr()
								vx, _ := GetDBValue(model.Interface(), 4, "")
								sa = append(sa, vx)
							}
							results = append(results, strings.Join(sa, ","))
						}
					}
				} else {
					vx, _ := GetDBValue(p, 2, "")
					results = append(results, vx)
				}
			}
		}
	}
	if len(texts[length]) > 0 {
		results = append(results, texts[length])
	}
	return prefix + strings.Join(results, "") + suffix
}
func Build(obj map[string]interface{}, template set.Template) string {
	results := make([]string, 0)
	renderNodes := set.RenderTemplateNodes(obj, template.Templates)
	for _, sub := range renderNodes {
		skipArray := sub.Array == "skip"
		s := Merge(obj, sub.Format, skipArray, sub.Separator, sub.Prefix, sub.Suffix)
		if len(s) > 0 {
			results = append(results, s)
		}
	}
	return strings.Join(results, "")
}

type QueryBuilder struct {
	Template  set.Template
	ModelType *reflect.Type
	Map       func(interface{}, *reflect.Type, ...func(string, reflect.Type) string) map[string]interface{}
	BuildSort func(string, reflect.Type) string
	Q         func(string) string
}
type Builder interface {
	BuildQuery(f interface{}) string
}

func UseQuery(id string, m map[string]*set.Template, modelType *reflect.Type, mp func(interface{}, *reflect.Type, ...func(string, reflect.Type) string) map[string]interface{}, buildSort func(string, reflect.Type) string, opts ...func(string) string) (func(interface{}) string, error) {
	b, err := NewQueryBuilder(id, m, modelType, mp, buildSort, opts...)
	if err != nil {
		return nil, err
	}
	return b.BuildQuery, nil
}
func GetQuery(isTemplate bool, query func(interface{}) string, id string, m map[string]*set.Template, modelType *reflect.Type, mp func(interface{}, *reflect.Type, ...func(string, reflect.Type) string) map[string]interface{}, buildSort func(string, reflect.Type) string, opts ...func(string) string) (func(interface{}) string, error) {
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
func (b *QueryBuilder) BuildQuery(f interface{}) string {
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
