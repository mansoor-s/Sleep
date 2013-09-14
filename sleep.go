// Package Sleep provides an intuitive ODM (Object Document Model) library for working
// with MongoDB documents.
// It builds on top of the awesome mgo library
package Sleep

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"reflect"
)

//Convenient access to bson.M
type M bson.M

//Convenient access to bson.D
type D bson.D

type Sleep struct {
	Db       *mgo.Database
	models   map[string]Model
	modelTag string
}

// New returns a new intance of the Sleep type
func New(session *mgo.Session, dbName string) *Sleep {
	sleep := &Sleep{Db: session.DB(dbName), modelTag: "model"}
	sleep.models = make(map[string]Model)
	return sleep
}

// SetModelTag changes the default tag key of `model` to an arbitrary key.
// This value is read to make relationships for populting based on ObjectIds
func (z *Sleep) SetModelTag(key string) {
	z.modelTag = key
}

// Register registers a given schema and its corresponding collection name with Sleep.
// All schemas MUST be registered using this function.
func (z *Sleep) Register(schema interface{}, collectionName string) {
	typ := reflect.TypeOf(schema)
	structName := typ.Name()

	z.models[structName] = Model{C: z.Db.C(collectionName),
		isQueried: true, schema: schema, populated: make(map[string]interface{})}
}

// Find starts a chainable *Query value
// This function passes the supplied value to mgo.Collection.Find
// To borrow from the mgo docs: "The document(argument) may be a map or a struct value capable of being marshalled with bson.
// The map may be a generic one using interface{} for its key and/or values, such as bson.M, or it may be a properly typed map.
// Providing nil as the document is equivalent to providing an empty document such as bson.M{}".
//
// Further reading: http://godoc.org/labix.org/v2/mgo#Collection.Find
func (z *Sleep) Find(query interface{}) *Query {
	return &Query{query: query, z: z,
		populate:  make(map[string]*Query),
		populated: make(map[string]interface{})}
}

// FindId is a convenience function equivalent to:
//
//     query := collection.Find(bson.M{"_id": id})
//
// Unlike the Mgo.Collection.FindId function, this function will accept Id both in Hex representation as a string or a bson.ObjectId
// FindId will return a chainable *Query value
func (z *Sleep) FindId(id interface{}) *Query {
	typName := reflect.TypeOf(id).Name()
	var idActual bson.ObjectId
	if typName == "string" {
		str := id.(string)
		idActual = bson.ObjectIdHex(str)
	} else if typName == "ObjectId" {
		idActual = id.(bson.ObjectId)
	} else {
		panic("Invalid type passed to FindId! Will only accept `bson.ObjectId` or `string`")
	}
	return &Query{query: M{"_id": idActual}, z: z,
		populate:  make(map[string]*Query),
		populated: make(map[string]interface{})}
}

// Create conditions an instance of the model to become a document.
// What it means in pratical terms is that Create sets a value for the schema's Model(Sleep.Model) anonymous field. This will allow Sleep to work with the value as a document.
// Calling this function is only necessary when wishing to create documents "manually".
// It is not necessary to call this function on a value that will be holding the result of a query; Sleep will do that.
// After a document is created with this function, the document will expose all of the public methods and fields of the Sleep.Model struct as its own.
func (z *Sleep) Create(doc interface{}) {
	typ := reflect.TypeOf(doc).Elem()
	structName := typ.Name()
	model := z.models[structName]

	model.doc = doc
	val := reflect.ValueOf(doc).Elem()
	modelVal := val.FieldByName("Model")
	modelVal.Set(reflect.ValueOf(model))

	idField := reflect.ValueOf(doc).Elem().FieldByName("Id")
	id := bson.NewObjectId()
	idField.Set(reflect.ValueOf(id))
}

// C gives access to the underlying *mgo.Collection value for a model.
// The model name is case sensitive.
func (z *Sleep) C(model string) (*mgo.Collection, bool) {
	m, ok := z.models[model]
	c := m.C
	return c, ok
}
