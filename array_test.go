package immutable

import (
  "fmt"
  "os"
  "testing"

  "github.com/stretchr/testify/assert"
)

func Example_array() {
  a1 := EmptyArray.Set(0, "A").Set(1, "B").Set(2, "C")
  a2 := a1.Del(1).Set(911, "D")
  fmt.Printf("a1: %s\n", a1)
  fmt.Printf("a2: %s\n", a2)
  // Output:
  // a1: [A, B, C]
  // a2: [A, C, D]
}

func TestArrayIteration(t *testing.T) {
  assert := assert.New(t)
  assert.True(1 == 1)

  a1 := EmptyArray.
    Set(1107296256, "A").
    Set(1124073472, "B").
    Set(1129119744, "C").
    Set(1140850688, "D")
  fmt.Printf("a1: %s\n", a1)

  i := 0
  for it := a1.h.Iterator(); it.Next(); {
    fmt.Printf("#%d it.Value %v\n", i, it.Value)
    i++
  }

  os.Exit(0)
}
