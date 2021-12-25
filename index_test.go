package querysan

import (
	"fmt"
	"os"
	"testing"
)

func TestAddIndex(t *testing.T) {
	dbPath := "/tmp/test.db"
	if err := os.Remove(dbPath); err != nil {
		t.Fail()
	}
	if err := InitializeDatabase(dbPath); err != nil {
		t.Fail()
	}
	if err := AddIndex("data/1.txt"); err != nil {
		t.Fail()
	}
	if err := AddIndex("data/2.txt"); err != nil {
		t.Fail()
	}
	paths, err := Query("適*")
	if err != nil {
		t.Fail()
	}
	fmt.Println(paths)
	CloseDb()
}
