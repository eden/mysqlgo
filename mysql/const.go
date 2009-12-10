// Copyright 2009 Eden Li. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Common mysql constants and conversion functions
package mysql

type MysqlType int

const (
	MysqlTypeDecimal	= iota;
	MysqlTypeTiny;
	MysqlTypeShort;
	MysqlTypeLong;
	MysqlTypeFloat;
	MysqlTypeDouble;
	MysqlTypeNull;
	MysqlTypeTimestamp;
	MysqlTypeLonglong;
	MysqlTypeInt24;
	MysqlTypeDate;
	MysqlTypeTime;
	MysqlTypeDatetime;
	MysqlTypeYear;
	MysqlTypeNewdate;
	MysqlTypeVarchar;
	MysqlTypeBit;
	MysqlTypeNewdecimal	= 246;
	MysqlTypeEnum		= 247;
	MysqlTypeSet		= 248;
	MysqlTypeTinyBlob	= 249;
	MysqlTypeMedium_Blob	= 250;
	MysqlTypeLongBlob	= 251;
	MysqlTypeBlob		= 252;
	MysqlTypeVarString	= 253;
	MysqlTypeString		= 254;
	MysqlTypeGeometry	= 255;
)
