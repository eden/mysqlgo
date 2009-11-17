// Copyright 2009 Eden Li. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Wrapper library to hide mysql structures from cgo since it can't yet
// properly parse it.  All mysql typedefs are hidden behind void pointers.  Not
// the safest thing, but at least it gets us by until cgo is fixed.  See the
// README for more information.

#include "mw.h"
#include <stdlib.h>
#include <stdio.h>
#include <mysql.h>

void mw_library_init() {
    mysql_library_init(0, NULL, NULL);
}

mw mw_init(mw h) {
    return (mw)mysql_init((MYSQL *)h);
}

const char *mw_error(mw h) {
    return mysql_error((MYSQL *)h);
}

mw mw_real_connect(
    mw h, const char *host, const char *uname,
    const char *passwd, const char *db, int port)
{
    return (mw)mysql_real_connect(
		(MYSQL *)h,
        host, uname, passwd, db, port, NULL, 0);
}

void mw_close(mw h) {
    mysql_close((MYSQL *)h);
}

void mw_free_result(mwres res) {
    mysql_free_result((MYSQL_RES *)res);
}

int mw_query(mw h, const char *q) {
	return mysql_query((MYSQL *)h, q);
}

mwres mw_store_result(mw h) {
    return (mwres)mysql_store_result((MYSQL *)h);
}

char *mw_row(mwrow row, int i) {
    return (char *)((MYSQL_ROW)row)[i];
}

const char *mw_field_name_at(mwfield field, int i) {
    return ((MYSQL_FIELD *)field)[i].name;
}

int mw_field_type_at(mwfield field, int i) {
    return ((MYSQL_FIELD *)field)[i].type;
}

int mw_field_count(mw h) {
    return mysql_field_count((MYSQL *)h);
}

int mw_num_fields(mwres res) {
    return mysql_num_fields((MYSQL_RES *)res);
}

mwfield mw_fetch_fields(mwres res) {
	return (mwfield)mysql_fetch_fields((MYSQL_RES *)res);
}

mwrow mw_fetch_row(mwres res) {
    return mysql_fetch_row((MYSQL_RES *)res);
}

unsigned long long mw_num_rows(mwres res) {
    return mysql_num_rows((MYSQL_RES *)res);
}

void mw_thread_init(void) {
	(void)mysql_thread_init();
}

void mw_thread_end(void) {
	mysql_thread_end();
}
