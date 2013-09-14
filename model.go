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

// Save uses MongoDB's upsert command to either update an existing document or insert it into the collection.
// The document's schma MUST have an Id field.
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

// Get gives access to a populated field.
//
// The path must be exactly the same as what was passed to Query.Populate() or Query.PopulateQuery() and is case sensitive.
//
// The result parameter must be of the correct Type.
// For example, if the field was defined as such in the schema:
//
//		Foo: bson.ObjectId   `model:"Bar"`
//
// Then the argument must be of type   *Bar
// Or, if the field was defined as:
//
//		Foo: []bson.ObjectId   `model:"Bar"`
//
// Then the argument must be of type:   *[]*Bar
//
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
