// Copyright (c) 2015, Emir Pasic. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package arraylist implements the array list.
//
// Structure is not thread safe.
//
// Reference: https://en.wikipedia.org/wiki/List_%28abstract_data_type%29
package sender

import (
	"fmt"
	eventhub "github.com/Azure/azure-event-hubs-go/v3"
	"strings"
)

// List holds the elements in a slice
type List struct {
	elements [][]*eventhub.Event
	size     int
}

const (
	growthFactor = float32(2.0)  // growth by 100%
	shrinkFactor = float32(0.25) // shrink when size is 25% of capacity (0 means never shrink)
)

// New instantiates a new list and adds the passed values, if any, to the list
func New(values ...[]*eventhub.Event) *List {
	list := &List{}
	if len(values) > 0 {
		list.Add(values...)
	}
	return list
}

// Add appends a value at the end of the list
func (list *List) Add(values ...[]*eventhub.Event) {
	list.growBy(len(values))
	for _, value := range values {
		list.elements[list.size] = value
		list.size++
	}
}

// Get returns the element at index.
// Second return parameter is true if index is within bounds of the array and array is not empty, otherwise false.
func (list *List) Get(index int) ([]*eventhub.Event, bool) {

	if !list.withinRange(index) {
		return nil, false
	}

	return list.elements[index], true
}

// Remove removes the element at the given index from the list.
func (list *List) Remove(index int) {

	if !list.withinRange(index) {
		return
	}

	list.elements[index] = nil                                    // cleanup reference
	copy(list.elements[index:], list.elements[index+1:list.size]) // shift to the left by one (slow operation, need ways to optimize this)
	list.size--

	list.shrink()
}

// Values returns all elements in the list.
//func (list *List) Values() []*eventhub.Event {
//	newElements := make([]*eventhub.Event, list.size, list.size)
//	copy(newElements, list.elements[:list.size])
//	return newElements
//}

//IndexOf returns index of provided element
//func (list *List) IndexOf(value *eventhub.Event) int {
//	if list.size == 0 {
//		return -1
//	}
//	for index, element := range list.elements {
//		if element == value {
//			return index
//		}
//	}
//	return -1
//}

// Empty returns true if list does not contain any elements.
func (list *List) Empty() bool {
	return list.size == 0
}

// Size returns number of elements within the list.
func (list *List) Size() int {
	return list.size
}

// Clear removes all elements from the list.
func (list *List) Clear() {
	list.size = 0
	list.elements = [][]*eventhub.Event{}
}

// Swap swaps the two values at the specified positions.
func (list *List) Swap(i, j int) {
	if list.withinRange(i) && list.withinRange(j) {
		list.elements[i], list.elements[j] = list.elements[j], list.elements[i]
	}
}

// Insert inserts values at specified index position shifting the value at that position (if any) and any subsequent elements to the right.
// Does not do anything if position is negative or bigger than list's size
// Note: position equal to list's size is valid, i.e. append.
func (list *List) Insert(index int, values ...[]*eventhub.Event) {

	if !list.withinRange(index) {
		// Append
		if index == list.size {
			list.Add(values...)
		}
		return
	}

	l := len(values)
	list.growBy(l)
	list.size += l
	copy(list.elements[index+l:], list.elements[index:list.size-l])
	copy(list.elements[index:], values)
}

// Set the value at specified index
// Does not do anything if position is negative or bigger than list's size
// Note: position equal to list's size is valid, i.e. append.
func (list *List) Set(index int, value []*eventhub.Event) {

	if !list.withinRange(index) {
		// Append
		if index == list.size {
			list.Add(value)
		}
		return
	}

	list.elements[index] = value
}

// String returns a string representation of container
func (list *List) String() string {
	str := "ArrayList\n"
	values := []string{}
	for _, value := range list.elements[:list.size] {
		values = append(values, fmt.Sprintf("%v", value))
	}
	str += strings.Join(values, ", ")
	return str
}

// Check that the index is within bounds of the list
func (list *List) withinRange(index int) bool {
	return index >= 0 && index < list.size
}

func (list *List) resize(cap int) {
	newElements := make([][]*eventhub.Event, cap)
	copy(newElements, list.elements)
	list.elements = newElements
}

// Expand the array if necessary, i.e. capacity will be reached if we add n elements
func (list *List) growBy(n int) {
	// When capacity is reached, grow by a factor of growthFactor and add number of elements
	currentCapacity := cap(list.elements)
	if list.size+n >= currentCapacity {
		newCapacity := int(growthFactor * float32(currentCapacity+n))
		list.resize(newCapacity)
	}
}

// Shrink the array if necessary, i.e. when size is shrinkFactor percent of current capacity
func (list *List) shrink() {
	if shrinkFactor == 0.0 {
		return
	}
	// Shrink when size is at shrinkFactor * capacity
	currentCapacity := cap(list.elements)
	if list.size <= int(float32(currentCapacity)*shrinkFactor) {
		list.resize(list.size)
	}
}
