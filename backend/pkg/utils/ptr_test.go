package utils

import (
	"testing"
)

func TestPtr(t *testing.T) {
	t.Parallel()

	t.Run("int", func(t *testing.T) {
		t.Parallel()

		value := 42
		ptr := Ptr(value)

		if ptr == nil {
			t.Fatal("Ptr() returned nil")
		}

		if *ptr != value {
			t.Errorf("Ptr() = %v, want %v", *ptr, value)
		}
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()

		value := "test"
		ptr := Ptr(value)

		if ptr == nil {
			t.Fatal("Ptr() returned nil")
		}

		if *ptr != value {
			t.Errorf("Ptr() = %v, want %v", *ptr, value)
		}
	})

	t.Run("bool", func(t *testing.T) {
		t.Parallel()

		value := true
		ptr := Ptr(value)

		if ptr == nil {
			t.Fatal("Ptr() returned nil")
		}

		if *ptr != value {
			t.Errorf("Ptr() = %v, want %v", *ptr, value)
		}
	})

	t.Run("zero value", func(t *testing.T) {
		t.Parallel()

		value := 0
		ptr := Ptr(value)

		if ptr == nil {
			t.Fatal("Ptr() returned nil for zero value")
		}

		if *ptr != value {
			t.Errorf("Ptr() = %v, want %v", *ptr, value)
		}
	})

	t.Run("struct", func(t *testing.T) {
		t.Parallel()

		type TestStruct struct {
			Name  string
			Value int
		}

		value := TestStruct{Name: "test", Value: 42}
		ptr := Ptr(value)

		if ptr == nil {
			t.Fatal("Ptr() returned nil")
		}

		if ptr.Name != value.Name || ptr.Value != value.Value {
			t.Errorf("Ptr() = %v, want %v", *ptr, value)
		}
	})

	t.Run("pointer independence", func(t *testing.T) {
		t.Parallel()

		value := 42
		ptr1 := Ptr(value)
		ptr2 := Ptr(value)

		if ptr1 == ptr2 {
			t.Error("Ptr() should return different pointers for each call")
		}

		*ptr1 = 100
		if *ptr2 != 42 {
			t.Error("Modifying one pointer should not affect another")
		}
	})

	t.Run("slice", func(t *testing.T) {
		t.Parallel()

		value := []int{1, 2, 3}
		ptr := Ptr(value)

		if ptr == nil {
			t.Fatal("Ptr() returned nil")
		}

		if len(*ptr) != len(value) {
			t.Errorf("Ptr() slice length = %v, want %v", len(*ptr), len(value))
		}
	})

	t.Run("map", func(t *testing.T) {
		t.Parallel()

		value := map[string]int{"a": 1, "b": 2}
		ptr := Ptr(value)

		if ptr == nil {
			t.Fatal("Ptr() returned nil")
		}

		if len(*ptr) != len(value) {
			t.Errorf("Ptr() map length = %v, want %v", len(*ptr), len(value))
		}
	})
}