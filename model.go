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
	schema    interface{}
}

func (m *Model) Save() error {
	idField := reflect.ValueOf(m.doc).Elem().FieldByName("Id")
	if !idField.IsValid() {
		panic("Model `" + reflect.TypeOf(m.doc).Elem().Name() + "` must have an `Id` field")
	}

	id := idField.Interface()
	_, err := m.C.UpsertId(id, m.doc)
	return err
}

func (m *Model) Remove() error {
	return nil
}

//
func (m *Model) Get(path string, result interface{}) bool {
	value, ok := m.populated[path]
	if !ok {
		return ok
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(value).Elem())
	return ok
}

//implement stand-in hooks here

func (m *Model) PreSave() {

}

func (m *Model) PostSave() {

}

func (m *Model) PreRemove() {

}

func (m *Model) PostRemove() {

}

func (m *Model) Create() {

}

func (m *Model) PreUpdate() {

}

func (m *Model) PostUpdate() {

}
