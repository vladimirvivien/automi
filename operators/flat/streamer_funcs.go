package flat

func Slice[IN []ITEM, ITEM any]() *SliceStreamer[IN, ITEM] {
	return NewSliceStreamer[IN]()
}

func Map[IN map[KEY]ITEM, KEY comparable, ITEM any]() *MapStreamer[IN, KEY, ITEM] {
	return NewMapStreamer[IN]()
}
