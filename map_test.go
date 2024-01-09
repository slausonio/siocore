package siocore

import (
	"reflect"
	"testing"
)

func TestMergeMaps(t *testing.T) {
	tests := []struct {
		name   string
		maps   []map[string]int
		expect map[string]int
	}{
		{
			name:   "Nil",
			maps:   nil,
			expect: make(map[string]int),
		},
		{
			name:   "Empty",
			maps:   []map[string]int{},
			expect: make(map[string]int),
		},
		{
			name:   "Single",
			maps:   []map[string]int{{"key": 1}},
			expect: map[string]int{"key": 1},
		},
		{
			name:   "Multiple",
			maps:   []map[string]int{{"key1": 1}, {"key2": 2}},
			expect: map[string]int{"key1": 1, "key2": 2},
		},
		{
			name:   "Overlap",
			maps:   []map[string]int{{"key": 1}, {"key": 2}},
			expect: map[string]int{"key": 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeMaps(tt.maps...); !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("MergeMaps() = %v, want %v", got, tt.expect)
			}
		})
	}
}
