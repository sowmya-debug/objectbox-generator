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

package binding

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/objectbox/objectbox-generator/internal/generator/model"
)

// Object holds common entity information used by specialized code parsers/generators.
// Additionally, it groups some shared logic, e.g. annotation processing
type Object struct {
	ModelEntity *model.Entity
	Name        string
	Namespace   string
	IsSkipped   bool
}

func CreateObject(entity *model.Entity) *Object {
	return &Object{ModelEntity: entity}
}

func (object *Object) SetName(name string) {
	// look for namespace separators
	var lastDot = strings.LastIndex(name, ".")
	if lastDot > 0 {
		object.Namespace = name[:lastDot]
		name = name[lastDot+1:]
	}

	object.Name = name
	if len(object.ModelEntity.Name) == 0 {
		object.ModelEntity.Name = name
	}
}

// ProcessAnnotations checks all set annotations for any inconsistencies and sets local/entity properties (uid, name, ...)
// TODO move generator.Annotation to this package
func (object *Object) ProcessAnnotations(a map[string]*Annotation) error {
	for _, alternative := range []string{"-", "transient"} {
		if a[alternative] != nil {
			if len(a) != 1 || a[alternative].Value != "" {
				return errors.New("to ignore the entity, use only `objectbox:\"" + alternative + "\"` as an annotation")
			}
			object.IsSkipped = true
			return nil
		}
	}

	if a["name"] != nil {
		if len(a["name"].Value) == 0 {
			return fmt.Errorf("name annotation value must not be empty - it's the entity name in DB")
		}
		object.ModelEntity.Name = a["name"].Value
	}

	if a["uid"] != nil {
		if len(a["uid"].Value) == 0 {
			// in case the user doesn't provide `objectbox:"uid"` value, it's considered in-process of setting up UID
			// this flag is handled by the merge mechanism and prints the UID of the already existing entity
			object.ModelEntity.UidRequest = true
		} else if uid, err := strconv.ParseUint(a["uid"].Value, 10, 64); err != nil {
			return fmt.Errorf("can't parse uid - %s", err)
		} else if id, err := object.ModelEntity.Id.GetIdAllowZero(); err != nil {
			return fmt.Errorf("can't parse entity Id - %s", err)
		} else {
			object.ModelEntity.Id = model.CreateIdUid(id, uid)
		}
	}

	return nil
}
