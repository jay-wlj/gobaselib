package base

import (
	"fmt"
	"testing"
)

func TestUnique(t *testing.T) {
	fmt.Println(UniqueInt64Slice([]int64{1, 2, 1, 3, 4, 1, 2, 4, 3}))
}

func TestIntSliceToString(t *testing.T) {
	fmt.Println(IntSliceToString([]int{1, 3, 4, 25, 3456, 234}, "-"))
}
