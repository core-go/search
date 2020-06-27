package search

type Int32Range struct {
	Min   *int32 `json:"min,omitempty" bson:"min,omitempty" gorm:"column:min"`
	Max   *int32 `json:"max,omitempty" bson:"max,omitempty" gorm:"column:max"`
	Lower *int32 `json:"lower,omitempty" bson:"lower,omitempty" gorm:"column:lower"`
	Upper *int32 `json:"upper,omitempty" bson:"upper,omitempty" gorm:"column:upper"`
}
