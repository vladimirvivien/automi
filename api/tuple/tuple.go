package tuple

type Pair[T1, T2 any] struct {
	Val1 T1
	Val2 T2
}

type Triple[T1, T2, T3 any] struct {
	Val1 T1
	Val2 T2
	Val3 T3
}

type Quad[T1, T2, T3, T4 any] struct {
	Val1 T1
	Val2 T2
	Val3 T3
	Val4 T4
}

type Quint[T1, T2, T3, T4, T5 any] struct {
	Val1 T1
	Val2 T2
	Val3 T3
	Val4 T4
	Val5 T5
}

type Sext[T1, T2, T3, T4, T5, T6 any] struct {
	Val1 T1
	Val2 T2
	Val3 T3
	Val4 T4
	Val5 T5
	Val6 T6
}

type Sept[T1, T2, T3, T4, T5, T6, T7 any] struct {
	Val1 T1
	Val2 T2
	Val3 T3
	Val4 T4
	Val5 T5
	Val6 T6
	Val7 T7
}

type Oct[T1, T2, T3, T4, T5, T6, T7, T8 any] struct {
	Val1 T1
	Val2 T2
	Val3 T3
	Val4 T4
	Val5 T5
	Val6 T6
	Val7 T7
	Val8 T8
}

type Non[T1, T2, T3, T4, T5, T6, T7, T8, T9 any] struct {
	Val1 T1
	Val2 T2
	Val3 T3
	Val4 T4
	Val5 T5
	Val6 T6
	Val7 T7
	Val8 T8
	Val9 T9
}

type Dec[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10 any] struct {
	Val1  T1
	Val2  T2
	Val3  T3
	Val4  T4
	Val5  T5
	Val6  T6
	Val7  T7
	Val8  T8
	Val9  T9
	Val10 T10
}
