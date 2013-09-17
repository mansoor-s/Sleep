package Sleep

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type Model struct {
	*mgo.Collection
	C *mgo.Collection
	z *Sleep
}

func newModel(collection *mgo.Collection, z *Sleep) *Model {
	model := &Model{collection, collection, z}
	return model
}

func (m *Model) CreateDoc(i interface{}) {
	m.z.CreateDoc(i)
}

// Find starts a chainable *Query value
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
// If a hex string representation of the ObjectId is passed, it will get parsed into to a bson.ObjectId value.
//
// FindId will return a chainable *Query value
func (m *Model) FindId(id interface{}) *Query {
	return m.Find(bson.M{"_id": ObjectId(id)})
}

func (m *Model) RemoveId(id interface{}) error {
	return m.C.RemoveId(ObjectId(id))
}

func (m *Model) UpdateId(id interface{}, change interface{}) error {
	return m.C.UpdateId(ObjectId(id), change)
}

func (m *Model) UpsertId(id interface{}, change interface{}) (*mgo.ChangeInfo, error) {
	return m.C.UpsertId(ObjectId(id), change)
}
