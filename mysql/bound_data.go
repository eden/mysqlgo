// Copyright 2009 Eden Li. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// BoundData - represents a result or parameter bind.
package mysql

import (
	"unsafe";
	"encoding/binary";
)

type BoundData struct {
	buffer	[]byte;
	blen	int;
	is_null	[1]byte;
	error	[1]byte;
	myType	MysqlType;
}

func NewBoundData(t MysqlType, buf []byte, n int) (data *BoundData) {
	data = &BoundData{
		is_null: [1]byte{0},
		error: [1]byte{0},
		myType: t,
		buffer: buf,
		blen: correctSize(t, n)
	};
	if buf == nil && n > 0 {
		data.buffer = make([]byte, data.blen);
	}
	return;
}

func correctSize(t MysqlType, n int) (sz int) {
	switch(t) {
	case MysqlTypeNewdecimal: fallthrough;
	case MysqlTypeDecimal:
		sz = unsafe.Sizeof(float64(0)) // TODO test this case...

	case MysqlTypeTiny:
		sz = unsafe.Sizeof(int8(0))

	case MysqlTypeShort:
		sz = unsafe.Sizeof(int16(0))

	case MysqlTypeInt24: fallthrough
	case MysqlTypeLong:
		sz = unsafe.Sizeof(int32(0))

	case MysqlTypeLonglong:
		sz = unsafe.Sizeof(int64(0))

	case MysqlTypeFloat:
		sz = unsafe.Sizeof(float(0))

	case MysqlTypeDouble:
		sz = unsafe.Sizeof(float64(0))

	// TODO
	case MysqlTypeDate: fallthrough;
	case MysqlTypeTime: fallthrough;
	case MysqlTypeDatetime: fallthrough;
	case MysqlTypeYear: fallthrough;
	case MysqlTypeNewdate: fallthrough;
	default:
		sz = n
	}
	return
}

func bytesForUnsafePointer(buf unsafe.Pointer, n int) (b []byte) {
	b = make([]byte, n);
	for i := range(b) {
		b[i] = *((*byte)(unsafe.Pointer(uintptr(buf) + uintptr(i))))
	}
	return;
}

func (d *BoundData) Value() (v interface {}, ok bool) {
	if d.is_null[0] == 1 {
		return nil, true
	}

	// TODO remove this once it becomes clear why d.buffer[0] segfaults
	bytes := bytesForUnsafePointer(unsafe.Pointer(&d.buffer), d.blen);
	ok = false;

	switch(d.myType) {
	case MysqlTypeLong:
		if len(bytes) == 4 { v, ok = binary.LittleEndian.Uint32(bytes), true }

	case MysqlTypeVarchar: fallthrough;
	case MysqlTypeString:
		v = string(bytes);
		ok = true

	case MysqlTypeTiny:
		if len(bytes) == 1 { v, ok = bytes[0], true }

	case MysqlTypeShort:
		if len(bytes) == 2 { v, ok = binary.LittleEndian.Uint16(bytes), true }

	case MysqlTypeLonglong:
		if len(bytes) == 8 { v, ok = binary.LittleEndian.Uint64(bytes), true }

	case MysqlTypeNewdecimal: fallthrough;
	case MysqlTypeDecimal: fallthrough;
		//v, ok = i.(float64)

	case MysqlTypeFloat: fallthrough;
		//v, ok = i.(float)

	case MysqlTypeDouble: fallthrough;
		//v, ok = i.(float64)

	case MysqlTypeTinyBlob: fallthrough;
	case MysqlTypeMedium_Blob: fallthrough;
	case MysqlTypeLongBlob: fallthrough;
	case MysqlTypeBlob: fallthrough;

	case MysqlTypeNull: fallthrough;
	case MysqlTypeTimestamp: fallthrough;
	case MysqlTypeInt24: fallthrough;
	case MysqlTypeDate: fallthrough;
	case MysqlTypeTime: fallthrough;
	case MysqlTypeDatetime: fallthrough;
	case MysqlTypeYear: fallthrough;
	case MysqlTypeNewdate: fallthrough;
	case MysqlTypeBit: fallthrough;
	case MysqlTypeEnum: fallthrough;
	case MysqlTypeSet: fallthrough;
	case MysqlTypeVarString: fallthrough;
	case MysqlTypeGeometry:
		v = bytes;
		ok = true
	}

	return;
}
