package Sleep

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"reflect"
)

type Document struct {
	C     *mgo.Collection
	Model *Model
	// a pointer to the schema
	schema    interface{}
	isQueried bool
	populated map[string]interface{}
	//an instance of the schema that this document represents
	schemaStruct interface{}
	Found        bool
	Virtual      *Virtual
}

// Save uses MongoDB's upsert command to either update an existing document or insert it into the collection.
// The document's schma MUST have an Id field.
func (d *Document) Save() error {
	idField := reflect.ValueOf(d.schema).Elem().FieldByName("Id")
	reflect.ValueOf(d.schema).MethodByName("PreSave").Call([]reflect.Value{})
	id := idField.Interface()
	_, err := d.C.UpsertId(id, d.schema)

	if err != nil {
		d.Found = true
		reflect.ValueOf(d.schema).MethodByName("PostSave").Call([]reflect.Value{})
	}
	return err
}

// Use this method to check if this document is in fact populated with data from the database.
// Sleep suppresses mgo's ErrNotFound error and instead provides this interface for checking if results were returned.
func (d *Document) IsValid() bool {
	return d.Found
}

//Same as Query.Populate() except it can be called on an existing document.
func (d *Document) Populate(fields ...string) error {
	dummyQuery := d.Model.Find(bson.M{}).Populate(fields...)
	err := dummyQuery.populateExec(d.schema)
	return err
}

//Same as populate but used to populate only a single field. Its last parameter is a pointer
//to the variable to hold the value of the result
func (d *Document) PopulateOne(field string, value interface{}) error {
	dummyQuery := d.Model.Find(bson.M{}).Populate(field)
	err := dummyQuery.populateExec(d.schema)
	if err != nil {
		return err
	}
	populatedField, ok := d.populated[field]
	if ok {
		reflect.ValueOf(value).Elem().Set(reflect.ValueOf(populatedField))
	}
	return nil
}

//Same as Query.PopulateQuery() except the last parameter is a pointer to the variable to hold
//the value of the result.
func (d *Document) PopulateQuery(path string, q *Query, value interface{}) error {
	dummyQuery := d.Model.Find(bson.M{}).PopulateQuery(path, q)
	err := dummyQuery.populateExec(d.schema)

	if err != nil {
		return err
	}
	populatedField, ok := d.populated[path]
	if ok {
		reflect.ValueOf(value).Elem().Set(reflect.ValueOf(populatedField))
	}
	return nil
}

// Populated gives access to the document's populated fields. This method does NOT make a database query.
// It returns only existing populated fields.
//
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
func (d *Document) Populated(path string, result interface{}) bool {
	value, ok := d.populated[path]
	if !ok {
		return ok
	}
	if reflect.ValueOf(result).Type().Kind() != reflect.Ptr {
		panic("Expected a pointer, got a value")
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(value).Elem())
	return ok
}

// Removes the document from the database
func (d *Document) Remove() error {
	reflect.ValueOf(d.schema).MethodByName("PreRemove").Call([]reflect.Value{})
	id := reflect.ValueOf(d.schema).Elem().FieldByName("Id").Interface().(bson.ObjectId)
	err := d.C.Remove(bson.M{"_id": id})
	//if we want it gone and it's already gone, should we really freak out?
	if err == mgo.ErrNotFound {
		err = nil
	}

	if err != nil {
		reflect.ValueOf(d.schema).MethodByName("PostRemove").Call([]reflect.Value{})
	}
	return err
}

//implement Apply function here
// it takes care of applying changes/merging to the document from another document
func (d *Document) Apply(update interface{}) error {
	change := mgo.Change{
		Update:    update,
		Upsert:    true,
		ReturnNew: true}

	id := reflect.ValueOf(d.schema).Elem().FieldByName("Id").Interface().(bson.ObjectId)
	_, err := d.C.FindId(id).Apply(change, d.schema)
	return err
}

// implement populate function here so that  a document is able to be populated
// after the initial query for its value

//implement stand-in hooks here

// PreSave is a stand-in method that can be implemented in the schema defination struct
// to be called before the document is saved to the database.
//
// The method should have a reciever that is a pointer to the schema type
func (d *Document) PreSave() {

}

// PostSave is a stand-in method that can be implemented in the schema defination struct
// to be called after the document is saved to the database.
//
// The method should have a reciever that is a pointer to the schema type
func (d *Document) PostSave() {

}

// PreRemove is a stand-in method that can be implemented in the schema defination struct
// to be called before the document is removed from the database.
//
// The method should have a reciever that is a pointer to the schema type
func (d *Document) PreRemove() {

}

// PostRemove is a stand-in method that can be implemented in the schema defination struct
// to be called after the document is removed from the database.
//
// The method should have a reciever that is a pointer to the schema type
func (d *Document) PostRemove() {

}

// OnCreate is a stand-in method that can be implemented in the schema defination struct
// to be called when the document is created using Sleep.Model.CreateDoc method.
// Use `OnResult()` to be called then the document is queried from the database.
//
// The method should have a reciever that is a pointer to the schema type
func (d *Document) OnCreate() {

}

// OnResult is a stand-in method that can be implemented in the schema defination struct
// to be called when the document is created from the results out of the database.
// Use `OnCreate()` to be called then the document is created using Sleep.Model.CreateDoc method
//
// The method should have a reciever that is a pointer to the schema type
func (d *Document) OnResult() {

}
