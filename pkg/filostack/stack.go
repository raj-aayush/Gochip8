package filostack

import "fmt"

//FiloStack - Define filoStack struct
type FiloStack struct {
	stack []uint16
	limit int
}

//Push elem to top of stack
func (st *FiloStack) Push(elem uint16) error {
	if len(st.stack) < st.limit {
		st.stack = append(st.stack, elem)
		return nil
	}
	return fmt.Errorf("stack error: limit reached")
}

//Pop elem from top of stack
func (st *FiloStack) Pop() (uint16, error) {
	if len(st.stack) > 0 {
		elem := st.stack[len(st.stack)-1]
		st.stack = st.stack[:len(st.stack)-1]
		return elem, nil
	}
	return uint16(0), fmt.Errorf("stack error: nothing to pop")
}
