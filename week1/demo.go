package week1

import (
	"fmt"
)

// 作业：实现切片的删除操作

// 实现删除切片特定下标元素的方法。
// 要求一：能够实现删除操作就可以。
// 要求二：考虑使用比较高性能的实现。
// 要求三：改造为泛型方法
// 要求四：支持缩容，并且设计缩容机制。

func SliceDelete[T any](s []T, i int) []T {
	if i < 0 || i > len(s) {
		return s
	}

	// delete slice
	copy(s[i:], s[i+1:])
	s = s[:len(s)-1]

	// 缩容
	if cap(s) > 2*len(s) {
		newSlice := make([]T, len(s))
		copy(newSlice, s)
		return newSlice
	}

	return s
}

func Demo() {
	// delete int
	intSlice := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	intSlice = SliceDelete(intSlice, 0)
	fmt.Printf("intSlice: %v, len: %d, cap: %d\n", intSlice, len(intSlice), cap(intSlice))

	// 缩容
	for i := 0; i < 6; i++ {
		intSlice = SliceDelete(intSlice, len(intSlice)-1)
		fmt.Printf("intSlice: %v, len: %d, cap: %d\n", intSlice, len(intSlice), cap(intSlice))
	}

	// 扩容
	intSlice = append(intSlice, 11, 12)
	fmt.Printf("intSlice: %v, len: %d, cap: %d\n", intSlice, len(intSlice), cap(intSlice))

	// delete string
	strSlice := []string{"a", "b", "c", "d", "e"}
	strSlice = SliceDelete(strSlice, 2)
	fmt.Printf("strSlice: %v, len: %d, cap: %d\n", strSlice, len(strSlice), cap(strSlice))

}
