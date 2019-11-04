package yf

import "testing"

func TestSever(t *testing.T) {
	s := NewServer()
	s.AddRouter(func() (prefix string, vs []RouterInfo) {
		return "", []RouterInfo{}
	})

	s.Start(&Config{})
}
