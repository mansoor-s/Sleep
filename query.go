package Sleep

import (
	"fmt"
	"reflect"
	"strings"
)

type Query struct {
	query     interface{}
	selection interface{}
	skip      int
	limit     int
	sort      []string
	populate  map[string]*Query
	path      string
	z         *Sleep
	populated map[string]interface{}
	isPopOp   bool
	parentStruct intrface{}
	populateField interface{}
	isSlice bool
}

func (q *Query) populate() error {
	for key, val := range q.populate {
		//figure out type of the field being populated
	}

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

func (q *Query) Populate(fields ...string) *Query {
	for _, elem := range fields {
		field, isSlice, typ := findPopulatePath(field)
		q.populate[elem] = &Query{isPopOp: true, populateField: field, 
			,isSlice: isSlice}
		q.populate[field] = query
	}
	return q
}

func (q *Query) PopulateQuery(field string, query *Query) *Query {
	query.populateField, query.isSlice = findPopulatePath(field)
	query.isPopOp = true
	q.populate[field] = query
	return q
}

func (q *Query) findPopulatePath(path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	resultVal := reflect.ValueOf(q.parentStruct).Elem()
	
	var refVal reflect.Value
	for i, elem := range parts {
		if i == 0 {
			refVal = resultVal.fieldByName(elem)
		} else {
			refVal = refVal.fieldByName(elem)
		}
		
		if !refVal.IsValid() {
			panic("field `" + elem + "` not found in populate path `" + path + "`")
		}
	}

	isSlice := false
	if refVal.Kind() == reflect.Slice {
		isSlice = true
		elemType := refVal.Type().Elem()
	}

	return refVal, isSlice
}

func (query *Query) Exec(result interface{}) error {
	if reflect.TypeOf(result).Kind() != reflect.Ptr {
		panic(fmt.Sprintf("Expecting a pointer type but recieved %v. If you are passing in a slice, make sure to pass a pointer to it.", reflect.TypeOf(result)))
	}
	typ := reflect.TypeOf(result).Elem()
	var structName string
	isSlice := false
	if typ.Kind() == reflect.Slice {
		structName = typ.Elem().Elem().Name()
		isSlice = true
	} else {
		structName = typ.Name()
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

		val := reflect.ValueOf(result).Elem()
		elemCount := val.Len()
		for i := 0; i < elemCount; i++ {
			modelCpy := query.z.models[structName]
			sliceElem := val.Index(i)
			modelCpy.doc = sliceElem.Interface()
			modelElem := sliceElem.Elem().FieldByName("Model")
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


