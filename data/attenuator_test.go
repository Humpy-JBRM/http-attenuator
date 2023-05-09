package data

import (
	"testing"
	"time"
)

type simpleAttenuated struct {
	fired int
}

func (s *simpleAttenuated) Fire(params ...interface{}) error {
	s.fired++
	return nil
}

func TestAttenuate(t *testing.T) {
	sa := &simpleAttenuated{}
	a, err := NewAttenuator("test1", 2, false, false, sa)
	if err != nil {
		t.Fatal(err)
	}

	a.Start()
	select {
	case <-time.After(2300 * time.Millisecond):
		break
	}

	if sa.fired != 4 {
		t.Errorf("Should only have fired 4 times, but fired %d times", sa.fired)
	}
}
