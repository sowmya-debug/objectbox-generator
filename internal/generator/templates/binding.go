/*
 * Copyright 2018 ObjectBox Ltd. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package templates

import (
	"text/template"
)

var BindingTemplate = template.Must(template.New("binding").Funcs(funcMap).Parse(
	`// Code generated by ObjectBox; DO NOT EDIT.

{{define "property-getter"}}{{/* used in Load*/}}
	{{- if .Converter}}{{.Converter}}ToEntityProperty(
	{{- else if .CastOnWrite}}{{.CastOnWrite}}({{end}}
		{{- if eq .GoType "bool"}} table.GetBoolSlot({{.FbvTableOffset}}, false)
    	{{- else if eq .GoType "int"}} int(table.GetUint64Slot({{.FbvTableOffset}}, 0))
    	{{- else if eq .GoType "uint"}} uint(table.GetUint64Slot({{.FbvTableOffset}}, 0))
		{{- else if eq .GoType "rune"}} rune(table.GetInt32Slot({{.FbvTableOffset}}, 0))
		{{- else if eq .FbType "UOffsetT"}} fbutils.Get{{.ObType}}Slot(table, {{.FbvTableOffset}})
    	{{- else}} table.Get{{.GoType | StringTitle}}Slot({{.FbvTableOffset}}, 0)
    	{{- end}}
	{{- if or .Converter .CastOnWrite}}){{end}}
{{- end -}}

{{define "property-converter-encode"}}{{/* used in Flatten*/ -}}
	{{- if .Converter}}{{.Converter}}ToDatabaseValue(obj.{{.Name}})
	{{- else if .CastOnRead}}{{.CastOnRead}}(obj.{{.Name}})
	{{- else}}obj.{{.Name}}{{end}}
{{- end -}}

{{define "put-related-entity"}}{{/* used in PutRelated*/ -}}
	rId, err := {{.Target.Name}}Binding.GetId(rel)
	if err != nil {
		return err
	} else if rId == 0 {
		if rCursor, err := txn.CursorForName("{{.Target.Name}}"); err != nil {
			return err 
		} else if rId, err = rCursor.Put(rel); err != nil { 
			return err 
		} 
	}
{{- end -}}

				

package {{.Binding.Package.Name}}

import (
	"github.com/google/flatbuffers/go"
	"github.com/objectbox/objectbox-go/objectbox"
	"github.com/objectbox/objectbox-go/objectbox/fbutils"
	{{range $path := .Binding.Imports -}}
		"{{$path}}"
	{{end}}
)

{{range $entity := .Binding.Entities -}}
{{$entityNameCamel := $entity.Name | StringCamel -}}
type {{$entityNameCamel}}_EntityInfo struct {
	Id objectbox.TypeId
	Uid uint64
}

var {{$entity.Name}}Binding = {{$entityNameCamel}}_EntityInfo {
	Id: {{$entity.Id}}, 
	Uid: {{$entity.Uid}},
}

// {{$entity.Name}}_ contains type-based Property helpers to facilitate some common operations such as Queries. 
var {{$entity.Name}}_ = struct {
	{{range $property := $entity.Properties -}}
    {{$property.Name}} *objectbox.Property{{$property.GoType | TypeIdentifier}}
    {{end -}}
}{
	{{range $property := $entity.Properties -}}
    {{$property.Name}}: &objectbox.Property{{$property.GoType | TypeIdentifier}}{
		BaseProperty: &objectbox.BaseProperty{
			Id: {{$property.Id}},
			Entity: &objectbox.Entity{
				Id: {{$entity.Id}},
			},
		},
	},
    {{end -}}
}

// GeneratorVersion is called by ObjectBox to verify the compatibility of the generator used to generate this code	
func ({{$entityNameCamel}}_EntityInfo) GeneratorVersion() int {
	return {{$.GeneratorVersion}}
}

// AddToModel is called by ObjectBox during model build
func ({{$entityNameCamel}}_EntityInfo) AddToModel(model *objectbox.Model) {
    model.Entity("{{$entity.Name}}", {{$entity.Id}}, {{$entity.Uid}})
    {{range $property := $entity.Properties -}}
    model.Property("{{$property.ObName}}", objectbox.PropertyType_{{$property.ObType}}, {{$property.Id}}, {{$property.Uid}})
    {{if len $property.ObFlags -}}
        model.PropertyFlags(
        {{- range $key, $flag := $property.ObFlags -}}
            {{if gt $key 0}} | {{end}}objectbox.PropertyFlags_{{$flag}}
        {{- end}})
    {{end -}}
	{{if $property.Relation}}model.PropertyRelation("{{$property.Relation.Target.Name}}", {{$property.Index.Id}}, {{$property.Index.Uid}})
	{{else if $property.Index}}model.PropertyIndex({{$property.Index.Id}}, {{$property.Index.Uid}})
    {{end -}}
    {{end -}}
    model.EntityLastPropertyId({{$entity.LastPropertyId.GetId}}, {{$entity.LastPropertyId.GetUid}})
	{{range $relation := $entity.Relations -}}
    model.Relation({{$relation.Id}}, {{$relation.Uid}}, {{$relation.Target.Id}}, {{$relation.Target.Uid}})
    {{end -}}
}

// GetId is called by ObjectBox during Put operations to check for existing ID on an object
func ({{$entityNameCamel}}_EntityInfo) GetId(object interface{}) (uint64, error) {
	{{- if $.Options.ByValue}}
		if obj, ok := object.(*{{$entity.Name}}); ok {
			return {{template "property-converter-encode" $entity.IdProperty}}, nil
		} else {
			return {{if $entity.IdProperty.Converter}}{{$entity.IdProperty.Converter}}ToDatabaseValue({{end -}}
					object.({{$entity.Name}}).{{$entity.IdProperty.Path}}{{if $entity.IdProperty.Converter}}){{end}}, nil
		}
	{{- else -}}
		return {{if $entity.IdProperty.Converter}}{{$entity.IdProperty.Converter}}ToDatabaseValue({{end -}}
				object.(*{{$entity.Name}}).{{$entity.IdProperty.Path}}{{if $entity.IdProperty.Converter}}){{end}}, nil
	{{- end}}
}

// SetId is called by ObjectBox during Put to update an ID on an object that has just been inserted
func ({{$entityNameCamel}}_EntityInfo) SetId(object interface{}, id uint64) {
	{{- if $.Options.ByValue}}
		if obj, ok := object.(*{{$entity.Name}}); ok {
			obj.{{$entity.IdProperty.Path}} =   
				{{- if $entity.IdProperty.Converter}}{{$entity.IdProperty.Converter}}ToEntityProperty({{end}}id{{if $entity.IdProperty.Converter}}){{end}}
		} else {
			// NOTE while this can't update, it will at least behave consistently (panic in case of a wrong type)
			_ = object.({{$entity.Name}}).{{$entity.IdProperty.Path}}
		}
	{{- else -}}
		object.(*{{$entity.Name}}).{{$entity.IdProperty.Path}} =  
			{{- if $entity.IdProperty.Converter}}{{$entity.IdProperty.Converter}}ToEntityProperty({{end}}id{{if $entity.IdProperty.Converter}}){{end}}
	{{- end}}
}

// PutRelated is called by ObjectBox to put related entities before the object itself is flattened and put
func ({{$entityNameCamel}}_EntityInfo) PutRelated(txn *objectbox.Transaction, object interface{}, id uint64) error {
	{{- /* TODO ideally we should be using BoxForTarget() with a manually assigned txn */}}
	{{- range $field := .Fields}}
	{{- if $field.SimpleRelation}}
		if rel := {{if not $field.IsPointer}}&{{end}}object.(*{{$entity.Name}}).{{$field.Name}}; rel != nil {
			{{template "put-related-entity" $field.SimpleRelation}}
			// NOTE Put/PutAsync() has a side-effect of setting the rel.ID, so at this point, it is already set
		}
	{{- else if $field.StandaloneRelation}}
		if cursor, err := txn.CursorForName("{{$entity.Name}}"); err != nil {
			panic(err)
		} else if rSlice := object.(*{{$entity.Name}}).{{$field.Name}}; rSlice != nil {
			// get id from the object, if inserting, it would be 0 even if the argument id is already non-zero
			// this saves us an unnecessary request to RelationIds for new objects (there can't be any relations yet)
			objId, err := {{$entity.Name}}Binding.GetId(object)
			if err != nil {
				return err
			}

			// make a map of related target entity IDs, marking those that were originally related but should be removed
			var idsToRemove = make(map[uint64]bool)

			if objId != 0 {
				if oldRelIds, err := cursor.RelationIds({{$field.StandaloneRelation.Id}}, id); err != nil {
					return err
				} else {
					for _, rId := range oldRelIds {
						idsToRemove[rId] = true
					}
				}
			}

			// walk over the current related objects, mark those that still exist, add the new ones
			{{if $field.StandaloneRelation.Target.IsPointer -}}
			for _, rel := range rSlice {
			{{- else -}}
			for k := range rSlice {
				var rel = &rSlice[k] // take a pointer to the slice element so that it is updated during Put()
			{{- end}}
				{{template "put-related-entity" $field.StandaloneRelation}}

				if idsToRemove[rId] {
					// old relation that still exists, keep it
					delete(idsToRemove, rId) 
				} else {
					// new relation, add it
					if err := cursor.RelationPut({{$field.StandaloneRelation.Id}}, id, rId); err != nil {
						return err
					}
				}
			}

			// remove those that were not found in the rSlice but were originally related to this entity
			for rId := range idsToRemove {
				if err := cursor.RelationRemove({{$field.StandaloneRelation.Id}}, id, rId); err != nil {
					return err
				}
			}
		}
	{{- end}}{{end}}
	return nil
}

// Flatten is called by ObjectBox to transform an object to a FlatBuffer
func ({{$entityNameCamel}}_EntityInfo) Flatten(object interface{}, fbb *flatbuffers.Builder, id uint64) {
    {{if $entity.HasNonIdProperty}}obj := object.(*{{$entity.Name}}){{end -}}

    {{- range $property := $entity.Properties}}{{if eq $property.FbType "UOffsetT"}}
    var offset{{$property.Name}} = fbutils.Create{{$property.ObType}}Offset(fbb, {{template "property-converter-encode" $property}})
	{{- end}}{{end}}

	{{- /* store relations, TODO return error, ideally we should be using BoxForTarget() with a manually assigned txn */}}
	{{- range $field := .Fields}}
	{{if $field.SimpleRelation}}
		var rId{{$field.Property.Name}} uint64
		if rel := {{if not $field.IsPointer}}&{{end}}obj.{{$field.Name}}; rel != nil {
			if rId, err := {{$field.SimpleRelation.Target.Name}}Binding.GetId(rel); err != nil {
				panic(err) // this must never happen but let's keep the check just to be sure
			} else {
				rId{{$field.Property.Name}} = rId
			}
		}
	{{- else if $field.Property}}{{if $field.Property.Relation}}{{/* manual relation links (just ID)*/}}
		var rId{{$field.Property.Name}} = {{template "property-converter-encode" $field.Property}} 
	{{- end}}{{end}}{{end}}

    // build the FlatBuffers object
    fbb.StartObject({{$entity.LastPropertyId.GetId}})
    {{range $property := $entity.Properties -}}
    fbutils.Set{{$property.FbType}}Slot(fbb, {{$property.FbSlot}},
		{{- if $property.Relation}}rId{{$property.Name}})
        {{- else if eq $property.FbType "UOffsetT"}} offset{{$property.Name}})
        {{- else if eq $property.Name $entity.IdProperty.Name}} id)
        {{- else if eq $property.GoType "int"}} int64({{template "property-converter-encode" $property}}))
        {{- else if eq $property.GoType "uint"}} uint64({{template "property-converter-encode" $property}}))
        {{- else}} {{template "property-converter-encode" $property}})
        {{- end}}
    {{end -}}
}

// Load is called by ObjectBox to load an object from a FlatBuffer 
func ({{$entityNameCamel}}_EntityInfo) Load(txn *objectbox.Transaction, bytes []byte) interface{} {
	var table = &flatbuffers.Table{
		Bytes: bytes,
		Pos:   flatbuffers.GetUOffsetT(bytes),
	}
	var id = table.Get{{$entity.IdProperty.GoType | StringTitle}}Slot({{$entity.IdProperty.FbvTableOffset}}, 0)

	{{- /* load relations, TODO return error, ideally we should be using BoxForTarget() with a manually assigned txn */}}
	{{- range $field := .Fields}}{{if $field.SimpleRelation}}

		var rel{{$field.Name}} *{{$field.Type}}
		if rId := {{template "property-getter" $field.Property}}; rId > 0 { 
			if cursor, err := txn.CursorForName("{{$field.SimpleRelation.Target.Name}}"); err != nil {
				panic(err) 
			} else if relObject, err := cursor.Get(rId); err != nil {
				panic(err) 
			} else if relObj, ok := relObject.(*{{$field.Type}}); ok {
				rel{{$field.Name}} = relObj
			} else {
				var relObj = relObject.({{$field.Type}})
				rel{{$field.Name}} = &relObj
			}
		{{if not $field.IsPointer -}} 
		} else {
			rel{{$field.Name}} = &{{$field.Type}}{}
		{{end -}}
		}
	{{- else if $field.StandaloneRelation}}
	
		var rel{{$field.Name}} {{$field.Type}} 
		if cursor, err := txn.CursorForName("{{$entity.Name}}"); err != nil {
			panic(err) 
		} else if rSlice, err := cursor.RelationGetAll({{$field.StandaloneRelation.Id}}, {{$field.StandaloneRelation.Target.Id}}, id); err != nil {
			panic(err)
		} else {
			rel{{$field.Name}} = {{if $field.IsPointer}}&{{end}}rSlice.({{$field.Type}})
		}
	{{- end}}{{end}}

	return &{{$entity.Name}}{
	{{- block "fields-initializer" $entity}}
		{{- range $field := .Fields}}
			{{$field.Name}}: 
				{{- if or $field.SimpleRelation}}{{if not $field.IsPointer}}*{{end}}rel{{$field.Name}}
				{{- else if $field.StandaloneRelation}}rel{{$field.Name}}
        		{{- else if $field.IsId}}{{with $field.Property}}
					{{- if .Converter}}{{.Converter}}ToEntityProperty(
					{{- else if .CastOnWrite}}{{.CastOnWrite}}({{end -}}
					id
					{{- if or .Converter .CastOnWrite}}){{end}}{{end}}
				{{- else if $field.Property}}{{template "property-getter" $field.Property}}
				{{- else}}{{if $field.IsPointer}}&{{end}}{{$field.Type}}{ {{template "fields-initializer" $field}} }{{end}},
		{{- end}}
	{{end}}
	}
}

// MakeSlice is called by ObjectBox to construct a new slice to hold the read objects  
func ({{$entityNameCamel}}_EntityInfo) MakeSlice(capacity int) interface{} {
	return make([]{{if not $.Options.ByValue}}*{{end}}{{$entity.Name}}, 0, capacity)
}

// AppendToSlice is called by ObjectBox to fill the slice of the read objects
func ({{$entityNameCamel}}_EntityInfo) AppendToSlice(slice interface{}, object interface{}) interface{} {
	return append(slice.([]{{if not $.Options.ByValue}}*{{end}}{{$entity.Name}}), {{if $.Options.ByValue}}*{{end}}object.(*{{$entity.Name}}))
}

// Box provides CRUD access to {{$entity.Name}} objects
type {{$entity.Name}}Box struct {
	*objectbox.Box
}

// BoxFor{{$entity.Name}} opens a box of {{$entity.Name}} objects 
func BoxFor{{$entity.Name}}(ob *objectbox.ObjectBox) *{{$entity.Name}}Box {
	return &{{$entity.Name}}Box{
		Box: ob.InternalBox({{$entity.Id}}),
	}
}

// Put synchronously inserts/updates a single object.
// In case the {{$entity.IdProperty.Path}} is not specified, it would be assigned automatically (auto-increment).
// When inserting, the {{$entity.Name}}.{{$entity.IdProperty.Path}} property on the passed object will be assigned the new ID as well.
func (box *{{$entity.Name}}Box) Put(object *{{$entity.Name}}) (uint64, error) {
	return box.Box.Put(object)
}

// PutAsync asynchronously inserts/updates a single object.
// When inserting, the {{$entity.Name}}.{{$entity.IdProperty.Path}} property on the passed object will be assigned the new ID as well.
// 
// It's executed on a separate internal thread for better performance.
//
// There are two main use cases:
//
// 1) "Put & Forget:" you gain faster puts as you don't have to wait for the transaction to finish.
//
// 2) Many small transactions: if your write load is typically a lot of individual puts that happen in parallel,
// this will merge small transactions into bigger ones. This results in a significant gain in overall throughput.
//
//
// In situations with (extremely) high async load, this method may be throttled (~1ms) or delayed (<1s).
// In the unlikely event that the object could not be enqueued after delaying, an error will be returned.
//
// Note that this method does not give you hard durability guarantees like the synchronous Put provides.
// There is a small time window (typically 3 ms) in which the data may not have been committed durably yet.
func (box *{{$entity.Name}}Box) PutAsync(object *{{$entity.Name}}) (uint64, error) {
	return box.Box.PutAsync(object)
}

// PutAll inserts multiple objects in single transaction.
// In case {{$entity.IdProperty.Path}}s are not set on the objects, they would be assigned automatically (auto-increment).
// 
// Returns: IDs of the put objects (in the same order).
// When inserting, the {{$entity.Name}}.{{$entity.IdProperty.Path}} property on the objects in the slice will be assigned the new IDs as well.
//
// Note: In case an error occurs during the transaction, some of the objects may already have the {{$entity.Name}}.{{$entity.IdProperty.Path}} assigned    
// even though the transaction has been rolled back and the objects are not stored under those IDs.
//
// Note: The slice may be empty or even nil; in both cases, an empty IDs slice and no error is returned.
func (box *{{$entity.Name}}Box) PutAll(objects []{{if not $.Options.ByValue}}*{{end}}{{$entity.Name}}) ([]uint64, error) {
	return box.Box.PutAll(objects)
}

// Get reads a single object.
//
// Returns nil (and no error) in case the object with the given ID doesn't exist.
func (box *{{$entity.Name}}Box) Get(id uint64) (*{{$entity.Name}}, error) {
	object, err := box.Box.Get(id)
	if err != nil {
		return nil, err
	} else if object == nil {
		return nil, nil
	}
	return object.(*{{$entity.Name}}), nil
}

// Get reads all stored objects
func (box *{{$entity.Name}}Box) GetAll() ([]{{if not $.Options.ByValue}}*{{end}}{{$entity.Name}}, error) {
	objects, err := box.Box.GetAll()
	if err != nil {
		return nil, err
	}
	return objects.([]{{if not $.Options.ByValue}}*{{end}}{{$entity.Name}}), nil
}

// Remove deletes a single object
func (box *{{$entity.Name}}Box) Remove(object *{{$entity.Name}}) (err error) {
	return box.Box.Remove({{if $entity.IdProperty.Converter}}{{$entity.IdProperty.Converter}}ToDatabaseValue({{end -}}
					object.{{$entity.IdProperty.Path}}{{if $entity.IdProperty.Converter}}){{end}})
}

// Creates a query with the given conditions. Use the fields of the {{$entity.Name}}_ struct to create conditions.
// Keep the *{{$entity.Name}}Query if you intend to execute the query multiple times.
// Note: this function panics if you try to create illegal queries; e.g. use properties of an alien type.
// This is typically a programming error. Use QueryOrError instead if you want the explicit error check.
func (box *{{$entity.Name}}Box) Query(conditions ...objectbox.Condition) *{{$entity.Name}}Query {
	return &{{$entity.Name}}Query{
		box.Box.Query(conditions...),
	}
}

// Creates a query with the given conditions. Use the fields of the {{$entity.Name}}_ struct to create conditions.
// Keep the *{{$entity.Name}}Query if you intend to execute the query multiple times.
func (box *{{$entity.Name}}Box) QueryOrError(conditions ...objectbox.Condition) (*{{$entity.Name}}Query, error) {
	if query, err := box.Box.QueryOrError(conditions...); err != nil {
		return nil, err
	} else {
		return &{{$entity.Name}}Query{query}, nil
	}
}

// Query provides a way to search stored objects
//
// For example, you can find all {{$entity.Name}} which {{$entity.IdProperty.Name}} is either 42 or 47:
// 		box.Query({{$entity.Name}}_.{{$entity.IdProperty.Name}}.In(42, 47)).Find()
type {{$entity.Name}}Query struct {
	*objectbox.Query
}

// Find returns all objects matching the query
func (query *{{$entity.Name}}Query) Find() ([]{{if not $.Options.ByValue}}*{{end}}{{$entity.Name}}, error) {
	objects, err := query.Query.Find()
	if err != nil {
		return nil, err
	}
	return objects.([]{{if not $.Options.ByValue}}*{{end}}{{$entity.Name}}), nil
}

// Offset defines the index of the first object to process (how many objects to skip)
func (query *{{$entity.Name}}Query) Offset(offset uint64) *{{$entity.Name}}Query {
	query.Query.Offset(offset)
	return query
}

// Limit sets the number of elements to process by the query
func (query *{{$entity.Name}}Query) Limit(limit uint64) *{{$entity.Name}}Query {
	query.Query.Limit(limit)
	return query
}
{{end -}}`))
