package template

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
)

const (
	TypeText       = "text"
	TypeIsNotEmpty = "isNotEmpty"
	TypeIsEmpty    = "isEmpty"
	TypeIsEqual    = "isEqual"
	TypeIsNotEqual = "isNotEqual"
	TypeIsNull     = "isNull"
	TypeIsNotNull  = "isNotNull"
	ParamText      = "text"
)

var ns = []string{"isNotNull", "isNull", "isEqual", "isNotEqual", "isEmpty", "isNotEmpty"}

func isValidNode(n string) bool {
	for _, s := range ns {
		if n == s {
			return true
		}
	}
	return false
}

type StringFormat struct {
	Texts      []string    `yaml:"" mapstructure:"texts" json:"texts,omitempty" gorm:"column:texts" bson:"texts,omitempty" dynamodbav:"texts,omitempty" firestore:"texts,omitempty"`
	Parameters []Parameter `yaml:"" mapstructure:"parameters" json:"parameters,omitempty" gorm:"column:parameters" bson:"parameters,omitempty" dynamodbav:"parameters,omitempty" firestore:"parameters,omitempty"`
}
type Parameter struct {
	Name string `yaml:"" mapstructure:"name" json:"name,omitempty" gorm:"column:name" bson:"name,omitempty" dynamodbav:"name,omitempty" firestore:"name,omitempty"`
	Type string `yaml:"" mapstructure:"type" json:"type,omitempty" gorm:"column:type" bson:"type,omitempty" dynamodbav:"type,omitempty" firestore:"type,omitempty"`
}
type TemplateNode struct {
	Type      string       `yaml:"type" mapstructure:"type" json:"type,omitempty" gorm:"column:type" bson:"type,omitempty" dynamodbav:"type,omitempty" firestore:"type,omitempty"`
	Text      string       `yaml:"text" mapstructure:"text" json:"text,omitempty" gorm:"column:text" bson:"text,omitempty" dynamodbav:"text,omitempty" firestore:"text,omitempty"`
	Property  string       `yaml:"property" mapstructure:"property" json:"property,omitempty" gorm:"column:property" bson:"property,omitempty" dynamodbav:"property,omitempty" firestore:"property,omitempty"`
	Value     string       `yaml:"value" mapstructure:"value" json:"value,omitempty" gorm:"column:value" bson:"value,omitempty" dynamodbav:"value,omitempty" firestore:"value,omitempty"`
	Array     string       `yaml:"array" mapstructure:"array" json:"array,omitempty" gorm:"column:array" bson:"array,omitempty" dynamodbav:"array,omitempty" firestore:"array,omitempty"`
	Separator string       `yaml:"separator" mapstructure:"separator" json:"array,omitempty" gorm:"column:separator" bson:"separator,omitempty" dynamodbav:"separator,omitempty" firestore:"separator,omitempty"`
	Prefix    string       `yaml:"prefix" mapstructure:"prefix" json:"array,omitempty" gorm:"column:prefix" bson:"prefix,omitempty" dynamodbav:"prefix,omitempty" firestore:"prefix,omitempty"`
	Suffix    string       `yaml:"suffix" mapstructure:"suffix" json:"array,omitempty" gorm:"column:suffix" bson:"suffix,omitempty" dynamodbav:"suffix,omitempty" firestore:"suffix,omitempty"`
	Format    StringFormat `yaml:"format" mapstructure:"format" json:"format,omitempty" gorm:"column:format" bson:"format,omitempty" dynamodbav:"format,omitempty" firestore:"format,omitempty"`
}
type Template struct {
	Id        string         `yaml:"id" mapstructure:"id" json:"id,omitempty" gorm:"column:id" bson:"id,omitempty" dynamodbav:"id,omitempty" firestore:"id,omitempty"`
	Text      string         `yaml:"text" mapstructure:"text" json:"text,omitempty" gorm:"column:text" bson:"text,omitempty" dynamodbav:"text,omitempty" firestore:"text,omitempty"`
	Templates []TemplateNode `yaml:"templates" mapstructure:"templates" json:"templates,omitempty" gorm:"column:templates" bson:"templates,omitempty" dynamodbav:"templates,omitempty" firestore:"templates,omitempty"`
}
type TStatement struct {
	Query  string        `yaml:"query" mapstructure:"query" json:"query,omitempty" gorm:"column:query" bson:"query,omitempty" dynamodbav:"query,omitempty" firestore:"query,omitempty"`
	Params []interface{} `yaml:"params" mapstructure:"params" json:"params,omitempty" gorm:"column:params" bson:"params,omitempty" dynamodbav:"params,omitempty" firestore:"params,omitempty"`
	Index  int           `yaml:"index" mapstructure:"index" json:"index,omitempty" gorm:"column:index" bson:"index,omitempty" dynamodbav:"index,omitempty" firestore:"index,omitempty"`
}

func LoadTemplates(trim func(string) string, files ...string) (map[string]*Template, error) {
	if len(files) == 0 {
		return loadTemplates(trim, "configs/query.xml")
	}
	return loadTemplates(trim, files...)
}
func loadTemplates(trim func(string) string, files ...string) (map[string]*Template, error) {
	l := len(files)
	f0, er0 := ReadFile(files[0])
	if er0 != nil {
		return nil, er0
	}
	if trim != nil {
		f0 = trim(f0)
	}
	templates, er0 := BuildTemplates(f0)
	if er0 != nil {
		return nil, er0
	}
	if l >= 2 {
		for i := 1; i < l; i++ {
			file, err := ReadFile(files[i])
			if err != nil {
				return templates, err
			}
			sub, er := BuildTemplates(file)
			if er0 != nil {
				return templates, er
			}
			for key, element := range sub {
				templates[key] = element
			}
		}
	}
	return templates, nil
}
func BuildTemplates(stream string) (map[string]*Template, error) {
	data := []byte(stream)
	buf := bytes.NewBuffer(data)
	dec := xml.NewDecoder(buf)
	ns := make([]TemplateNode, 0)
	ts := make(map[string]*Template)
	texts := make([]string, 0)
	start := false
	id := ""
	for {
		token, er0 := dec.Token()
		if token == nil {
			break
		}
		if er0 != nil {
			return nil, er0
		}
		switch element := token.(type) {
		case xml.CharData:
			if start == true {
				s := string([]byte(element))
				if !isEmptyNode(s) {
					n := TemplateNode{Type: "text", Text: s}
					texts = append(texts, s)
					n.Format = buildFormat(n.Text)
					ns = append(ns, n)
				}
			}
		case xml.EndElement:
			n := element.Name.Local
			if n == "select" || n == "insert" || n == "update" || n == "delete" {
				t := Template{Id: id}
				t.Text = strings.Join(texts, " ")
				t.Templates = ns
				ts[id] = &t
				ns = make([]TemplateNode, 0)
				start = false
			}
		case xml.StartElement:
			n := element.Name.Local
			if n == "select" || n == "insert" || n == "update" || n == "delete" {
				id = getValue(element.Attr, "id")
				texts = make([]string, 0)
				start = true
			} else {
				if element.Name.Local == "if" {
					test := getValue(element.Attr, "test")
					if len(test) > 0 {
						n := buildIf(test)
						if n != nil {
							n.Array = getValue(element.Attr, "array")
							n.Prefix = getValue(element.Attr, "prefix")
							n.Suffix = getValue(element.Attr, "suffix")
							n.Separator = getValue(element.Attr, "separator")
							sub, er1 := dec.Token()
							if er1 != nil {
								return nil, er1
							}
							switch inner := sub.(type) {
							case xml.CharData:
								s2 := string([]byte(inner))
								n.Text = s2
								n.Format = buildFormat(n.Text)
								texts = append(texts, s2)
							}
							ns = append(ns, *n)
						}
					}
				} else {
					if isEmptyNode(element.Name.Local) {
						property := getValue(element.Attr, "property")
						v := getValue(element.Attr, "value")
						array := getValue(element.Attr, "array")
						prefix := getValue(element.Attr, "prefix")
						suffix := getValue(element.Attr, "suffix")
						separator := getValue(element.Attr, "separator")
						n := TemplateNode{Type: element.Name.Local, Property: property, Value: v, Array: array, Prefix: prefix, Suffix: suffix, Separator: separator}
						sub, er1 := dec.Token()
						if er1 != nil {
							return nil, er1
						}
						switch inner := sub.(type) {
						case xml.CharData:
							s2 := string([]byte(inner))
							n.Text = s2
							n.Format = buildFormat(n.Text)
							texts = append(texts, s2)
						}
						ns = append(ns, n)
					}
				}
			}
		}
	}
	return ts, nil
}
func isEmptyNode(s string) bool {
	v := strings.Replace(s, "\n", " ", -1)
	v = strings.Replace(v, "\r", " ", -1)
	v = strings.TrimSpace(s)
	return len(v) == 0
}

func BuildTemplate(stream string) (*Template, error) {
	data := []byte(stream)
	buf := bytes.NewBuffer(data)
	dec := xml.NewDecoder(buf)
	ns := make([]TemplateNode, 0)
	texts := make([]string, 0)
	for {
		token, er0 := dec.Token()
		if token == nil {
			break
		}
		if er0 != nil {
			return nil, er0
		}
		switch element := token.(type) {
		case xml.CharData:
			s := string([]byte(element))
			if s != "\n" {
				n := TemplateNode{Type: "text", Text: s}
				texts = append(texts, s)
				n.Format = buildFormat(n.Text)
				ns = append(ns, n)
			}
		case xml.StartElement:
			if element.Name.Local == "if" {
				test := getValue(element.Attr, "test")
				if len(test) > 0 {
					n := buildIf(test)
					if n != nil {
						n.Array = getValue(element.Attr, "array")
						n.Prefix = getValue(element.Attr, "prefix")
						n.Suffix = getValue(element.Attr, "suffix")
						n.Separator = getValue(element.Attr, "separator")
						sub, er1 := dec.Token()
						if er1 != nil {
							return nil, er1
						}
						switch inner := sub.(type) {
						case xml.CharData:
							s2 := string([]byte(inner))
							n.Text = s2
							n.Format = buildFormat(n.Text)
							texts = append(texts, s2)
						}
						ns = append(ns, *n)
					}
				}
			} else {
				if isEmptyNode(element.Name.Local) {
					property := getValue(element.Attr, "property")
					v := getValue(element.Attr, "value")
					array := getValue(element.Attr, "array")
					prefix := getValue(element.Attr, "prefix")
					suffix := getValue(element.Attr, "suffix")
					separator := getValue(element.Attr, "separator")
					n := TemplateNode{Type: element.Name.Local, Property: property, Value: v, Array: array, Prefix: prefix, Suffix: suffix, Separator: separator}
					sub, er1 := dec.Token()
					if er1 != nil {
						return nil, er1
					}
					switch inner := sub.(type) {
					case xml.CharData:
						s2 := string([]byte(inner))
						n.Text = s2
						n.Format = buildFormat(n.Text)
						texts = append(texts, s2)
					}
					ns = append(ns, n)
				}
			}
		}
	}
	t := Template{}
	t.Text = strings.Join(texts, " ")
	t.Templates = ns
	return &t, nil
}
func getValue(attrs []xml.Attr, name string) string {
	if len(attrs) <= 0 {
		return ""
	}
	for _, attr := range attrs {
		if attr.Name.Local == name {
			return attr.Value
		}
	}
	return ""
}
func buildFormat(str string) StringFormat {
	str2 := str
	str2b := str
	var str3 string
	texts := make([]string, 0)
	parameters := make([]Parameter, 0)
	var from, i, j int
	for {
		i = strings.Index(str2b, "{")
		if i >= 0 {
			str3 = str2b[i+1:]
			j = strings.Index(str3, "}")
			if j >= 0 {
				pro := str2b[i+1 : i+j+1]
				if isValidProperty(pro) {
					p := Parameter{}
					p.Name = pro
					if i >= 1 {
						var chr = string(str2b[i-1])
						if chr == "#" {
							texts = append(texts, str2[:from+i-1])
							p.Type = "param"
						} else if chr == "$" {
							texts = append(texts, str2[:from+i-1])
							p.Type = "text"
						} else {
							texts = append(texts, str2[:from+i])
							p.Type = "text"
						}
					} else {
						texts = append(texts, str2[:from+i])
						p.Type = "text"
					}
					parameters = append(parameters, p)
					from = from + i + j + 2
					str2 = str2[from:]
					str2b = str2
					from = 0
				} else {
					from = i + 1
					str2b = str2[i+1:]
				}
			} else {
				from = i + 1
				str2b = str2[from:]
			}
		} else {
			texts = append(texts, str2)
			break
		}
	}
	f := StringFormat{}
	f.Texts = texts
	f.Parameters = parameters
	return f
}
func RenderTemplateNodes(obj map[string]interface{}, templateNodes []TemplateNode) []TemplateNode {
	nodes := make([]TemplateNode, 0)
	for _, sub := range templateNodes {
		t := sub.Type
		if t == TypeText {
			nodes = append(nodes, sub)
		} else {
			attr := valueOf(obj, sub.Property)
			if t == TypeIsNotNull {
				if attr != nil {
					vo := reflect.Indirect(reflect.ValueOf(attr))
					if vo.Kind() == reflect.Slice {
						if vo.Len() > 0 {
							nodes = append(nodes, sub)
						}
					} else {
						nodes = append(nodes, sub)
					}
				} else {
					vo := reflect.Indirect(reflect.ValueOf(attr))
					if vo.Kind() == reflect.Slice {
						if vo.Len() > 0 {
							nodes = append(nodes, sub)
						}
					}
				}
			} else if t == TypeIsNull {
				if attr == nil {
					nodes = append(nodes, sub)
				} else {
					vo := reflect.Indirect(reflect.ValueOf(attr))
					if vo.Kind() == reflect.Slice {
						if vo.Len() == 0 {
							nodes = append(nodes, sub)
						}
					}
				}
			} else if t == TypeIsEqual {
				if attr != nil {
					s := fmt.Sprintf("%v", attr)
					if sub.Value == s {
						nodes = append(nodes, sub)
					}
				}
			} else if t == TypeIsNotEqual {
				if attr != nil {
					s := fmt.Sprintf("%v", attr)
					if sub.Value != s {
						nodes = append(nodes, sub)
					}
				}
			} else if t == TypeIsEmpty {
				if attr != nil {
					s := fmt.Sprintf("%v", attr)
					if len(s) == 0 {
						nodes = append(nodes, sub)
					}
				}
			} else if t == TypeIsNotEmpty {
				if attr != nil {
					s := fmt.Sprintf("%v", attr)
					if len(s) > 0 {
						nodes = append(nodes, sub)
					}
				}
			}
		}
	}
	return nodes
}
func isValidProperty(v string) bool {
	var len = len(v) - 1
	for i := 0; i <= len; i++ {
		var chr = string(v[i])
		if !((chr >= "0" && chr <= "9") || (chr >= "A" && chr <= "Z") || (chr >= "a" && chr <= "z") || chr == "_" || chr == ".") {
			return false
		}
	}
	return true
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
func buildIf(t string) *TemplateNode {
	i := strings.Index(t, "!=")
	if i > 0 {
		s1 := strings.TrimSpace(t[0:i])
		s2 := strings.TrimSpace(t[i+2:])
		if len(s1) > 0 {
			if s2 == "null" {
				return &TemplateNode{Type: "isNotNull", Property: s1}
			} else {
				return &TemplateNode{Type: "isNotEqual", Property: s1, Value: trimQ(s2)}
			}
		}
	} else {
		i = strings.Index(t, "==")
		if i > 0 {
			s1 := strings.TrimSpace(t[0:i])
			s2 := strings.TrimSpace(t[i+2:])
			if len(s1) > 0 {
				if s2 == "null" {
					return &TemplateNode{Type: "isNull", Property: s1}
				} else {
					return &TemplateNode{Type: "isEqual", Property: s1, Value: trimQ(s2)}
				}
			}
		}
	}
	return nil
}
func trimQ(s string) string {
	if strings.HasPrefix(s, "'") {
		s = s[1:]
	} else if strings.HasPrefix(s, `"`) {
		s = s[1:]
	}
	if strings.HasSuffix(s, "'") {
		s = s[len(s)-1:]
	} else if strings.HasSuffix(s, `"`) {
		s = s[len(s)-1:]
	}
	return s
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
