/*
Copyright 2021 The Vitess Authors.

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
// Code generated by Sizegen. DO NOT EDIT.

package evalengine

import hack "vitess.io/vitess/go/hack"

type cachedObject interface {
	CachedSize(alloc bool) int64
}

func (cached *ArithmeticExpr) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field BinaryExpr vitess.io/vitess/go/vt/vtgate/evalengine.BinaryExpr
	size += cached.BinaryExpr.CachedSize(false)
	// field Op vitess.io/vitess/go/vt/vtgate/evalengine.opArith
	if cc, ok := cached.Op.(cachedObject); ok {
		size += cc.CachedSize(true)
	}
	return size
}
func (cached *BinaryExpr) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(32)
	}
	// field Left vitess.io/vitess/go/vt/vtgate/evalengine.Expr
	if cc, ok := cached.Left.(cachedObject); ok {
		size += cc.CachedSize(true)
	}
	// field Right vitess.io/vitess/go/vt/vtgate/evalengine.Expr
	if cc, ok := cached.Right.(cachedObject); ok {
		size += cc.CachedSize(true)
	}
	return size
}
func (cached *BindVariable) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(32)
	}
	// field Key string
	size += hack.RuntimeAllocSize(int64(len(cached.Key)))
	return size
}
func (cached *BitwiseExpr) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field BinaryExpr vitess.io/vitess/go/vt/vtgate/evalengine.BinaryExpr
	size += cached.BinaryExpr.CachedSize(false)
	// field Op vitess.io/vitess/go/vt/vtgate/evalengine.opBit
	if cc, ok := cached.Op.(cachedObject); ok {
		size += cc.CachedSize(true)
	}
	return size
}
func (cached *BitwiseNotExpr) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(16)
	}
	// field UnaryExpr vitess.io/vitess/go/vt/vtgate/evalengine.UnaryExpr
	size += cached.UnaryExpr.CachedSize(false)
	return size
}
func (cached *CallExpr) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field Arguments vitess.io/vitess/go/vt/vtgate/evalengine.TupleExpr
	{
		size += hack.RuntimeAllocSize(int64(cap(cached.Arguments)) * int64(16))
		for _, elem := range cached.Arguments {
			if cc, ok := elem.(cachedObject); ok {
				size += cc.CachedSize(true)
			}
		}
	}
	// field Method string
	size += hack.RuntimeAllocSize(int64(len(cached.Method)))
	return size
}
func (cached *CaseExpr) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field cases []vitess.io/vitess/go/vt/vtgate/evalengine.WhenThen
	{
		size += hack.RuntimeAllocSize(int64(cap(cached.cases)) * int64(32))
		for _, elem := range cached.cases {
			size += elem.CachedSize(false)
		}
	}
	// field Else vitess.io/vitess/go/vt/vtgate/evalengine.Expr
	if cc, ok := cached.Else.(cachedObject); ok {
		size += cc.CachedSize(true)
	}
	return size
}
func (cached *CollateExpr) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(24)
	}
	// field UnaryExpr vitess.io/vitess/go/vt/vtgate/evalengine.UnaryExpr
	size += cached.UnaryExpr.CachedSize(false)
	return size
}
func (cached *Column) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(24)
	}
	return size
}
func (cached *ComparisonExpr) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field BinaryExpr vitess.io/vitess/go/vt/vtgate/evalengine.BinaryExpr
	size += cached.BinaryExpr.CachedSize(false)
	// field Op vitess.io/vitess/go/vt/vtgate/evalengine.ComparisonOp
	if cc, ok := cached.Op.(cachedObject); ok {
		size += cc.CachedSize(true)
	}
	return size
}
func (cached *CompiledExpr) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(64)
	}
	// field code []vitess.io/vitess/go/vt/vtgate/evalengine.frame
	{
		size += hack.RuntimeAllocSize(int64(cap(cached.code)) * int64(8))
	}
	// field original vitess.io/vitess/go/vt/vtgate/evalengine.Expr
	if cc, ok := cached.original.(cachedObject); ok {
		size += cc.CachedSize(true)
	}
	return size
}
func (cached *ConvertExpr) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(64)
	}
	// field UnaryExpr vitess.io/vitess/go/vt/vtgate/evalengine.UnaryExpr
	size += cached.UnaryExpr.CachedSize(false)
	// field Type string
	size += hack.RuntimeAllocSize(int64(len(cached.Type)))
	return size
}
func (cached *ConvertUsingExpr) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(24)
	}
	// field UnaryExpr vitess.io/vitess/go/vt/vtgate/evalengine.UnaryExpr
	size += cached.UnaryExpr.CachedSize(false)
	return size
}
func (cached *InExpr) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field BinaryExpr vitess.io/vitess/go/vt/vtgate/evalengine.BinaryExpr
	size += cached.BinaryExpr.CachedSize(false)
	return size
}
func (cached *IsExpr) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(32)
	}
	// field UnaryExpr vitess.io/vitess/go/vt/vtgate/evalengine.UnaryExpr
	size += cached.UnaryExpr.CachedSize(false)
	return size
}
func (cached *LikeExpr) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(64)
	}
	// field BinaryExpr vitess.io/vitess/go/vt/vtgate/evalengine.BinaryExpr
	size += cached.BinaryExpr.CachedSize(false)
	// field Match vitess.io/vitess/go/mysql/collations.WildcardPattern
	if cc, ok := cached.Match.(cachedObject); ok {
		size += cc.CachedSize(true)
	}
	return size
}
func (cached *Literal) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(16)
	}
	// field inner vitess.io/vitess/go/vt/vtgate/evalengine.eval
	if cc, ok := cached.inner.(cachedObject); ok {
		size += cc.CachedSize(true)
	}
	return size
}
func (cached *LogicalExpr) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(64)
	}
	// field BinaryExpr vitess.io/vitess/go/vt/vtgate/evalengine.BinaryExpr
	size += cached.BinaryExpr.CachedSize(false)
	// field opname string
	size += hack.RuntimeAllocSize(int64(len(cached.opname)))
	return size
}
func (cached *NegateExpr) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(16)
	}
	// field UnaryExpr vitess.io/vitess/go/vt/vtgate/evalengine.UnaryExpr
	size += cached.UnaryExpr.CachedSize(false)
	return size
}
func (cached *NotExpr) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(16)
	}
	// field UnaryExpr vitess.io/vitess/go/vt/vtgate/evalengine.UnaryExpr
	size += cached.UnaryExpr.CachedSize(false)
	return size
}
func (cached *UnaryExpr) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(16)
	}
	// field Inner vitess.io/vitess/go/vt/vtgate/evalengine.Expr
	if cc, ok := cached.Inner.(cachedObject); ok {
		size += cc.CachedSize(true)
	}
	return size
}
func (cached *WhenThen) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(32)
	}
	// field when vitess.io/vitess/go/vt/vtgate/evalengine.Expr
	if cc, ok := cached.when.(cachedObject); ok {
		size += cc.CachedSize(true)
	}
	// field then vitess.io/vitess/go/vt/vtgate/evalengine.Expr
	if cc, ok := cached.then.(cachedObject); ok {
		size += cc.CachedSize(true)
	}
	return size
}
func (cached *builtinASCII) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinAbs) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinAcos) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinAsin) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinAtan) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinAtan2) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinBitCount) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinBitLength) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinCeil) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinChangeCase) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinCharLength) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinCoalesce) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinCollation) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinConv) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinConvertTz) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinCos) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinCot) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinCrc32) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinCurdate) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinDatabase) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinDateFormat) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinDegrees) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinExp) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinFloor) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinFromBase64) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinHex) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinJSONArray) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinJSONContainsPath) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinJSONDepth) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinJSONExtract) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinJSONKeys) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinJSONLength) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinJSONObject) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinJSONUnquote) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinLength) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinLn) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinLog) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinLog10) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinLog2) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinMD5) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinMultiComparison) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinNow) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinPi) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinPow) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinRadians) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinRandomBytes) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinRepeat) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinRound) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinSHA1) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinSHA2) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinSign) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinSin) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinSqrt) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinSysdate) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinTan) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinToBase64) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinTruncate) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinUser) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinUtcDate) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinVersion) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field CallExpr vitess.io/vitess/go/vt/vtgate/evalengine.CallExpr
	size += cached.CallExpr.CachedSize(false)
	return size
}
func (cached *builtinWeightString) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(48)
	}
	// field String vitess.io/vitess/go/vt/vtgate/evalengine.Expr
	if cc, ok := cached.String.(cachedObject); ok {
		size += cc.CachedSize(true)
	}
	// field Cast string
	size += hack.RuntimeAllocSize(int64(len(cached.Cast)))
	return size
}
func (cached *evalBytes) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(32)
	}
	// field bytes []byte
	{
		size += hack.RuntimeAllocSize(int64(cap(cached.bytes)))
	}
	return size
}
func (cached *evalDecimal) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(16)
	}
	// field dec vitess.io/vitess/go/mysql/decimal.Decimal
	size += cached.dec.CachedSize(false)
	return size
}
func (cached *evalFloat) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(8)
	}
	return size
}
func (cached *evalInt64) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(8)
	}
	return size
}
func (cached *evalTemporal) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(16)
	}
	return size
}
func (cached *evalTuple) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(24)
	}
	// field t []vitess.io/vitess/go/vt/vtgate/evalengine.eval
	{
		size += hack.RuntimeAllocSize(int64(cap(cached.t)) * int64(16))
		for _, elem := range cached.t {
			if cc, ok := elem.(cachedObject); ok {
				size += cc.CachedSize(true)
			}
		}
	}
	return size
}
func (cached *evalUint64) CachedSize(alloc bool) int64 {
	if cached == nil {
		return int64(0)
	}
	size := int64(0)
	if alloc {
		size += int64(16)
	}
	return size
}
