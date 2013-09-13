package Sleep

import (
	"labix.org/v2/mgo"
	"reflect"
)

type Model struct {
	C         *mgo.Collection
	doc       interface{}
	isQueried bool
	populated map[string]interface{}
}

func (m *Model) Save() error {
	idField := reflect.ValueOf(m.doc).Elem().FieldByName("Id")
	id := idField.Interface()
	_, err := m.C.UpsertId(id, m.doc)
	return err
}

//implement stand-in hooks here
