package throttle

import (
	"testing"
	"time"
)

func TestThrottle(t *testing.T) {
	thro := NewThrottle(3, time.Second)
	for i := 0; i < 3; i++ {
		if thro.Doable() == false {
			t.Error("Doable() should return true")
		}
	}
	for i := 0; i < 2; i++ {
		if thro.Doable() == true {
			t.Error("Doable() should return false")
		}
	}

	time.Sleep(2 * time.Second)

	for i := 0; i < 3; i++ {
		if thro.Doable() == false {
			t.Error("Doable() should return true")
		}
	}
	for i := 0; i < 2; i++ {
		if thro.Doable() == true {
			t.Error("Doable() should return false")
		}
	}
}
