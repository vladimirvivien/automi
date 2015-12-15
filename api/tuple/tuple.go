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

func (t Tuple) Values() []interface{} {
	return t.values
}

func (t Tuple) ValuesByKeys(keys ...string) []interface{} {
	return nil
}

func (t Tuple) ValuesByPos(pos ...int) []interface{} {
	return nil
}

func (t Tuple) Fields() []string {
	return t.fields
}
