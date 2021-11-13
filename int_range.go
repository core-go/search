package search

type IntRange struct {
	Min     *int `mapstructure:"min" json:"min,omitempty" gorm:"column:min" bson:"min,omitempty" dynamodbav:"min,omitempty" firestore:"min,omitempty"`
	Max     *int `mapstructure:"max" json:"max,omitempty" gorm:"column:max" bson:"max,omitempty" dynamodbav:"max,omitempty" firestore:"max,omitempty"`
	Bottom  *int `mapstructure:"bottom" json:"bottom,omitempty" gorm:"column:bottom" bson:"bottom,omitempty" dynamodbav:"bottom,omitempty" firestore:"bottom,omitempty"`
	Top     *int `mapstructure:"top" json:"top,omitempty" gorm:"column:top" bson:"top,omitempty" dynamodbav:"top,omitempty" firestore:"top,omitempty"`
	Floor   *int `mapstructure:"floor" json:"floor,omitempty" gorm:"column:floor" bson:"floor,omitempty" dynamodbav:"floor,omitempty" firestore:"floor,omitempty"`
	Ceiling *int `mapstructure:"ceiling" json:"ceiling,omitempty" gorm:"column:ceiling" bson:"ceiling,omitempty" dynamodbav:"ceiling,omitempty" firestore:"ceiling,omitempty"`
	Lower   *int `mapstructure:"lower" json:"lower,omitempty" gorm:"column:lower" bson:"lower,omitempty" dynamodbav:"lower,omitempty" firestore:"lower,omitempty"`
	Upper   *int `mapstructure:"upper" json:"upper,omitempty" gorm:"column:upper" bson:"upper,omitempty" dynamodbav:"upper,omitempty" firestore:"upper,omitempty"`
}
