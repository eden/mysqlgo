#ifndef __mw_h
#define __mw_h

typedef void *mw;
typedef void *mwres;
typedef void *mwrow;
typedef void *mwfield;

void mw_library_init(void);
mw mw_init(mw h);
mw mw_real_connect(
    mw h, const char *host, const char *uname,
    const char *passwd, const char *db, int port);
const char *mw_error(mw h);
void mw_close(mw h);
void mw_free_result(mwres res);
int mw_query(mw h, const char *q);
mwres mw_store_result(mw h);
char *mw_row(mwrow row, int i);
const char *mw_field_name_at(mwfield field, int i);
int mw_field_type_at(mwfield field, int i);
int mw_field_count(mw h);
int mw_num_fields(mwres res);
mwfield mw_fetch_fields(mwres res);
mwrow mw_fetch_row(mwres res);
unsigned long long mw_num_rows(mwres res);

void mw_thread_init(void);
void mw_thread_end(void);

#endif // __mw_h
