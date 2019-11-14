package search

type SearchResult struct {
	Results   interface{} `json:"results,omitempty" bson:"results,omitempty" gorm:"column:results"`
	ItemTotal int         `json:"itemTotal,omitempty" bson:"itemTotal,omitempty" gorm:"column:itemtotal"`
	LastPage  bool        `json:"lastPage,omitempty" bson:"lastPage,omitempty" gorm:"column:lastpage"`
}
