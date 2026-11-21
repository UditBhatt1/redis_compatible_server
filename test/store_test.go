package test

import (
	"testing"

	"github.com/UditBhatt1/redis_compatible_server/server"
)

func checkSetGet(t *testing.T) {
	store := server.NewStore()

	store.Set("name", "udit")
	value, exist := store.Get("name")

	if value != "udit" || !exist {
		t.Errorf("wanted value udit and exist true, got %v and %v", value, exist)
	}
}

func checkGetMissing(t *testing.T) {
	store := server.NewStore()
	_, exist := store.Get("name")

	if exist {
		t.Errorf("wanted false got true for get missing")
	}
}

func checkSetOverwrite(t *testing.T) {
	store := server.NewStore()
	newValue := "xyz"

	store.Set("name", "udit")
	store.Set("name", newValue)

	value, _ := store.Get("name")

	if value != newValue {
		t.Errorf("wanted %v got %v", newValue, value)
	}
}

func checkSetDeleteAndGet(t *testing.T) {
	store := server.NewStore()
	store.Set("name", "udit")

	reply := store.Delete("name")

	if !reply {
		t.Fatal("expected delete to return true")
	}

	_, exist := store.Get("name")

	if exist {
		t.Errorf("wanted false got true for get after delete")
	}
}

func TestAll(t *testing.T) {
	t.Run("SetGet", checkSetGet)
	t.Run("GetMissing", checkGetMissing)
	t.Run("SetOverwrite", checkSetOverwrite)
	t.Run("DeleteAndGet", checkSetDeleteAndGet)
}
