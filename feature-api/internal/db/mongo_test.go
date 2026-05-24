package db

import "testing"

func TestDisconnect_Nil(t *testing.T) {
	Disconnect(nil)
}
