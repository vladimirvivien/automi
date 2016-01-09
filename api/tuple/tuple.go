package tuple

type Pair [2]interface{}

func MakePair(k, v interface{}) Pair {
	return [2]interface{}{k, v}
}
