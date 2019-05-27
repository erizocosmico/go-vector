package vector

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Vector implements a persistent bit-partitioned vector trie, an array-like
// persistent data structure.
type Vector struct {
	count uint64
	shift uint
	root  *node
	tail  *node
	start int
}

// New returns a new vector containing the given elements.
func New(elems ...interface{}) *Vector {
	v := emptyVector
	for _, e := range elems {
		v = v.Append(e)
	}
	return v
}

// Append returns a new vector appending the element at the end of the vector.
func (v *Vector) Append(elem interface{}) *Vector {
	if v.count-v.tailOffset() < uint64(vectorWidth) {
		lenTail := len(v.tail.values)
		tail := v.tail.cloneWithLen(lenTail + 1)
		tail.values[lenTail] = elem
		return &Vector{v.count + 1, v.shift, v.root, tail, 0}
	}

	var root *node
	tail := v.tail
	shift := v.shift
	if (v.count >> vectorBits) > (1 << v.shift) {
		root = &node{make([]interface{}, vectorWidth)}
		root.values[0] = v.root
		root.values[1] = newPath(tail)
		shift += uint(vectorBits)
	} else {
		root = v.pushTail(shift, v.root, tail)
	}

	tail = &node{[]interface{}{elem}}
	return &Vector{v.count + 1, shift, root, tail, 0}
}

// Get returns the element at the given position. If the position is negative, returns
// elements in reverse order. If the element cannot be found in the vector, it
// will return nil.
func (v *Vector) Get(i int) interface{} {
	var key = uint64(i + v.start)
	if i < 0 {
		key = v.count + uint64(i)
	}

	if key >= v.count {
		return nil
	}

	tailOffset := v.tailOffset()
	if tailOffset == 0 || tailOffset-1 < key {
		return v.tail.values[key-tailOffset]
	}

	n := v.root
	for lvl := v.shift; lvl > 0; lvl -= uint(vectorBits) {
		n = n.values[(key>>lvl)&uint64(vectorMask)].(*node)
	}

	return n.values[key&uint64(vectorMask)]
}

// Set will change the value of the element at the given index. If the element
// does not exist it will panic.
func (v *Vector) Set(i int, elem interface{}) *Vector {
	var key = uint64(i + v.start)
	if i < 0 {
		key = v.count + uint64(i)
	}

	if key >= v.count {
		panic(fmt.Errorf("vector: index out of bounds, tried to get "+
			"element %d of a vector with %d elements", key, v.count))
	}

	tailOffset := v.tailOffset()
	if tailOffset == 0 || tailOffset-1 < key {
		newTail := v.tail.clone()
		newTail.values[key-tailOffset] = elem
		return &Vector{v.count, v.shift, v.root, newTail, v.start}
	}

	root := v.root.clone()
	n := root
	for lvl := v.shift; lvl > 0; lvl -= uint(vectorBits) {
		idx := (key >> lvl) & uint64(vectorMask)
		newNode := n.values[idx].(*node).clone()
		n.values[idx] = newNode
		n = newNode
	}

	n.values[key&uint64(vectorMask)] = elem
	return &Vector{v.count, v.shift, root, v.tail, v.start}
}

// ErrStop may be returned to stop iterating a vector.
var ErrStop = errors.New("stop")

// Range iterates over the vector to access all its elements. In order to stop
// the iteration, ErrStop may be returned. Any other error will also terminate
// the iteration and will also return that error.
func (v *Vector) Range(f func(a interface{}) error) error {
	for i := 0; i < int(v.count); i++ {
		if err := f(v.Get(i)); err != nil {
			if err == ErrStop {
				return nil
			}
			return err
		}
	}
	return nil
}

// First returns the first element of the vector.
func (v *Vector) First() interface{} {
	return v.Get(0)
}

// Last returns the last element of the vector.
func (v *Vector) Last() interface{} {
	return v.Get(-1)
}

// Tail returns all the elements in the vector except for the first one.
func (v *Vector) Tail() *Vector {
	return v.Drop(1)
}

// Count returns the number of elements in the vector.
func (v *Vector) Count() int {
	return int(v.count) - int(v.start)
}

// pushTail pushes the tail to the rightmost node available and returns a new root.
func (v *Vector) pushTail(shift uint, root, tail *node) *node {
	newRoot := root.clone()
	newNode := tail
	idx := ((v.count - 1) >> shift) & uint64(vectorWidth-1)
	if shift > uint(vectorBits) {
		shift -= uint(vectorBits)
		if n, ok := root.values[idx].(*node); ok {
			newNode = v.pushTail(shift, n, tail)
		} else {
			newNode = newPath(tail)
		}
	}

	newRoot.values[idx] = newNode
	return newRoot
}

// tailOffset returns the offset of elements that are not on the tail.
func (v *Vector) tailOffset() uint64 {
	if v.count < uint64(vectorWidth) {
		return 0
	}
	return ((v.count - 1) >> 5) << 5
}

// Slice returns the elements of the vector in a slice.
func (v *Vector) Slice() []interface{} {
	var result = make([]interface{}, int(v.count))
	for i := 0; i < int(v.count); i++ {
		result[i] = v.Get(i)
	}
	return result
}

// Map returns a new vector with the elements of the current vector after
// applying the given map function.
func (v *Vector) Map(f func(interface{}) interface{}) *Vector {
	result := New()
	for i := 0; i < int(v.count); i++ {
		result = result.Append(f(v.Get(i)))
	}
	return result
}

// Filter returns a new vector with the elements of the current vector if they
// satisfy the given filter function.
func (v *Vector) Filter(f func(interface{}) bool) *Vector {
	result := New()
	for i := 0; i < int(v.count); i++ {
		elem := v.Get(i)
		if f(elem) {
			result = result.Append(elem)
		}
	}
	return result
}

// Take returns a new vector with the first n elements of this vector.
func (v *Vector) Take(n int) *Vector {
	if uint64(n) >= v.count {
		return v
	}

	result := New()
	for i := 0; i < n; i++ {
		result = result.Append(v.Get(i))
	}
	return result
}

// Drop returns a new vector with all the elements in this vector dropping the
// first n elements.
func (v *Vector) Drop(n int) *Vector {
	if uint64(v.start+n) >= v.count {
		return New()
	}

	return &Vector{
		v.count,
		v.shift,
		v.root,
		v.tail,
		v.start + n,
	}
}

// String returns a string representation of the persistent vector.
func (v *Vector) String() string {
	var items []string
	for i := uint64(0); i < v.count-uint64(v.start); i++ {
		items = append(items, fmt.Sprint(v.Get(int(i))))
	}
	return fmt.Sprintf("[%s]", strings.Join(items, ", "))
}

// Equal returns whether a vector has the same items as another vector.
// The comparison between elements is done using reflect.DeepEqual.
func Equal(v1, v2 *Vector) bool {
	return EqualFunc(v1, v2, reflect.DeepEqual)
}

// EqualFn is a function used to tell whether two elements in a vector are
// the same.
type EqualFn func(a, b interface{}) bool

// EqualFunc returns whether a vector has the same items as another vector
// using the given function to determine whether they're equal or not.
func EqualFunc(v1, v2 *Vector, fn EqualFn) bool {
	len1 := v1.Count()
	len2 := v2.Count()

	if len1 != len2 {
		return false
	}

	for i := 0; i < len1; i++ {
		a := v1.Get(i)
		b := v2.Get(i)
		if !fn(a, b) {
			return false
		}
	}

	return true
}

const (
	vectorBits  uint32 = 5
	vectorWidth uint32 = 1 << 5
	vectorMask  uint32 = (1 << 5) - 1
)

type node struct {
	values []interface{}
}

func (n *node) clone() *node {
	return n.cloneWithLen(len(n.values))
}

func (n *node) cloneWithLen(length int) *node {
	newNode := &node{make([]interface{}, length)}
	copy(newNode.values, n.values)
	return newNode
}

var (
	emptyNode   = &node{make([]interface{}, vectorWidth)}
	emptyVector = &Vector{0, 5, emptyNode, &node{nil}, 0}
)

// newPath creates a new path all the way through a branch inserting at the leftmost leaf.
func newPath(n *node) *node {
	node := &node{make([]interface{}, vectorWidth)}
	node.values[0] = n
	return node
}
