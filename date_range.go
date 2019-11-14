package search

import "time"

type DateRange struct {
	StartDate *time.Time `json:"startDate,omitempty" bson:"startDate,omitempty" gorm:"column:startdate"`
	EndDate   *time.Time `json:"endDate,omitempty" bson:"endDate,omitempty" gorm:"column:enddate"`
}
