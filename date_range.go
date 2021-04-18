package search

import "time"

type DateRange struct {
	StartDate *time.Time `mapstructure:"startDate" json:"startDate,omitempty" gorm:"column:startdate" bson:"startDate,omitempty" dynamodbav:"startDate,omitempty" firestore:"startDate,omitempty"`
	EndDate   *time.Time `mapstructure:"endDate" json:"endDate,omitempty" gorm:"column:endDate" bson:"endDate,omitempty" dynamodbav:"endDate,omitempty" firestore:"endDate,omitempty"`
}
