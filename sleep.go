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
	Db     *mgo.Database
	models map[string]Model
}

func New(session *mgo.Session, dbName string) *Sleep {
	sleep := &Sleep{Db: session.DB(dbName)}
	sleep.models = make(map[string]Model)
	return sleep
}

func (z *Sleep) Register(schema interface{}, collectionName string) {
	typ := reflect.TypeOf(schema)
	structName := typ.Name()

	z.models[structName] = Model{C: z.Db.C(collectionName)}
}

func (z *Sleep) Find(query interface{}) *Query {
	return &Query{query: query, z: z}
}

func (z *Sleep) Create(query interface{}) *Query {
	return &Query{query: query}
}

type Query struct {
	query     interface{}
	selection interface{}
	skip      int
	limit     int
	sort      []string
	populate  []Query
	path      string
	z         *Sleep
}

type O struct {
	Upsert bool
	Multi  bool
}

func (q *Query) Select(selection interface{}) *Query {
	q.selection = selection
	return q
}

func (q *Query) Skip(skip int) *Query {
	q.skip = skip
	return q
}

func (q *Query) Limit(lim int) *Query {
	q.limit = lim
	return q
}

func (q *Query) Sort(fields ...string) *Query {
	q.sort = fields
	return q
}

func (query *Query) Exec(result interface{}) error {
	typ := reflect.TypeOf(result)
	var structName string
	isSlice := false

	if typ.Kind() == reflect.Slice {
		structName = typ.Elem().Name()
		isSlice = true
	} else {
		structName = typ.Elem().Name()
	}

	model := query.z.models[structName]

	q := model.C.Find(query.query)

	if query.limit != 0 {
		q = q.Limit(query.limit)
	}

	if query.skip != 0 {
		q = q.Skip(query.skip)
	}

	sortLen := len(query.sort)
	if sortLen != 0 {
		for i := 0; i < sortLen; i++ {
			q = q.Sort(query.sort[i])
		}
	}

	if query.selection != nil {
		q = q.Select(query.selection)
	}

	var err error

	if isSlice == true {
		err = q.All(result)
		if err != nil {
			return err
		}

		val := reflect.ValueOf(result)
		elemCount := val.Len()
		for i := 0; i < elemCount; i++ {
			modelCpy := query.z.models[structName]
			sliceElem := val.Index(i)
			modelCpy.doc = sliceElem.Interface()
			modelElem := sliceElem.Elem().FieldByName("Sleep.Model")
			modelElem.Set(reflect.ValueOf(model))
		}
		return err
	}

	err = q.One(result)
	if err != nil {
		return err
	}

	model.doc = result
	val := reflect.ValueOf(result).Elem()
	modelVal := val.FieldByName("Model")
	modelVal.Set(reflect.ValueOf(model))

	return err
}

type Model struct {
	C     *mgo.Collection
	doc   interface{}
	isNew bool
}

func (m *Model) Save() {
}
