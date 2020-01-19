package db

import (
	"database/sql"
	"reflect"
	"strings"
	"time"
)

//DbResultSetMetaData DbResultSet meta data container
type DbResultSetMetaData struct {
	cols     []string
	colTypes []*ColumnType
}

//Columns column name(s)
func (rsetmeta *DbResultSetMetaData) Columns() []string {
	return rsetmeta.cols
}

//ColumnTypes ColumnType(s) definition(s)
func (rsetmeta *DbResultSetMetaData) ColumnTypes() []*ColumnType {
	return rsetmeta.colTypes
}

//ColumnType structure defining column definition
type ColumnType struct {
	name string

	hasNullable       bool
	hasLength         bool
	hasPrecisionScale bool

	nullable     bool
	length       int64
	databaseType string
	precision    int64
	scale        int64
	scanType     reflect.Type
}

//Name ColumnType.Name()
func (colType *ColumnType) Name() string {
	return colType.name
}

//Numeric ColumnType is Numeric() bool
func (colType *ColumnType) Numeric() bool {
	if colType.hasPrecisionScale {
		return true
	}
	return strings.Index(colType.databaseType, "CHAR") == -1 && strings.Index(colType.databaseType, "DATE") == -1 && strings.Index(colType.databaseType, "TIME") == -1
}

//HasNullable ColumnType content has NULL able content
func (colType *ColumnType) HasNullable() bool {
	return colType.hasNullable
}

//HasLength ColumnType content has Length definition
func (colType *ColumnType) HasLength() bool {
	return colType.hasLength
}

//HasPrecisionScale ColumnType content has PrecisionScale
func (colType *ColumnType) HasPrecisionScale() bool {
	return colType.hasPrecisionScale
}

//Nullable ColumnType content is Nullable
func (colType *ColumnType) Nullable() bool {
	return colType.nullable
}

//Length ColumnType content lenth must be used in conjunction with HasLength
func (colType *ColumnType) Length() int64 {
	return colType.length
}

//DatabaseType ColumnType underlying db type as defined by cnstring of DbConnection
func (colType *ColumnType) DatabaseType() string {
	return colType.databaseType
}

//Precision ColumnType numeric Precision. Used in conjunction with HasPrecisionScale
func (colType *ColumnType) Precision() int64 {
	return colType.precision
}

//Scale ColumnType Scale. Used in conjunction with HasPrecisionScale
func (colType *ColumnType) Scale() int64 {
	return colType.scale
}

//Type ColumnType reflect.Type as specified by golang sql/database
func (colType *ColumnType) Type() reflect.Type {
	return colType.scanType
}

//DbResultSet DbResultSet container
type DbResultSet struct {
	rsmetadata  *DbResultSetMetaData
	stmnt       *DbStatement
	rset        *sql.Rows
	data        []interface{}
	dispdata    []interface{}
	dataref     []interface{}
	err         error
	dosomething chan bool
}

//MetaData DbResultSet=>DbResultSetMetaData
func (rset *DbResultSet) MetaData() *DbResultSetMetaData {
	return rset.rsmetadata
}

//Data return Displayable data in the form of a slice, 'array', of interface{} values
func (rset *DbResultSet) Data() []interface{} {
	go func(somethingDone chan bool) {
		defer func(){
			somethingDone <- true
		}()
		for n := range rset.data {
			coltype := rset.rsmetadata.colTypes[n]
			rset.dispdata[n] = castSQLTypeValue(rset.data[n], coltype)
		}
	}(rset.dosomething)
	<-rset.dosomething
	return rset.dispdata
}

func castSQLTypeValue(valToCast interface{}, colType *ColumnType) (castedVal interface{}) {
	if valToCast != nil {
		if d, dok := valToCast.([]uint8); dok {
			castedVal = string(d)
		} else if sd, dok := valToCast.(string); dok {
			castedVal = sd
		} else if dtime, dok := valToCast.(time.Time); dok {
			castedVal = dtime.Format("2006-01-02T15:04:05")
		} else {
			castedVal = valToCast
		}
	} else {
		castedVal = valToCast
	}
	return castedVal
}

//Next return true if able to move focus of DbResultSet to the next underlying record
// or false if the end is reached
func (rset *DbResultSet) Next() (next bool, err error) {
	if next = rset.rset.Next(); next {
		if rset.data == nil {
			rset.data = make([]interface{}, len(rset.rsmetadata.cols))
			rset.dataref = make([]interface{}, len(rset.rsmetadata.cols))
			rset.dispdata = make([]interface{}, len(rset.rsmetadata.cols))
		}
		go func(somthingDone chan bool) { 
			defer func(){
				somethingDone<-true
			}()
			for n := range rset.data {
				rset.dataref[n] = &rset.data[n]
			}
			if scerr := rset.rset.Scan(rset.dataref...); scerr != nil {
				rset.Close()
				err = scerr
				next = false
			}
		}(rset.dosomething)
		<-rset.dosomething
	} else {
		if rseterr := rset.rset.Err(); rseterr != nil {
			err = rseterr
		}
		rset.Close()
	}
	return next, err
}

//Close the DbResultSet as well as the underlying DbStatement related to this DbResultSet
//After this action the DbResultSet is 'empty' or cleaned up in a golang world
func (rset *DbResultSet) Close() (err error) {
	if rset.data != nil {
		rset.data = nil
	}
	if rset.dataref != nil {
		rset.dataref = nil
	}
	if rset.dispdata != nil {
		rset.dispdata = nil
	}
	if rset.rsmetadata != nil {
		rset.rsmetadata.colTypes = nil
		rset.rsmetadata.cols = nil
	}
	if rset.stmnt != nil {
		err = rset.stmnt.Close()
		rset.stmnt = nil
	}
	if rset.dosomething != nil {
		close(rset.dosomething)
		rset.dosomething = nil
	}
	return
}
