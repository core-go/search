package search

import "time"

type DateRange struct {
	Min     *time.Time `mapstructure:"min" json:"min,omitempty" gorm:"column:startdate" bson:"min,omitempty" dynamodbav:"min,omitempty" firestore:"min,omitempty"`
	Max     *time.Time `mapstructure:"max" json:"max,omitempty" gorm:"column:max" bson:"max,omitempty" dynamodbav:"max,omitempty" firestore:"max,omitempty"`
	Bottom  *time.Time `mapstructure:"bottom" json:"bottom,omitempty" gorm:"column:bottom" bson:"bottom,omitempty" dynamodbav:"bottom,omitempty" firestore:"bottom,omitempty"`
	Top     *time.Time `mapstructure:"top" json:"top,omitempty" gorm:"column:top" bson:"top,omitempty" dynamodbav:"top,omitempty" firestore:"top,omitempty"`
	Floor   *time.Time `mapstructure:"floor" json:"floor,omitempty" gorm:"column:floor" bson:"floor,omitempty" dynamodbav:"floor,omitempty" firestore:"floor,omitempty"`
	Ceiling *time.Time `mapstructure:"ceiling" json:"ceiling,omitempty" gorm:"column:ceiling" bson:"ceiling,omitempty" dynamodbav:"ceiling,omitempty" firestore:"ceiling,omitempty"`
	Lower   *time.Time `mapstructure:"lower" json:"lower,omitempty" gorm:"column:lower" bson:"lower,omitempty" dynamodbav:"lower,omitempty" firestore:"lower,omitempty"`
	Upper   *time.Time `mapstructure:"upper" json:"upper,omitempty" gorm:"column:upper" bson:"upper,omitempty" dynamodbav:"upper,omitempty" firestore:"upper,omitempty"`
}
