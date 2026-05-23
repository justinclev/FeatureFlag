package cache

import "testing"

func TestClose_Nil(t *testing.T) {
	Close(nil)
}
