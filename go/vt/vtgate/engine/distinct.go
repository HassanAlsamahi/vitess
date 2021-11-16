/*
Copyright 2020 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package engine

import (
	"vitess.io/vitess/go/mysql/collations"
	"vitess.io/vitess/go/sqltypes"
	querypb "vitess.io/vitess/go/vt/proto/query"
	"vitess.io/vitess/go/vt/vtgate/evalengine"
)

// Distinct Primitive is used to uniqueify results
var _ Primitive = (*Distinct)(nil)

// Distinct Primitive is used to uniqueify results
type Distinct struct {
	Source        Primitive
	ColCollations []collations.ID
}

type row = []sqltypes.Value

type probeTable struct {
	seenRows      map[evalengine.HashCode][]row
	colCollations []collations.ID
}

func (pt *probeTable) exists(inputRow row) (bool, error) {
	// calculate hashcode from all column values in the input row
	code := evalengine.HashCode(17)
	for idx, value := range inputRow {
		// We use unknown collations when we do not have collation information
		// This is safe for types which do not require collation information like
		// numeric types. It will fail at runtime for text types.
		collation := collations.Unknown
		if len(pt.colCollations) > idx {
			collation = pt.colCollations[idx]
		}
		hashcode, err := evalengine.NullsafeHashcode(value, collation, value.Type())
		if err != nil {
			return false, err
		}
		code = code*31 + hashcode
	}

	existingRows, found := pt.seenRows[code]
	if !found {
		// nothing with this hash code found, we can be sure it's a not seen row
		pt.seenRows[code] = []row{inputRow}
		return false, nil
	}

	// we found something in the map - still need to check all individual values
	// so we don't just fall for a hash collision
	for _, existingRow := range existingRows {
		exists, err := equal(existingRow, inputRow, pt.colCollations)
		if err != nil {
			return false, err
		}
		if exists {
			return true, nil
		}
	}

	pt.seenRows[code] = append(existingRows, inputRow)

	return false, nil
}

func equal(a, b []sqltypes.Value, colCollations []collations.ID) (bool, error) {
	for i, aVal := range a {
		collation := collations.Unknown
		if len(colCollations) > i {
			collation = colCollations[i]
		}
		cmp, err := evalengine.NullsafeCompare(aVal, b[i], collation)
		if err != nil {
			return false, err
		}
		if cmp != 0 {
			return false, nil
		}
	}
	return true, nil
}

func newProbeTable(colCollations []collations.ID) *probeTable {
	return &probeTable{
		seenRows:      map[uintptr][]row{},
		colCollations: colCollations,
	}
}

// TryExecute implements the Primitive interface
func (d *Distinct) TryExecute(vcursor VCursor, bindVars map[string]*querypb.BindVariable, wantfields bool) (*sqltypes.Result, error) {
	input, err := vcursor.ExecutePrimitive(d.Source, bindVars, wantfields)
	if err != nil {
		return nil, err
	}

	result := &sqltypes.Result{
		Fields:   input.Fields,
		InsertID: input.InsertID,
	}

	pt := newProbeTable(d.ColCollations)

	for _, row := range input.Rows {
		exists, err := pt.exists(row)
		if err != nil {
			return nil, err
		}
		if !exists {
			result.Rows = append(result.Rows, row)
		}
	}

	return result, err
}

// TryStreamExecute implements the Primitive interface
func (d *Distinct) TryStreamExecute(vcursor VCursor, bindVars map[string]*querypb.BindVariable, wantfields bool, callback func(*sqltypes.Result) error) error {
	pt := newProbeTable(d.ColCollations)

	err := vcursor.StreamExecutePrimitive(d.Source, bindVars, wantfields, func(input *sqltypes.Result) error {
		result := &sqltypes.Result{
			Fields:   input.Fields,
			InsertID: input.InsertID,
		}
		for _, row := range input.Rows {
			exists, err := pt.exists(row)
			if err != nil {
				return err
			}
			if !exists {
				result.Rows = append(result.Rows, row)
			}
		}
		return callback(result)
	})

	return err
}

// RouteType implements the Primitive interface
func (d *Distinct) RouteType() string {
	return d.Source.RouteType()
}

// GetKeyspaceName implements the Primitive interface
func (d *Distinct) GetKeyspaceName() string {
	return d.Source.GetKeyspaceName()
}

// GetTableName implements the Primitive interface
func (d *Distinct) GetTableName() string {
	return d.Source.GetTableName()
}

// GetFields implements the Primitive interface
func (d *Distinct) GetFields(vcursor VCursor, bindVars map[string]*querypb.BindVariable) (*sqltypes.Result, error) {
	return d.Source.GetFields(vcursor, bindVars)
}

// NeedsTransaction implements the Primitive interface
func (d *Distinct) NeedsTransaction() bool {
	return d.Source.NeedsTransaction()
}

// Inputs implements the Primitive interface
func (d *Distinct) Inputs() []Primitive {
	return []Primitive{d.Source}
}

func (d *Distinct) description() PrimitiveDescription {
	var other map[string]interface{}
	if d.ColCollations != nil {
		allUnknown := true
		other = map[string]interface{}{}
		var colls []string
		for _, collation := range d.ColCollations {
			coll := collations.Default().LookupByID(collation)
			if coll == nil {
				colls = append(colls, "UNKNOWN")
			} else {
				colls = append(colls, coll.Name())
				allUnknown = false
			}
		}
		if !allUnknown {
			other["Collations"] = colls
		}
	}
	return PrimitiveDescription{
		Other:        other,
		OperatorType: "Distinct",
	}
}
