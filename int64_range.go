package search

type Int64Range struct {
	Min     *int64 `mapstructure:"min" json:"min,omitempty" gorm:"column:min" bson:"min,omitempty" dynamodbav:"min,omitempty" firestore:"min,omitempty"`
	Max     *int64 `mapstructure:"max" json:"max,omitempty" gorm:"column:max" bson:"max,omitempty" dynamodbav:"max,omitempty" firestore:"max,omitempty"`
	Bottom  *int64 `mapstructure:"bottom" json:"bottom,omitempty" gorm:"column:bottom" bson:"bottom,omitempty" dynamodbav:"bottom,omitempty" firestore:"bottom,omitempty"`
	Top     *int64 `mapstructure:"top" json:"top,omitempty" gorm:"column:top" bson:"top,omitempty" dynamodbav:"top,omitempty" firestore:"top,omitempty"`
	Floor   *int64 `mapstructure:"floor" json:"floor,omitempty" gorm:"column:floor" bson:"floor,omitempty" dynamodbav:"floor,omitempty" firestore:"floor,omitempty"`
	Ceiling *int64 `mapstructure:"ceiling" json:"ceiling,omitempty" gorm:"column:ceiling" bson:"ceiling,omitempty" dynamodbav:"ceiling,omitempty" firestore:"ceiling,omitempty"`
	Lower   *int64 `mapstructure:"lower" json:"lower,omitempty" gorm:"column:lower" bson:"lower,omitempty" dynamodbav:"lower,omitempty" firestore:"lower,omitempty"`
	Upper   *int64 `mapstructure:"upper" json:"upper,omitempty" gorm:"column:upper" bson:"upper,omitempty" dynamodbav:"upper,omitempty" firestore:"upper,omitempty"`
}
