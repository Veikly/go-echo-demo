package domain

// 筛选类型
type FilterKind string

const (
	FilterEqual FilterKind = "equal"
	FilterRange FilterKind = "range"
)

// 游标取值类型
type CursorValueType string

const (
	CursorTime   CursorValueType = "time"
	CursorString CursorValueType = "string"
	CursorNumber CursorValueType = "number"
)

type FilterSpec struct {
	Name              string
	Field             string
	Kind              FilterKind
	RequireBothBounds bool
	AllowOpenStart    bool
	AllowOpenEnd      bool
}

type SortDirection string

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)

type SortSpec struct {
	Key   string
	Field string
	Dir   SortDirection
	Type  CursorValueType
}

type QueryProfile struct {
	Key         string
	Resource    string
	Filters     map[string]FilterSpec
	Sorts       map[string]SortSpec
	DefaultSort string
	MaxLimit    int
}
