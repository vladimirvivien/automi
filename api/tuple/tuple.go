package tuple

type Tuple struct {
	values []interface{}
	fields []string
}

func New(vals ...interface{}) Tuple {
	return Tuple{
		fields: make([]string, 0),
		values: vals,
	}
}

func (t Tuple) WithValues(vals ...interface{}) Tuple {
	t.values = append(t.values, vals...)
	return t
}

func (t Tuple) WithFields(fds ...string) Tuple {
	t.fields = append(t.fields, fds...)
	return t
}

func (t Tuple) Get(i int) interface{} {
	return t.values[i]
}

func (t Tuple) GetAt(pos ...int) []interface{} {
	result := make([]interface{}, len(pos))
	for _, i := range pos {
		result = append(result, t.values[i])
	}
	return result
}

func (t Tuple) AsInt(i int) int {
	return t.values[i].(int)
}

func (t Tuple) AsString(i int) string {
	return t.values[i].(string)
}

func (t Tuple) Values() []interface{} {
	result := make([]interface{}, len(t.values))
	copy(result, t.values)
	return result
}

func (t Tuple) Fields() []string {
	result := make([]string, len(t.fields))
	copy(result, t.fields)
	return result
}
