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
	Db        *mgo.Database
	documents map[string]Document
	models    map[string]*Model
	modelTag  string
}

// New returns a new intance of the Sleep type
func New(session *mgo.Session, dbName string) *Sleep {
	sleep := &Sleep{Db: session.DB(dbName), modelTag: "model"}
	sleep.documents = make(map[string]Document)
	sleep.models = make(map[string]*Model)
	return sleep
}

// SetModelTag changes the default tag key of `model` to an arbitrary key.
// This value is read to make relationships for populting based on ObjectIds
func (z *Sleep) SetModelTag(key string) {
	z.modelTag = key
}

// Register registers a given schema and its corresponding collection name with Sleep.
// All schemas MUST be registered using this function.
// Function will return a pointer to the Sleep.Model value for this model
func (z *Sleep) Register(schema interface{}, collectionName string) *Model {
	typ := reflect.TypeOf(schema)
	structName := typ.Name()

	z.documents[structName] = Document{C: z.Db.C(collectionName),
		isQueried: true, schema: schema,
		populated: make(map[string]interface{}), Found: true}

	model := newModel(z.Db.C(collectionName), z)
	z.models[structName] = model
	return model
}

// CreateDoc conditions an instance of the model to become a document.
//
// What it means in pratical terms is that Create sets a value for the schema's *Sleep.Document anonymous field. This will allow Sleep to work with the value as a document.
// Calling this function is only necessary when wishing to create documents "manually".
// It is not necessary to call this function on a value that will be holding the result of a query; Sleep will do that.
//
// After a document is created with this function, the document will expose all of the public methods and fields of the Sleep.Model struct as its own.
func (z *Sleep) CreateDoc(doc interface{}) {
	typ := reflect.TypeOf(doc).Elem()
	structName := typ.Name()
	document := z.documents[structName]

	document.doc = doc
	val := reflect.ValueOf(doc).Elem()
	docVal := val.FieldByName("Document")
	docVal.Set(reflect.ValueOf(document))

	idField := reflect.ValueOf(doc).Elem().FieldByName("Id")
	id := bson.NewObjectId()
	idField.Set(reflect.ValueOf(id))
}

// C gives access to the underlying *mgo.Collection value for a model.
// The model name is case sensitive.
func (z *Sleep) C(model string) (*mgo.Collection, bool) {
	m, ok := z.documents[model]
	c := m.C
	return c, ok
}

func (z *Sleep) Model(name string) *Model {
	return z.models[name]
}

func (z *Sleep) ObjectId(id interface{}) *bson.ObjectId {
	return ObjectId(id)
}

func ObjectId(id interface{}) bson.ObjectId {
	var idActual bson.ObjectId
	switch id.(type) {
	case string:
		idActual = bson.ObjectIdHex(id.(string))
		break
	case bson.ObjectId:
		idActual = id.(bson.ObjectId)
	default:
		panic("Only accepts types `string` and `bson.ObjectId` accepted as Id")
	}
	return idActual
}
