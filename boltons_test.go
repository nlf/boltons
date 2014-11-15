package boltons

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	ID         string
	TestString string
	TestNumber int
	TestBool   bool
}

type WrappedTestStruct struct {
	ID          string
	TestStructs []TestStruct
}

func TestCreate(t *testing.T) {
	assert := assert.New(t)

	db, err := Open("test.db", 0600, nil)
	defer db.Close()
	assert.NoError(err, "should not error")
}

func TestSave(t *testing.T) {
	assert := assert.New(t)

	db, err := Open("test.db", 0600, nil)
	defer db.Close()
	assert.NoError(err, "should not error")

	err = db.Save("testing")
	assert.Error(err, "should return an error for non-structs")

	s := "testing"
	err = db.Save(&s)
	assert.Error(err, "should return an error for a pointer to a non-struct")

	ts := TestStruct{"test-id", "string", 1, false}
	err = db.Save(ts)
	assert.Error(err, "should return an error for a direct struct")

	err = db.Save(&ts)
	assert.NoError(err, "should not error")

	ts2 := TestStruct{
		TestString: "string",
		TestNumber: 2,
		TestBool:   true,
	}
	err = db.Save(&ts2)
	assert.NoError(err, "should not error")
	assert.NotEqual(ts2.ID, "", "should not be empty")

	ts3 := WrappedTestStruct{
		ID:          "nested",
		TestStructs: []TestStruct{{"test-inner", "inner-string", 3, true}, {"test-inner-2", "inner-string-2", 4, true}},
	}
	err = db.Save(&ts3)
	assert.NoError(err, "should not error")
	assert.NotEqual(ts3.ID, "", "should not be empty")
}

func TestGet(t *testing.T) {
	assert := assert.New(t)

	db, err := Open("test.db", 0600, nil)
	defer db.Close()
	assert.NoError(err, "should not error")

	ts := TestStruct{}
	err = db.Get(&ts)
	assert.Error(err, "cannot fetch without being given an ID")

	ts = TestStruct{
		ID: "test-id",
	}

	err = db.Get(&ts)
	assert.NoError(err, "should not error")
	assert.Equal(ts.ID, "test-id", "should have the ID still set")
	assert.Equal(ts.TestString, "string", "should have the TestString field set")
	assert.Equal(ts.TestNumber, 1, "should have the TestNumber field set")
	assert.Equal(ts.TestBool, false, "should have the TestBool field set")

	wts := WrappedTestStruct{
		ID: "nested",
	}

	err = db.Get(&wts)
	assert.NoError(err, "should not error")
	assert.Equal(wts.ID, "nested", "should have the ID still set")
	assert.NotEqual(len(wts.TestStructs), 0, "should have nested structs")
}

func TestAll(t *testing.T) {
	assert := assert.New(t)

	db, err := Open("test.db", 0600, nil)
	defer db.Close()
	assert.NoError(err, "should not error")

	tsList := []TestStruct{}
	err = db.All(&tsList)
	assert.NoError(err, "should not error")
	assert.NotEqual(len(tsList), 0, "should have members")
}

func TestKeys(t *testing.T) {
	assert := assert.New(t)

	db, err := Open("test.db", 0600, nil)
	defer db.Close()
	assert.NoError(err, "should not error")

	keys, err := db.Keys(TestStruct{})
	assert.NoError(err, "should not error")
	assert.NotEqual(len(keys), 0, "should have keys")

	keys, err = db.Keys(&TestStruct{})
	assert.NoError(err, "should not error")
	assert.NotEqual(len(keys), 0, "should have keys")
}

func TestExists(t *testing.T) {
	assert := assert.New(t)

	db, err := Open("test.db", 0600, nil)
	defer db.Close()
	assert.NoError(err, "should not error")

	ts := TestStruct{
		ID: "test-id",
	}

	exists, err := db.Exists(&ts)
	assert.NoError(err, "should not error")
	assert.Equal(exists, true, "should return true")

	ts = TestStruct{
		ID: "not-here",
	}

	exists, err = db.Exists(&ts)
	assert.NoError(err, "should not error")
	assert.Equal(exists, false, "should return false")
}
