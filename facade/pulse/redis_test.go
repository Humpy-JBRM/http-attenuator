package facade

import "testing"

func TestRedis(t *testing.T) {
	r, err := NewRedisPulse("foofoo", 1, 1, 0, "localhost:6379")
	if err != nil {
		t.Fatal(err)
	}
	if r == nil {
		t.Fatal("No redis pulse")
	}
	select {}
}
