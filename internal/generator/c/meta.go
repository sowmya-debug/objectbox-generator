/*
 * Copyright (C) 2020 ObjectBox Ltd. All rights reserved.
 * https://objectbox.io
 *
 * This file is part of ObjectBox Generator.
 *
 * ObjectBox Generator is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 * ObjectBox Generator is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with ObjectBox Generator.  If not, see <http://www.gnu.org/licenses/>.
 */

package cgenerator

import (
	"strings"

	"github.com/objectbox/objectbox-generator/internal/generator/binding"
	"github.com/objectbox/objectbox-generator/internal/generator/flatbuffersc/reflection"
	"github.com/objectbox/objectbox-generator/internal/generator/model"
)

type fbsObject struct {
	*binding.Object
	fbsObject *reflection.Object
}

// Merge implements model.EntityMeta interface
func (mo *fbsObject) Merge(entity *model.Entity) model.EntityMeta {
	mo.ModelEntity = entity
	return mo
}

// CppName returns C++ symbol/variable name with reserved keywords suffixed by an underscore
func (mo *fbsObject) CppName() string {
	return cppName(mo.Name)
}

// CName returns CppName() prefixed by a namespace (with underscores)
func (mo *fbsObject) CName() string {
	var prefix string
	if len(mo.Namespace) != 0 {
		prefix = strings.Replace(mo.Namespace, ".", "_", -1) + "_"
	}

	return prefix + mo.CppName()
}

// CppNamespacePrefix returns c++ namespace prefix for symbol definition
func (mo *fbsObject) CppNamespacePrefix() string {
	if len(mo.Namespace) == 0 {
		return ""
	}
	return strings.Join(strings.Split(mo.Namespace, "."), "::") + "::"
}

// CppNamespaceStart returns c++ namespace opening declaration
func (mo *fbsObject) CppNamespaceStart() string {
	if len(mo.Namespace) == 0 {
		return ""
	}

	var nss = strings.Split(mo.Namespace, ".")
	for i, ns := range nss {
		nss[i] = "namespace " + ns + " {"
	}
	return strings.Join(nss, "\n")
}

// CppNamespaceEnd returns c++ namespace closing declaration
func (mo *fbsObject) CppNamespaceEnd() string {
	if len(mo.Namespace) == 0 {
		return ""
	}
	var result = ""
	var nss = strings.Split(mo.Namespace, ".")
	for _, ns := range nss {
		// print in reversed order
		result = "}  // namespace " + ns + "\n" + result
	}
	return result
}

type fbsField struct {
	*binding.Field
	fbsField *reflection.Field
}

// Merge implements model.PropertyMeta interface
func (mp *fbsField) Merge(property *model.Property) model.PropertyMeta {
	mp.ModelProperty = property
	return mp
}

// CppName returns C++ variable name with reserved keywords suffixed by an underscore
func (mp *fbsField) CppName() string {
	return cppName(mp.Name)
}

// CppType returns C++ type name
func (mp *fbsField) CppType() string {
	var fbsType = mp.fbsField.Type(nil)
	var baseType = fbsType.BaseType()
	var cppType = fbsTypeToCppType[baseType]
	if baseType == reflection.BaseTypeVector {
		cppType = cppType + "<" + fbsTypeToCppType[fbsType.Element()] + ">"
	}
	return cppType
}

// FbIsVector returns true if the property is considered a vector type.
func (mp *fbsField) FbIsVector() bool {
	switch mp.ModelProperty.Type {
	case model.PropertyTypeString:
		return true
	case model.PropertyTypeByteVector:
		return true
	case model.PropertyTypeStringVector:
		return true
	}
	return false
}

// CElementType returns C vector element type name
func (mp *fbsField) CElementType() string {
	switch mp.ModelProperty.Type {
	case model.PropertyTypeByteVector:
		return fbsTypeToCppType[mp.fbsField.Type(nil).Element()]
	case model.PropertyTypeString:
		return "char"
	case model.PropertyTypeStringVector:
		return "char*"
	}
	return ""
}

// FlatccFnPrefix returns the field's type as used in Flatcc.
func (mp *fbsField) FlatccFnPrefix() string {
	return fbsTypeToFlatccFnPrefix[mp.fbsField.Type(nil).BaseType()]
}

// FbTypeSize returns the field's type flatbuffers size.
func (mp *fbsField) FbTypeSize() uint8 {
	return fbsTypeSize[mp.fbsField.Type(nil).BaseType()]
}

// FbOffsetFactory returns an offset factory used to build flatbuffers if this property is a complex type.
// See also FbOffsetType().
func (mp *fbsField) FbOffsetFactory() string {
	switch mp.ModelProperty.Type {
	case model.PropertyTypeString:
		return "CreateString"
	case model.PropertyTypeByteVector:
		return "CreateVector"
	case model.PropertyTypeStringVector:
		return "CreateVectorOfStrings"
	}
	return ""
}

// FbOffsetType returns a type used to read flatbuffers if this property is a complex type.
// See also FbOffsetFactory().
func (mp *fbsField) FbOffsetType() string {
	switch mp.ModelProperty.Type {
	case model.PropertyTypeString:
		return "flatbuffers::Vector<char>"
	case model.PropertyTypeByteVector:
		return "flatbuffers::Vector<" + fbsTypeToCppType[mp.fbsField.Type(nil).Element()] + ">"
	case model.PropertyTypeStringVector:
		return "" // NOTE custom handling in the template
	}
	return ""
}

// FbDefaultValue returns a default value for scalars
func (mp *fbsField) FbDefaultValue() string {
	switch mp.ModelProperty.Type {
	case model.PropertyTypeFloat:
		return "0.0f"
	case model.PropertyTypeDouble:
		return "0.0"
	}
	return "0"
}
