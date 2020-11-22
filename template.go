package search

import (
	"bytes"
	"context"
	"encoding/xml"
	"strings"
)

const (
	TypeText       = "text"
	TypeNotEmpty   = "notEmpty"
	TypeEmpty      = "empty"
	TypeEqual      = "equal"
	TypeNotEqual   = "notEqual"
	ParamText      = "text"
	ParamParameter = "param"
)

type StringFormat struct {
	Texts      string `mapstructure:"texts" json:"texts,omitempty" gorm:"column:texts" bson:"texts,omitempty" dynamodbav:"texts,omitempty" firestore:"texts,omitempty"`
	Parameters string `mapstructure:"parameters" json:"parameters,omitempty" gorm:"column:parameters" bson:"parameters,omitempty" dynamodbav:"parameters,omitempty" firestore:"parameters,omitempty"`
}
type Parameter struct {
	Name string `mapstructure:"name" json:"name,omitempty" gorm:"column:name" bson:"name,omitempty" dynamodbav:"name,omitempty" firestore:"name,omitempty"`
	Type string `mapstructure:"type" json:"type,omitempty" gorm:"column:type" bson:"type,omitempty" dynamodbav:"type,omitempty" firestore:"type,omitempty"`
}
type TemplateNode struct {
	Type   string `mapstructure:"type" json:"type,omitempty" gorm:"column:type" bson:"type,omitempty" dynamodbav:"type,omitempty" firestore:"type,omitempty"`
	Text   string `mapstructure:"text" json:"text,omitempty" gorm:"column:text" bson:"text,omitempty" dynamodbav:"text,omitempty" firestore:"text,omitempty"`
	Name   string `mapstructure:"name" json:"name,omitempty" gorm:"column:name" bson:"name,omitempty" dynamodbav:"name,omitempty" firestore:"name,omitempty"`
	Encode string `mapstructure:"encode" json:"encode,omitempty" gorm:"column:encode" bson:"encode,omitempty" dynamodbav:"encode,omitempty" firestore:"encode,omitempty"`
	Value  string `mapstructure:"value" json:"value,omitempty" gorm:"column:value" bson:"value,omitempty" dynamodbav:"value,omitempty" firestore:"value,omitempty"`
}
type Template struct {
	Text      string         `mapstructure:"text" json:"text,omitempty" gorm:"column:text" bson:"text,omitempty" dynamodbav:"text,omitempty" firestore:"text,omitempty"`
	Templates []TemplateNode `mapstructure:"templates" json:"templates,omitempty" gorm:"column:templates" bson:"templates,omitempty" dynamodbav:"templates,omitempty" firestore:"templates,omitempty"`
}

type TemplateBuilder interface {
	Build(ctx context.Context, stream string) (*Template, error)
}
type XmlTemplateBuilder struct {
}
func NewXmlTemplateBuilder() *XmlTemplateBuilder{
	return &XmlTemplateBuilder{}
}
func (b *XmlTemplateBuilder) Build(ctx context.Context, stream string) (*Template, error) {
	return BuildTemplate(ctx, stream)
}
func BuildTemplate(ctx context.Context, stream string) (*Template, error) {
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
				ns = append(ns, n)
			}
		case xml.StartElement:
			if element.Name.Local == "notEmpty" {
				encode := GetValue(element.Attr, "encode")
				n := TemplateNode{ Type: "notEmpty", Encode: encode}
				sub, er1 := dec.Token()
				if er1 != nil {
					return nil, er1
				}
				switch inner := sub.(type) {
				case xml.CharData:
					s2 := string([]byte(inner))
					n.Text = s2
					texts = append(texts, s2)
				}
				ns = append(ns, n)
			} else if element.Name.Local == "empty" {
				encode := GetValue(element.Attr, "encode")
				n := TemplateNode{Type: "empty", Encode: encode}
				sub, er1 := dec.Token()
				if er1 != nil {
					return nil, er1
				}
				switch inner := sub.(type) {
				case xml.CharData:
					s2 := string([]byte(inner))
					n.Text = s2
					texts = append(texts, s2)
				}
				ns = append(ns, n)
			} else if element.Name.Local == "equal" {
				encode := GetValue(element.Attr, "encode")
				v := GetValue(element.Attr, "value")
				n := TemplateNode{Type: "equal", Encode: encode, Value: v}
				sub, er1 := dec.Token()
				if er1 != nil {
					return nil, er1
				}
				switch inner := sub.(type) {
				case xml.CharData:
					s2 := string([]byte(inner))
					n.Text = s2
					texts = append(texts, s2)
				}
				ns = append(ns, n)
			} else if element.Name.Local == "notEqual" {
				encode := GetValue(element.Attr, "encode")
				v := GetValue(element.Attr, "value")
				n := TemplateNode{Type: "notEqual", Encode: encode, Value: v}
				sub, er1 := dec.Token()
				if er1 != nil {
					return nil, er1
				}
				switch inner := sub.(type) {
				case xml.CharData:
					s2 := string([]byte(inner))
					n.Text = s2
					texts = append(texts, s2)
				}
				ns = append(ns, n)
			}
		}
	}
	t := Template{}
	t.Text = strings.Join(texts, " ")
	t.Templates = ns
	return &t, nil
}
func GetValue(attrs []xml.Attr, name string) string {
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
