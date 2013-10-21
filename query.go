package Sleep

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"reflect"
	"strings"
)

type Query struct {
	query         interface{}
	selection     interface{}
	skip          int
	limit         int
	sort          []string
	populate      map[string]*Query
	path          string
	z             *Sleep
	c             *mgo.Collection
	populated     map[string]interface{}
	isPopOp       bool
	parentStruct  interface{}
	populateField interface{}
	isSlice       bool
	popSchema     string
}

func (q *Query) populateExec(parentStruct interface{}) error {
	for key, val := range q.populate {
		val.parentStruct = parentStruct
		val.findPopulatePath(key)
		document, ok := q.z.documents[val.popSchema]
		if !ok {
			panic("Unable to find `" + val.popSchema + "` schema. Was it registered?")
		}

		var schemaStruct interface{}
		if val.isSlice {
			ids := val.populateField.([]bson.ObjectId)
			if len(ids) == 0 {
				return nil
			}

			schemaType := reflect.PtrTo(reflect.TypeOf(document.schemaStruct))
			slicedType := reflect.SliceOf(schemaType)
			schemaStruct = reflect.New(slicedType).Interface()
			val.query = M{"_id": M{"$in": ids}}

		} else {
			schemaStruct = &document.schemaStruct
			id := val.populateField.(bson.ObjectId)
			val.query = M{"_id": id}
		}

		err := val.Exec(schemaStruct)
		if err == mgo.ErrNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		parentModel := reflect.ValueOf(val.parentStruct).Elem().FieldByName("Document").Interface().(Document)
		parentModel.populated[key] = schemaStruct
	}
	return nil
}

// Populate sets the fields to be automatically populated based on the field's bson.ObjectId value.
//
// This function takes a variable number of arguments.
// Each argument must be the full path to the field to be populated.
// The field to be populated can be either of type bson.ObjectId or []bson.ObjectId.
// The field must also have a tag with the key "model" and a case sensative value with the name of the model.
//
// Example:
//
//	type Contact struct {
//		BusinessPartner   bson.ObjectId
//		Competitors       []bson.ObjectId
//	}
//
//	type Person struct {
//		Name           string
//		PhoneNumber    string
//		Friend         bson.ObjectId    `model:"Person"`
//		Acquaintances  []bson.ObjectId  `model:"Person"`
//		Contacts       []Contact
//	}
//
//	sleep.FindId("...").Populate("Friend", "Acquaintances").Exec(personResult)
//
// The path argument can also describe embeded structs. Every step into an embeded struct is seperated by a "."
//
// Example:
//
//	sleep.FindId("...").Populate("Contacts.BusinessPartner", "Contacts.Competitors").Exec(personResult)
//
func (q *Query) Populate(fields ...string) *Query {
	for _, elem := range fields {
		q.populate[elem] = &Query{isPopOp: true,
			populate:  make(map[string]*Query),
			populated: make(map[string]interface{}), z: q.z, c: q.c}
	}
	return q
}

// PopulateQuery does the same thing the Populate function does, except it only takes one field path at a time and the second parameter is a value of type *Sleep.Query
//
// Example (continuing with the Populate example):
//
//	//I want 10 acquantances that are older than 30 and sorted by their names
//	popQuery := sleep.Find(bson.M{"age": bson.M{"$gt": 30} }).Limit(10).Sort("name")
//	sleep.FindId("...").PopulateQuery("Acquaintances", popQuery).Exec(personResult)
//
// You may also call Populate() and PopulateQuery() on the populate Query value to run auto-population on the populated results.
//
//	popQuery := sleep.Find(bson.M{"age": bson.M{"$gt": 30} }).Limit(10).Sort("name").Populate("Friend")
//
func (q *Query) PopulateQuery(field string, query *Query) *Query {
	query.isPopOp = true
	query.populate = make(map[string]*Query)
	query.populated = make(map[string]interface{})
	query.z = q.z
	query.c = q.c
	q.populate[field] = query
	return q
}

func (q *Query) findPopulatePath(path string) {
	parts := strings.Split(path, ".")
	resultVal := reflect.ValueOf(q.parentStruct).Elem()

	var refVal reflect.Value
	partsLen := len(parts)
	for i := 0; i < partsLen; i++ {
		elem := parts[i]
		if i == 0 {
			refVal = resultVal.FieldByName(elem)
			structTag, _ := resultVal.Type().FieldByName(elem)
			q.popSchema = structTag.Tag.Get(q.z.modelTag)
		} else if i == partsLen-1 {
			structTag, _ := refVal.Type().FieldByName(elem)
			q.popSchema = structTag.Tag.Get(q.z.modelTag)
			refVal = refVal.FieldByName(elem)
		} else {
			//refVal = handleSlice(refVal, elem, path)
			refVal = refVal.FieldByName(elem)
		}

		if !refVal.IsValid() {
			panic("field `" + elem + "` not found in populate path `" + path + "`")
		}
	}

	if refVal.Kind() == reflect.Slice {
		q.isSlice = true
	}
	q.populateField = refVal.Interface()
}

/*
func handleSlice(parent reflect.Value, childField string, path string) reflect.Value {
	thisElem := parent.FieldByName(childField)
	if !thisElem.IsValid() {
		panic("field `" + childField + "` not found in populate path `" + path + "`")
	}

	thisType := parent.FieldByName(childField).Type()
	thisKind := thisType.Kind()
	if thisKind == reflect.Slice {
		//create an element with the type of the slice's elements so that we
		//can continue to walk the path
		thisElem = reflect.New(thisType.Elem()).Elem()
	}

	//handle slices of slices
	if thisElem.Type().Kind() == reflect.Slice {
		thisElem = handleSlice(thisElem, childField, path)
	}

	//handleSlice is only be called if the current field is neither the firt, nor the last.
	//If this field is somewhere in the middle of the path, it should either of type Struct,
	//or of type slice. Since we already handle slice situations *see recursion above*, the field
	//could only be a struct.
	if thisElem.Type().Kind() != reflect.Struct {
		panic("Malformed populate path: " + path)
	}
	return thisElem
}
*/

// Exec executes the query.
//
// What collection to query on is determined by the result parameter.
// Exec does the job of both mgo.Collection.One() and mgo.Collection.All().
//
// Example 1 (Equivalent to mgo.Collection.One() ):
//
//		type Foo struct {...}
//		foo := &Foo{} //foo is a pointer to the value for a single Foo struct
//		sleep.Find(bson.M{"location:": "Earth"}).Exec(foo)
//
// Example 2 (Equivalent to mgo.Collection.All() ):
//
//		type Foo struct {...}
//		foo := []*Foo{} //foo is the value for a slice of pointers to Foo structs
//		sleep.Find(bson.M{"location:": "Earth"}).Exec(&foo)
//		//Another example showing further filtering
//		sleep.Find(bson.M{"location:": "Earth"}).Sort("name", "age").Limit(200).Exec(&foo)
//
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

	q := query.c.Find(query.query)

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

	document := query.z.documents[structName]
	var err error
	if isSlice == true {
		err = q.All(result)
		if err != nil {
			if err == mgo.ErrNotFound {
				return nil
			} else {
				return err
			}
		}
		val := reflect.ValueOf(result).Elem()
		elemCount := val.Len()
		for i := 0; i < elemCount; i++ {
			documentCpy := query.z.documents[structName]
			sliceElem := val.Index(i)
			documentCpy.schema = sliceElem.Interface()
			documentCpy.Virtual = newVirtual()
			documentCpy.Model = query.z.models[structName]
			modelElem := sliceElem.Elem().FieldByName("Document")
			modelElem.Set(reflect.ValueOf(documentCpy))
		}
		return err
	}

	err = q.One(result)
	document.schema = result
	document.Virtual = newVirtual()
	document.Model = query.z.models[structName]
	val := reflect.ValueOf(result).Elem()
	documentVal := val.FieldByName("Document")
	documentVal.Set(reflect.ValueOf(document))

	if err == mgo.ErrNotFound {
		documentVal.FieldByName("Found").SetBool(false)
		return nil
	}

	if err != nil {
		if err == mgo.ErrNotFound {
			document.Found = false
			return nil
		} else {
			return err
		}
	}

	return query.populateExec(result)
}

// Select enables selecting which fields should be retrieved for the results found.
// For example, the following query would only retrieve the name field:
//
//		err := sleep.Find(bson.M{"age": 50}).Select(bson.M{"name": 1}).Exec(result)
//
// Note 1: The _id field is always selected.. unless explicitly stated otherwise
//
// Note 2**: If only some fields are selected for retrieval and then the Save() is called on the document, the fields not retrieved will be blank and will overwrite the database values with the default value for their respective types.
func (q *Query) Select(selection interface{}) *Query {
	q.selection = selection
	return q
}

// Skip skips over the n initial documents from the query results.
// Using Skip only makes sense with ordered results and capped collections where documents are naturally ordered by insertion time.
func (q *Query) Skip(skip int) *Query {
	q.skip = skip
	return q
}

// Limit sets the maximum number of document the database should return
func (q *Query) Limit(lim int) *Query {
	q.limit = lim
	return q
}

// Sort sets the fields by which the database should sort the query results
//
// Example:
// For example:
//
//		query1 := sleep.Find(nil).Sort("firstname", "lastname")
//		query2 := sleep.Find(nil).Sort("-age")
//		query3 := sleep.Find(nil).Sort("$natural")
//
//
// Further reading: http://godoc.org/labix.org/v2/mgo#Query.Sort
func (q *Query) Sort(fields ...string) *Query {
	q.sort = fields
	return q
}
