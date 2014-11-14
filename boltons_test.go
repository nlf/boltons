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
}

func TestGet(t *testing.T) {
	assert := assert.New(t)

	db, err := Open("test.db", 0600, nil)
	defer db.Close()
	assert.NoError(err, "should not error")

	ts := TestStruct{
		ID: "test-id",
	}

	err = db.Get(&ts)
	assert.NoError(err, "should not error")
	assert.Equal(ts.ID, "test-id", "should have the ID still set")
	assert.Equal(ts.TestString, "string", "should have the TestString field set")
	assert.Equal(ts.TestNumber, 1, "should have the TestNumber field set")
	assert.Equal(ts.TestBool, false, "should have the TestBool field set")
}
