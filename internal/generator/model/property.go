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

package model

import "fmt"

// Property in a model
type Property struct {
	Id             IdUid        `json:"id"`
	Name           string       `json:"name"`
	IndexId        *IdUid       `json:"indexId,omitempty"` // a pointer because it may be nil
	Type           int          `json:"type"`
	Flags          int          `json:"flags,omitempty"`
	RelationTarget string       `json:"relationTarget,omitempty"`
	Entity         *Entity      `json:"-"`
	UidRequest     bool         `json:"-"` // TODO
	Path           string       `json:"-"` // TODO
	Meta           PropertyMeta `json:"-"`
}

// CreateProperty creates a property
func CreateProperty(entity *Entity, id Id, uid Uid) *Property {
	return &Property{
		Entity: entity,
		Id:     CreateIdUid(id, uid),
	}
}

// Validate performs initial validation of loaded data so that it doesn't have to be checked in each function
func (property *Property) Validate() error {
	if property.Entity == nil {
		return fmt.Errorf("undefined parent entity")
	}

	if err := property.Id.Validate(); err != nil {
		return err
	}

	if property.IndexId != nil {
		if err := property.IndexId.Validate(); err != nil {
			return fmt.Errorf("indexId: %s", err)
		}
	}

	if len(property.Name) == 0 {
		return fmt.Errorf("name is undefined")
	}

	// NOTE type can't be validated because entities are update one-by-one and so
	// on the second one, validate() during load would failonly check this
	// if property.Type == 0 {
	//	return fmt.Errorf("type is undefined")
	// }

	// IDs must not be tagged unsigned for compatibility reasons
	if property.isIdProperty() {
		if !property.hasValidTypeAsId() {
			return fmt.Errorf("invalid type on property marked as ID: %d", property.Type)
		}
	}

	return nil
}

func (property *Property) finalize() error {
	if property.isIdProperty() {
		// IDs must not be tagged unsigned for compatibility reasons
		property.Flags = property.Flags & ^PropertyFlagUnsigned

		// always stored as Long
		property.Type = PropertyTypeLong
	}

	return property.Validate()
}

func (property *Property) isIdProperty() bool {
	return property.Flags&PropertyFlagId != 0
}

func (property *Property) hasValidTypeAsId() bool {
	return property.Type == PropertyTypeLong
}

// CreateIndex creates an index
func (property *Property) CreateIndex() error {
	if property.IndexId != nil {
		return fmt.Errorf("can't create an index - it already exists")
	}

	indexId, err := property.Entity.model.createIndexId()
	if err != nil {
		return err
	}

	property.IndexId = &indexId
	return nil
}

// RemoveIndex removes an index
func (property *Property) RemoveIndex() error {
	if property.IndexId == nil {
		return fmt.Errorf("can't remove index - it's not defined")
	}

	property.Entity.model.RetiredIndexUids = append(property.Entity.model.RetiredIndexUids, property.IndexId.getUidSafe())

	property.IndexId = nil

	return nil
}

// containsUid recursively checks whether given Uid is present in the model
func (property *Property) containsUid(searched Uid) bool {
	if property.Id.getUidSafe() == searched {
		return true
	}

	if property.IndexId != nil && property.IndexId.getUidSafe() == searched {
		return true
	}

	return false
}