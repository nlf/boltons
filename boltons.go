package boltons

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"

	"github.com/boltdb/bolt"
)

type DB struct {
	bolt *bolt.DB
}

func Open(path string, mode os.FileMode, options *bolt.Options) (*DB, error) {
	db, err := bolt.Open(path, mode, options)
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db *DB) Save(s interface{}) error {
	sType := reflect.TypeOf(s)
	if sType.Kind() != reflect.Ptr {
		return errors.New("Must be a pointer to a struct")
	}

	sValue := reflect.Indirect(reflect.ValueOf(s))
	if sValue.Kind() != reflect.Struct {
		return errors.New("Must be a pointer to a struct")
	}

	sType = sType.Elem()
	err := db.bolt.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(sType.Name()))
		if err != nil {
			return err
		}

		id := sValue.FieldByName("ID")
		if id.String() == "" {
			id.SetString("TEST_STRING")
		}

		inner, err := bucket.CreateBucketIfNotExists([]byte(id.String()))
		if err != nil {
			return err
		}

		for i := 0; i < sValue.NumField(); i++ {
			fValue := sValue.Field(i)
			fType := sType.Field(i)

			bVal, err := json.Marshal(fValue.Interface())
			if err != nil {
				return nil
			}

			inner.Put([]byte(fType.Name), bVal)
		}

		return nil
	})

	return err
}

func (db *DB) Get(s interface{}) error {
	sType := reflect.TypeOf(s)
	if sType.Kind() != reflect.Ptr {
		return errors.New("Must be a pointer to a struct")
	}

	sValue := reflect.Indirect(reflect.ValueOf(s))
	if sValue.Kind() != reflect.Struct {
		return errors.New("Must be a pointer to a struct")
	}

	sType = sType.Elem()
	err := db.bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(sType.Name()))

		id := sValue.FieldByName("ID")
		if id.String() == "" {
			// grab the first item
		}

		inner := bucket.Bucket([]byte(id.String()))

		for i := 0; i < sValue.NumField(); i++ {
			fValue := sValue.Field(i)
			fType := sType.Field(i)

			bVal := inner.Get([]byte(fType.Name))

			out := reflect.New(fType.Type).Interface()
			err := json.Unmarshal(bVal, &out)
			if err != nil {
				return err
			}

			fValue.Set(reflect.Indirect(reflect.ValueOf(out)))
		}

		return nil
	})

	return err
}

func (db *DB) Close() {
	db.bolt.Close()
}
