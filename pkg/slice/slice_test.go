package slice

import "testing"

func TestIntSliceContainsInt(t *testing.T) {
	tests := []struct{
		name  string
		value int
		slice []int
		want  bool
	} {
		{
			name: "search value at first position",
			value: 1,
			slice: []int{1, 2, 3, 4, 5},
			want: true,
		},
		{
			name: "search value at middle position",
			value: 4,
			slice: []int{1, 2, 3, 4, 5},
			want: true,
		},
		{
			name: "search value at last position",
			value: 5,
			slice: []int{1, 2, 3, 4, 5},
			want: true,
		},
		{
			name: "search value not present",
			value: 9,
			slice: []int{1, 2, 3, 4, 5},
			want: false,
		},
		{
			name: "nil slice",
			value: 1,
			slice: nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IntSliceContainsIntValue(tt.slice, tt.value)
			if got != tt.want {
				t.Fatalf("wanted %v, got %v", tt.want, got)
			}
		})
	}
}

func TestIsUniqueIntSlice(t *testing.T) {
	tests := []struct{
		name  string
		slice []int
		want  bool
	} {
		{
			name: "unique",
			slice: []int{1, 2, 3, 4, 5},
			want: true,
		},
		{
			name: "one repeating value",
			slice: []int{1, 2, 3, 3, 4, 5},
			want: false,
		},
		{
			name: "all repeating values",
			slice: []int{1, 1, 1, 1, 1},
			want: false,
		},
		{
			name: "empty slice",
			slice: []int{},
			want: true,
		},
		{
			name: "nil slice",
			slice: nil,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsUniqueIntSlice(tt.slice)
			if got != tt.want {
				t.Fatalf("wanted %v, got %v", tt.want, got)
			}
		})
	}
}
