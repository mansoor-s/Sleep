// Package Sleep provides an intuitive library for working with MongoDB
// documents (specially in a website environment).
// It builds on top of the awesome mgo library
package Sleep

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"reflect"
)

//convenience
type M bson.M
type D bson.D

type Sleep struct {
	Db     		*mgo.Database
	models 		map[string]Model
	modelTag 	string
}

func New(session *mgo.Session, dbName string) *Sleep {
	sleep := &Sleep{Db: session.DB(dbName)}
	sleep.models = make(map[string]Model)
	return sleep
}

func (z *Sleep) SetModelTag(key string) {
	z.modelTag = key
}


func (z *Sleep) Register(schema interface{}, collectionName string) {
	typ := reflect.TypeOf(schema)
	structName := typ.Name()

	z.models[structName] = Model{C: z.Db.C(collectionName), 
		isQueried: true, schema: schema}
}

func (z *Sleep) Find(query interface{}) *Query {
	return &Query{query: query, z: z}
}

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
	return &Query{query: M{"_id": idActual}, z: z}
}

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


func (z *Sleep) C(model string) (*mgo.Collection, bool) {
	model, ok := z.models[model]
	c := model.C
	return c, ok
}