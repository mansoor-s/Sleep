package Sleep

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"reflect"
)

type Document struct {
	C         *mgo.Collection
	doc       interface{}
	isQueried bool
	populated map[string]interface{}
	schema    interface{}
	Found     bool
	Virtual   *Virtual
}

// Save uses MongoDB's upsert command to either update an existing document or insert it into the collection.
// The document's schma MUST have an Id field.
func (d *Document) Save() error {
	idField := reflect.ValueOf(d.doc).Elem().FieldByName("Id")
	if !idField.IsValid() {
		panic("Document `" + reflect.TypeOf(d.doc).Elem().Name() + "` must have an `Id` field")
	}

	id := idField.Interface()
	_, err := d.C.UpsertId(id, d.doc)
	return err
}

func (d *Document) IsValid() bool {
	return d.Found
}

// Field gives access to a populated field.
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
func (d *Document) Field(path string, result interface{}) bool {
	value, ok := d.populated[path]
	if !ok {
		return ok
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(value).Elem())
	return ok
}

func (d *Document) Remove() error {
	id := reflect.ValueOf(d.doc).Elem().FieldByName("Id").Interface().(bson.ObjectId)
	return d.C.Remove(bson.M{"_id": id})
}

//implement Apply function here
// it takes care of applying changes/merging to the document from another document

// implement populate function here so that  a document is able to be populated
// after the initial query for its value

//implement stand-in hooks here

// PreSave is a stand-in method that can be implemented in the schema defination struct
// to be called before the document is saved to the database.
//
// The method should have a reciever that is a pointer to the schema type
func (m *Document) PreSave() {

}

// PostSave is a stand-in method that can be implemented in the schema defination struct
// to be called after the document is saved to the database.
//
// The method should have a reciever that is a pointer to the schema type
func (m *Document) PostSave() {

}

// PreRemove is a stand-in method that can be implemented in the schema defination struct
// to be called before the document is removed from the database.
//
// The method should have a reciever that is a pointer to the schema type
func (m *Document) PreRemove() {

}

// PostRemove is a stand-in method that can be implemented in the schema defination struct
// to be called after the document is removed from the database.
//
// The method should have a reciever that is a pointer to the schema type
func (m *Document) PostRemove() {

}

// OnCreate is a stand-in method that can be implemented in the schema defination struct
// to be called when the document is created using Sleep.Modle.CreateDoc method.
//
// The method should have a reciever that is a pointer to the schema type
func (m *Document) OnCreate() {

}
