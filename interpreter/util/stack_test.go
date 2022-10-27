package util

import "testing"

func TestForEach(t *testing.T) {
	s := NewStack[int]()
	s.Push(1)
	s.Push(2)
	s.Push(3)

	expectedIdx := 2
	expectedValue := 3

	s.ForEach(func(i, val int) bool {
		if i != expectedIdx {
			t.Errorf("expected idx %d, got %d", expectedIdx, i)
		}

		if val != expectedValue {
			t.Errorf("expected val %d, got %d", expectedValue, val)
		}

		expectedIdx = expectedIdx - 1
		expectedValue = expectedValue - 1

		return true
	})

	calls := 0
	s.ForEach(func(i, val int) bool {
		calls = calls + 1
		return false
	})

	if calls != 1 {
		t.Errorf("expected to short circuit foreach loop")
	}
}
