package search

type Int64Range struct {
	Min   *int64 `json:"min,omitempty" bson:"min,omitempty" gorm:"column:min"`
	Max   *int64 `json:"max,omitempty" bson:"max,omitempty" gorm:"column:max"`
	Lower *int64 `json:"lower,omitempty" bson:"lower,omitempty" gorm:"column:lower"`
	Upper *int64 `json:"upper,omitempty" bson:"upper,omitempty" gorm:"column:upper"`
}
