/*
 * Copyright 2019 ObjectBox Ltd. All rights reserved.
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

// TODO how to handle null values?

// CppBindingTemplate is used to generated the binding code
var CppBindingTemplate = template.Must(template.New("binding").Funcs(funcMap).Parse(
	`// Code generated by ObjectBox; DO NOT EDIT.

#pragma once

#include <stdbool.h>
#include <stdint.h>

#include "flatbuffers/flatbuffers.h"
#include "objectbox.h"
{{range $entity := .Model.EntitiesWithMeta}}{{with $entity.Meta.CppNamespaceStart}}
{{.}}{{end}}
struct {{$entity.Meta.CppName}} {
	{{- range $property := $entity.Properties}}
	{{$property.Meta.CppType}} {{$property.Meta.CppName}};
	{{- end}}
};

struct {{$entity.Meta.CppName}}_ {
{{- range $property := $entity.Properties}}
	static const obx_schema_id {{$property.Meta.CppName}} = {{$property.Id.GetId}};
{{- end}}

    static constexpr obx_schema_id entityId() { return {{$entity.Id.GetId}}; }

	static {{$entity.Meta.CppName}} entityType();

    static void setObjectId({{$entity.Meta.CppName}}& object, obx_id newId) { object.{{$entity.IdProperty.Meta.CppName}} = newId; }

	/// Write given object to the FlatBufferBuilder
	static void toFlatBuffer(flatbuffers::FlatBufferBuilder& fbb, const {{$entity.Meta.CppName}}& object) {
		fbb.Clear();
		{{- range $property := $entity.Properties}}{{$factory := $property.Meta.FbOffsetFactory}}{{if $factory}}
		auto offset{{$property.Meta.CppName}} = fbb.{{$factory}}(object.{{$property.Meta.CppName}});
		{{- end}}{{end}}
		flatbuffers::uoffset_t fbStart = fbb.StartTable();
		{{range $property := $entity.Properties}}
		{{- if $property.Meta.FbOffsetFactory}}fbb.AddOffset({{$property.FbvTableOffset}}, offset{{$property.Meta.CppName}});
		{{- else if eq "bool" $property.Meta.CppType}}fbb.TrackField({{$property.FbvTableOffset}}, fbb.PushElement<uint8_t>(object.{{$property.Meta.CppName}} ? 1 : 0));
		{{- else}}fbb.TrackField({{$property.FbvTableOffset}}, fbb.PushElement<{{$property.Meta.CppType}}>(object.{{$property.Meta.CppName}}));
		{{- end}}
		{{end -}}
		flatbuffers::Offset<flatbuffers::Table> offset;
		offset.o = fbb.EndTable(fbStart);
		fbb.Finish(offset);
	}

	/// Read an object from a valid FlatBuffer
	static {{$entity.Meta.CppName}} fromFlatBuffer(const void* data, size_t size) {
		{{$entity.Meta.CppName}} object;
		fromFlatBuffer(data, size, object);
		return object;
	}

	/// Read an object from a valid FlatBuffer
	static std::unique_ptr<{{$entity.Meta.CppName}}> newFromFlatBuffer(const void* data, size_t size) {
		auto object = std::unique_ptr<{{$entity.Meta.CppName}}>(new {{$entity.Meta.CppName}}());
		fromFlatBuffer(data, size, *object);
		return object;
	}

	/// Read an object from a valid FlatBuffer
	static void fromFlatBuffer(const void* data, size_t size, {{$entity.Meta.CppName}}& outObject) {
		const auto* table = flatbuffers::GetRoot<flatbuffers::Table>(data);
		assert(table);
		{{range $property := $entity.Properties}}
		{{- if eq "std::vector<std::string>" $property.Meta.CppType}}{
			auto* ptr = table->GetPointer<const flatbuffers::Vector<flatbuffers::Offset<flatbuffers::String>>*>({{$property.FbvTableOffset}});
			if (ptr) {
				outObject.{{$property.Meta.CppName}}.reserve(ptr->size());
				for (size_t i = 0; i < ptr->size(); i++) {
					auto* itemPtr = ptr->Get(i);
					if (itemPtr) outObject.{{$property.Meta.CppName}}.emplace_back(itemPtr->c_str());
				}
			}
		}{{else if eq "std::string" $property.Meta.CppType}}{
			auto* ptr = table->GetPointer<const flatbuffers::String*>({{$property.FbvTableOffset}});
			if (ptr) outObject.{{$property.Meta.CppName}}.assign(ptr->c_str());
		}{{else if $property.Meta.FbIsVector}}{
			auto* ptr = table->GetPointer<const {{$property.Meta.FbOffsetType}}*>({{$property.FbvTableOffset}});
			if (ptr) outObject.{{$property.Meta.CppName}}.assign(ptr->begin(), ptr->end());
		}{{- else if eq "bool" $property.Meta.CppType}}outObject.{{$property.Meta.CppName}} = table->GetField<uint8_t>({{$property.FbvTableOffset}}, 0) != 0;
		{{- else}}outObject.{{$property.Meta.CppName}} = table->GetField<{{$property.Meta.CppType}}>({{$property.FbvTableOffset}}, {{$property.Meta.FbDefaultValue}});
		{{- end}}
		{{end}}
	}
};
{{with $entity.Meta.CppNamespaceEnd}}{{.}}{{end -}}
{{end}}
`))