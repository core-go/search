package search

import "context"

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
	Build(ctx context.Context, stream string) Template
}
