package search

import "reflect"

type Filter struct {
	PageIndex     int64 `yaml:"page_index" mapstructure:"page_index" json:"pageIndex,omitempty" gorm:"column:pageindex" bson:"pageIndex,omitempty" dynamodbav:"pageIndex,omitempty" firestore:"pageIndex,omitempty"`
	PageSize      int64 `yaml:"page_size" mapstructure:"page_size" json:"pageSize,omitempty" gorm:"column:pagesize" bson:"pageSize,omitempty" dynamodbav:"pageSize,omitempty" firestore:"pageSize,omitempty"`
	FirstPageSize int64 `yaml:"first_page_size" mapstructure:"first_page_size" json:"firstPageSize,omitempty" gorm:"column:firstpagesize" bson:"firstPageSize,omitempty" dynamodbav:"firstPageSize,omitempty" firestore:"firstPageSize,omitempty"`

	Page          int64    `yaml:"page" mapstructure:"page" json:"page,omitempty" gorm:"column:pageindex" bson:"page,omitempty" dynamodbav:"page,omitempty" firestore:"page,omitempty"`
	Limit         int64    `yaml:"limit" mapstructure:"limit" json:"limit,omitempty" gorm:"column:limit" bson:"limit,omitempty" dynamodbav:"limit,omitempty" firestore:"limit,omitempty"`
	FirstLimit    int64    `yaml:"first_limit" mapstructure:"first_limit" json:"firstLimit,omitempty" gorm:"column:firstlimit" bson:"firstLimit,omitempty" dynamodbav:"firstLimit,omitempty" firestore:"firstLimit,omitempty"`
	Fields        []string `yaml:"fields" mapstructure:"fields" json:"fields,omitempty" gorm:"column:fields" bson:"fields,omitempty" dynamodbav:"fields,omitempty" firestore:"fields,omitempty"`
	Sort          string   `yaml:"sort" mapstructure:"sort" json:"sort,omitempty" gorm:"column:sortfield" bson:"sort,omitempty" dynamodbav:"sort,omitempty" firestore:"sort,omitempty"`
	CurrentUserId string   `yaml:"current_user_id" mapstructure:"current_user_id" json:"currentUserId,omitempty" gorm:"column:currentuserid" bson:"currentUserId,omitempty" dynamodbav:"currentUserId,omitempty" firestore:"currentUserId,omitempty"`
	Q             string   `yaml:"q" mapstructure:"q" json:"q,omitempty" gorm:"column:q" bson:"q,omitempty" dynamodbav:"q,omitempty" firestore:"q,omitempty"`
	Excluding     []string `yaml:"excluding" mapstructure:"excluding" json:"excluding,omitempty" gorm:"column:excluding" bson:"excluding,omitempty" dynamodbav:"excluding,omitempty" firestore:"excluding,omitempty"`
	Next          string   `yaml:"next" mapstructure:"next" json:"next,omitempty" gorm:"column:next" bson:"next,omitempty" dynamodbav:"next,omitempty" firestore:"next,omitempty"`
	RefId         string   `yaml:"ref_id" mapstructure:"ref_id" json:"refId,omitempty" gorm:"column:refid" bson:"refId,omitempty" dynamodbav:"refId,omitempty" firestore:"refId,omitempty"`
	NextPageToken string   `yaml:"next_page_token" mapstructure:"next_page_token" json:"nextPageToken,omitempty" gorm:"column:nextpagetoken" bson:"nextPageToken,omitempty" dynamodbav:"nextPageToken,omitempty" firestore:"nextPageToken,omitempty"`
}
type Result struct {
	List          interface{} `yaml:"list" mapstructure:"list" json:"list,omitempty" gorm:"column:list" bson:"list,omitempty" dynamodbav:"list,omitempty" firestore:"list,omitempty"`
	Total         int64       `yaml:"total" mapstructure:"total" json:"total,omitempty" gorm:"column:total" bson:"total,omitempty" dynamodbav:"total,omitempty" firestore:"total,omitempty"`
	Next          string      `yaml:"next" mapstructure:"next" json:"next,omitempty" gorm:"column:next" bson:"next,omitempty" dynamodbav:"next,omitempty" firestore:"next,omitempty"`
}

func GetFilter(m interface{}) *Filter {
	if sModel, ok := m.(*Filter); ok {
		return sModel
	} else {
		value := reflect.Indirect(reflect.ValueOf(m))
		numField := value.NumField()
		for i := 0; i < numField; i++ {
			if sModel1, ok := value.Field(i).Interface().(*Filter); ok {
				return sModel1
			}
		}
	}
	return nil
}
