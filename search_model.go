package search

import "reflect"

type SearchModel struct {
	Page          int64                    `mapstructure:"page" json:"page,omitempty" gorm:"column:pageindex" bson:"page,omitempty" dynamodbav:"page,omitempty" firestore:"page,omitempty"`
	Limit         int64                    `mapstructure:"limit" json:"limit,omitempty" gorm:"column:limit" bson:"limit,omitempty" dynamodbav:"limit,omitempty" firestore:"limit,omitempty"`
	FirstLimit    int64                    `mapstructure:"first_limit" json:"firstLimit,omitempty" gorm:"column:firstlimit" bson:"firstLimit,omitempty" dynamodbav:"firstLimit,omitempty" firestore:"firstLimit,omitempty"`
	Fields        []string                 `mapstructure:"fields" json:"fields,omitempty" gorm:"column:fields" bson:"fields,omitempty" dynamodbav:"fields,omitempty" firestore:"fields,omitempty"`
	Sort          string                   `mapstructure:"sort" json:"sort,omitempty" gorm:"column:sortfield" bson:"sort,omitempty" dynamodbav:"sort,omitempty" firestore:"sort,omitempty"`
	CurrentUserId string                   `mapstructure:"current_user_id" json:"currentUserId,omitempty" gorm:"column:currentuserid" bson:"currentUserId,omitempty" dynamodbav:"currentUserId,omitempty" firestore:"currentUserId,omitempty"`
	Keyword       string                   `mapstructure:"keyword" json:"keyword,omitempty" gorm:"column:keyword" bson:"keyword,omitempty" dynamodbav:"keyword,omitempty" firestore:"keyword,omitempty"`
	Excluding     map[string][]interface{} `mapstructure:"excluding" json:"excluding,omitempty" gorm:"column:excluding" bson:"excluding,omitempty" dynamodbav:"excluding,omitempty" firestore:"excluding,omitempty"`
}

func IsExtendedFromSearchModel(searchModelType reflect.Type) bool {
	var searchModel = reflect.New(searchModelType).Interface()
	if _, ok := searchModel.(*SearchModel); ok {
		return false
	} else {
		value := reflect.Indirect(reflect.ValueOf(searchModel))
		numField := value.NumField()
		for i := 0; i < numField; i++ {
			if _, ok := value.Field(i).Interface().(*SearchModel); ok {
				return true
			}
		}
	}
	return false
}
