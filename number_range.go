package search

type NumberRange struct {
	Min   *float64 `json:"min,omitempty" bson:"min,omitempty" gorm:"column:min"`
	Max   *float64 `json:"max,omitempty" bson:"max,omitempty" gorm:"column:max"`
	Lower *float64 `json:"lower,omitempty" bson:"lower,omitempty" gorm:"column:lower"`
	Upper *float64 `json:"upper,omitempty" bson:"upper,omitempty" gorm:"column:upper"`
}
