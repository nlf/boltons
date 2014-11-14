package boltons

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"

	"code.google.com/p/go-uuid/uuid"

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

type parsedBucket struct {
	name   []byte
	values map[string]reflect.Value
}

func bucketName(s interface{}) ([]byte, error) {
	sType := reflect.TypeOf(s)
	if sType.Kind() != reflect.Ptr {
		return []byte{}, errors.New("Must be a pointer to a struct")
	}

	sValue := reflect.Indirect(reflect.ValueOf(s))
	if sValue.Kind() != reflect.Struct {
		return []byte{}, errors.New("Must be a pointer to a struct")
	}

	sType = sType.Elem()
	return []byte(sType.Name()), nil
}

func parseInput(s interface{}) (parsedBucket, error) {
	bucket := parsedBucket{}

	sType := reflect.TypeOf(s)
	if sType.Kind() != reflect.Ptr {
		return bucket, errors.New("Must be a pointer to a struct")
	}

	sValue := reflect.Indirect(reflect.ValueOf(s))
	if sValue.Kind() != reflect.Struct {
		return bucket, errors.New("Must be a pointer to a struct")
	}

	sType = sType.Elem()
	bucket.name = []byte(sType.Name())
	bucket.values = make(map[string]reflect.Value)

	for i := 0; i < sValue.NumField(); i++ {
		fValue := sValue.Field(i)
		fType := sType.Field(i)

		bucket.values[fType.Name] = fValue
	}

	return bucket, nil
}

func (db *DB) Save(s interface{}) error {
	bucket, err := parseInput(s)
	if err != nil {
		return err
	}

	err = db.bolt.Update(func(tx *bolt.Tx) error {
		outer, err := tx.CreateBucketIfNotExists(bucket.name)
		if err != nil {
			return err
		}

		id := bucket.values["ID"]
		if id.String() == "" {
			id.SetString(uuid.New())
		}

		inner, err := outer.CreateBucketIfNotExists([]byte(id.String()))
		if err != nil {
			return err
		}

		for key, value := range bucket.values {
			bVal, err := json.Marshal(value.Interface())
			if err != nil {
				return nil
			}

			err = inner.Put([]byte(key), bVal)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func (db *DB) Get(s interface{}) error {
	bucket, err := parseInput(s)
	if err != nil {
		return err
	}

	err = db.bolt.View(func(tx *bolt.Tx) error {
		outer := tx.Bucket(bucket.name)

		id := bucket.values["ID"]
		if id.String() == "" {
			return errors.New("Unable to fetch without an ID")
		}

		inner := outer.Bucket([]byte(id.String()))

		for key, value := range bucket.values {
			bVal := inner.Get([]byte(key))

			out := reflect.New(value.Type()).Interface()
			err := json.Unmarshal(bVal, &out)
			if err != nil {
				return err
			}

			value.Set(reflect.Indirect(reflect.ValueOf(out)))
		}

		return nil
	})

	return err
}

func (db *DB) First(s interface{}) error {
	bucket, err := parseInput(s)
	if err != nil {
		return err
	}

	err = db.bolt.View(func(tx *bolt.Tx) error {
		outer := tx.Bucket(bucket.name)
		cursor := outer.Cursor()

		key, _ := cursor.First()
		inner := outer.Bucket(key)

		for key, value := range bucket.values {
			bVal := inner.Get([]byte(key))

			out := reflect.New(value.Type()).Interface()
			err := json.Unmarshal(bVal, &out)
			if err != nil {
				return err
			}

			value.Set(reflect.Indirect(reflect.ValueOf(out)))
		}

		return nil
	})

	return err
}

func (db *DB) Close() {
	db.bolt.Close()
}
