package main

import . "github.com/amitu/gutils"

func conv(v interface{}, err error) (*Message, error) {
	if v == nil {
		return nil, err
	}
	return v.(*Message), err
}

type CircularMessageArray struct {
	CircularArray
}

func NewCircularMessageArray(size uint) *CircularMessageArray {
	return &CircularMessageArray{CircularArray{Size: size}}
}

func (circ *CircularMessageArray) Push(buf *Message) {
	circ.CircularArray.Push(buf)
}

func (circ *CircularMessageArray) Pop() (*Message, error) {
	return conv(circ.CircularArray.Pop())
}

func (circ *CircularMessageArray) PopNewest() (*Message, error) {
	return conv(circ.CircularArray.PopNewest())
}

func (circ *CircularMessageArray) PeekOldest() (*Message, error) {
	return conv(circ.CircularArray.PeekOldest())
}

func (circ *CircularMessageArray) PeekNewest() (*Message, error) {
	return conv(circ.CircularArray.PeekNewest())
}

func (circ *CircularMessageArray) Ith(i uint) (*Message, error) {
	return conv(circ.CircularArray.Ith(i))
}
