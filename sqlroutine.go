package goroutine

type Query struct {
	Select []string
	In     []interface{}
	Range  []interface{}
}

func (q *Query) MySQL() string {
	return ""
}

func (q *Query) SQL() string {
	return ""
}
func (q *Query) Postgres() string {
	return ""
}
func (q *Query) MongoDB() string {
	return ""
}

func (q *Query) ElasticSearchs() string {
	return ""
}
