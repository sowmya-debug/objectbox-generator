// automatically generated by the ObjectBox, do not modify

package object

import (
	"github.com/google/flatbuffers/go"
	"github.com/objectbox/objectbox-go/objectbox"
	"github.com/objectbox/objectbox-go/objectbox/fbutils"
)

type ABinding struct {
}

func (ABinding) AddToModel(model *objectbox.Model) {
	model.Entity("A", 1, 8717895732742165505)
	model.Property("Id", objectbox.PropertyType_Long, 1, 2259404117704393152)
	model.PropertyFlags(objectbox.PropertyFlags_ID)
	model.Property("Name", objectbox.PropertyType_String, 2, 6050128673802995827)
	model.EntityLastPropertyId(2, 6050128673802995827)
}

func asA(entity interface{}) (*A, error) {
	ent, ok := entity.(*A)
	if !ok {
		// Programming error, OK to panic
		// TODO don't panic here, handle in the caller if necessary to panic
		panic("Object has wrong type, expecting 'A'")
	}
	return ent, nil
}

func asAs(entities interface{}) ([]*A, error) {
	ent, ok := entities.([]*A)
	if !ok {
		// Programming error, OK to panic
		// TODO don't panic here, handle in the caller if necessary to panic
		panic("Object has wrong type, expecting 'A'")
	}
	return ent, nil
}

func (ABinding) GetId(entity interface{}) (uint64, error) {
	if ent, err := asA(entity); err != nil {
		return 0, err
	} else {
		return ent.Id, nil
	}
}

func (ABinding) Flatten(entity interface{}, fbb *flatbuffers.Builder, id uint64) {
	ent, err := asA(entity)
	if err != nil {
		// TODO return error and panic in the caller if really, really necessary
		panic(err)
	}

	// prepare the "offset" properties
	var offsetName = fbutils.CreateStringOffset(fbb, ent.Name)

	// build the FlatBuffers object
	fbb.StartObject(2)
	fbb.PrependUint64Slot(0, id, 0)
	fbb.PrependUOffsetTSlot(1, offsetName, 0)
}

func (ABinding) ToObject(bytes []byte) interface{} {
	table := fbutils.GetRootAsTable(bytes, flatbuffers.UOffsetT(0))

	return &A{
		Id:   table.OffsetAsUint64(4),
		Name: table.OffsetAsString(6),
	}
}

func (ABinding) MakeSlice(capacity int) interface{} {
	return make([]*A, 0, capacity)
}

func (ABinding) AppendToSlice(slice interface{}, entity interface{}) interface{} {
	return append(slice.([]*A), entity.(*A))
}

type ABox struct {
	*objectbox.Box
}

func BoxForA(ob *objectbox.ObjectBox) *ABox {
	return &ABox{
		Box: ob.Box(1),
	}
}

func (box *ABox) Get(id uint64) (*A, error) {
	entity, err := box.Box.Get(id)
	if err != nil {
		return nil, err
	} else if entity == nil {
		return nil, nil
	}
	return asA(entity)
}

func (box *ABox) GetAll() ([]*A, error) {
	entities, err := box.Box.GetAll()
	if err != nil {
		return nil, err
	}
	return asAs(entities)
}

func (box *ABox) Remove(entity *A) (err error) {
	return box.Box.Remove(entity.Id)
}
