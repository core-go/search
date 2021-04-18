# Search Service for Golang
- Search Model
- Search Result With Config
- Search Service
- Search Handler

## Installation
Please make sure to initialize a Go module before installing common-go/search:

```shell
go get -u github.com/common-go/search
```

Import:

```go
import "github.com/common-go/search"
```

## Details:
#### search_model.go
```go
package search

type SearchModel struct {
	Page          int64                    `mapstructure:"page" json:"page,omitempty" gorm:"column:pageindex" bson:"page,omitempty" dynamodbav:"page,omitempty" firestore:"page,omitempty"`
	Limit         int64                    `mapstructure:"limit" json:"limit,omitempty" gorm:"column:limit" bson:"limit,omitempty" dynamodbav:"limit,omitempty" firestore:"limit,omitempty"`
	FirstLimit    int64                    `mapstructure:"first_limit" json:"firstLimit,omitempty" gorm:"column:firstlimit" bson:"firstLimit,omitempty" dynamodbav:"firstLimit,omitempty" firestore:"firstLimit,omitempty"`
	Fields        []string                 `mapstructure:"fields" json:"fields,omitempty" gorm:"column:fields" bson:"fields,omitempty" dynamodbav:"fields,omitempty" firestore:"fields,omitempty"`
	Sort          string                   `mapstructure:"sort" json:"sort,omitempty" gorm:"column:sortfield" bson:"sort,omitempty" dynamodbav:"sort,omitempty" firestore:"sort,omitempty"`
	CurrentUserId string                   `mapstructure:"current_user_id" json:"currentUserId,omitempty" gorm:"column:currentuserid" bson:"currentUserId,omitempty" dynamodbav:"currentUserId,omitempty" firestore:"currentUserId,omitempty"`
	Keyword       string                   `mapstructure:"keyword" json:"keyword,omitempty" gorm:"column:keyword" bson:"keyword,omitempty" dynamodbav:"keyword,omitempty" firestore:"keyword,omitempty"`
	Excluding     map[string][]interface{} `mapstructure:"excluding" json:"excluding,omitempty" gorm:"column:excluding" bson:"excluding,omitempty" dynamodbav:"excluding,omitempty" firestore:"excluding,omitempty"`
	RefId         string                   `mapstructure:"refid" json:"refId,omitempty" gorm:"column:refid" bson:"refId,omitempty" dynamodbav:"refId,omitempty" firestore:"refId,omitempty"`
}
```
#### search_service.go
```go
package search

import "context"

type SearchService interface {
	Search(ctx context.Context, searchModel interface{}, results interface{}, pageIndex int64, pageSize int64, options...int64) (int64, error)
}
```
