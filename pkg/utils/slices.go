package utils

import (
	"sync"
)

type ThreadSafeSlice struct {
	sync.Mutex
	Elements []interface{}
}

func (slice *ThreadSafeSlice) Push(elem interface{}) {
	slice.Lock()
	defer slice.Unlock()

	slice.Elements = append(slice.Elements, elem)
}

func (slice *ThreadSafeSlice) Find(elem interface{}) interface{} {
	slice.Lock()
	defer slice.Unlock()

	var foundElem interface{}
	for _, e := range slice.Elements {
		if e == elem {
			foundElem = e
			break
		}
	}
	return foundElem
}

func (slice *ThreadSafeSlice) Remove(elem interface{}) {
	slice.Lock()
	defer slice.Unlock()

	index := -1
	for i, el := range slice.Elements {
		if el == elem {
			index = i
			break
		}
	}
	if index > -1 {
		slice.Elements = append(slice.Elements[:index], slice.Elements[index+1:]...)
	}
}

func (slice *ThreadSafeSlice) ToSlice() []interface{} {
	slice.Lock()
	defer slice.Unlock()
	ret := []interface{}{}
	copy(ret, slice.Elements)
	return ret
}
