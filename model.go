package Sleep

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

// Model struct represents a collection in MongoDB.
// It inherits from mgo.Collection and overwrides some functions.
type Model struct {
	*mgo.Collection
	//C is the underlying mgo.collection value for this model.
	//Refer to http://godoc.org/labix.org/v2/mgo#Collection for full usage information
	C *mgo.Collection
	z *Sleep
}

func newModel(collection *mgo.Collection, z *Sleep) *Model {
	model := &Model{collection, collection, z}
	return model
}

// CreateDoc conditions an instance of the model to become a document.
//
// What it means in pratical terms is that Create sets a value for the schema's *Sleep.Document anonymous field. This will allow Sleep to work with the value as a document.
// Calling this function is only necessary when wishing to create documents "manually".
// It is not necessary to call this function on a value that will be holding the result of a query; Sleep will do that.
//
// After a document is created with this function, the document will expose all of the public methods and fields of the Sleep.Model struct as its own.
func (m *Model) CreateDoc(i interface{}) {
	m.z.CreateDoc(i)
}

// Find starts and returns a chainable *Query value
// This function passes the passed value to mgo.Collection.Find
//
// To borrow from the mgo docs: "The document(argument) may be a map or a struct value capable of being marshalled with bson.
// The map may be a generic one using interface{} for its key and/or values, such as bson.M, or it may be a properly typed map.
// Providing nil as the document is equivalent to providing an empty document such as bson.M{}".
//
// Further reading: http://godoc.org/labix.org/v2/mgo#Collection.Find
func (m *Model) Find(query interface{}) *Query {
	return &Query{query: query, z: m.z,
		populate:  make(map[string]*Query),
		populated: make(map[string]interface{}), c: m.C}
}

// FindId is a convenience function equivalent to:
//
//     query := myModel.Find(bson.M{"_id": id})
//
// Unlike the Mgo.Collection.FindId function, this function will accept Id both in hex representation as a string or a bson.ObjectId.
//
// FindId will return a chainable *Query value
func (m *Model) FindId(id interface{}) *Query {
	return m.Find(bson.M{"_id": getObjectId(id)})
}

func (m *Model) RemoveId(id interface{}) error {
	return m.C.RemoveId(getObjectId(id))
}

func (m *Model) UpdateId(id interface{}, change interface{}) error {
	return m.C.UpdateId(getObjectId(id), change)
}

func (m *Model) UpsertId(id interface{}, change interface{}) (*mgo.ChangeInfo, error) {
	return m.C.UpsertId(getObjectId(id), change)
}
