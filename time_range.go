package search

import "time"

type TimeRange struct {
	StartTime *time.Time `json:"startTime,omitempty" bson:"startTime,omitempty" gorm:"column:starttime"`
	EndTime   *time.Time `json:"endTime,omitempty" bson:"endTime,omitempty" gorm:"column:endtime"`
}
