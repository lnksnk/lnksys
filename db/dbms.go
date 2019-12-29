package db

import (
	iorw "../iorw"
	active "../iorw/active"
	"io"
	"sync"
	/**/
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/stdlib"
	_ "github.com/lib/pq"
	/* _ "gopkg.in/goracle.v2" */)

type DbmsManager struct {
	dbcns map[string]*DbConnection
}

var dbmsmngr *DbmsManager

func DBMSManager() *DbmsManager {
	if dbmsmngr == nil {
		dbmsmngr = &DbmsManager{}
	}
	return dbmsmngr
}

func (dbmsmngr *DbmsManager) Dbms(alias string) (dbcn *DbConnection) {
	if dbmsmngr.dbcns != nil {
		dbcn = dbmsmngr.dbcns[alias]
	}
	return
}

func (dbmsmngr *DbmsManager) RegisterDbms(alias string, cnsettings ...string) {
	if dbmsmngr.dbcns == nil || len(dbmsmngr.dbcns) == 0 {
		if dbmsmngr.dbcns == nil {
			dbmsmngr.dbcns = map[string]*DbConnection{}
		}
	}

	if alias != "" {
		if _, aliasok := dbmsmngr.dbcns[alias]; aliasok {
			dbmsmngr.dbcns[alias].LoadCnSettings(cnsettings...)
		} else {
			var dbcn = NewDbConnection(cnsettings...)
			dbmsmngr.dbcns[alias] = dbcn
			dbcn.dbms = dbmsmngr
		}
	}

}

func init() {
	if dbmsmngr == nil {
		dbmsmngr = &DbmsManager{}
	}

	active.MapGlobals("DBQuery", func(alias string, query string, args ...interface{}) (dbquery *DBQuery) {
		dbquery = DBMSManager().Query(alias, query, args...)
		return
	}, "DBExecute", func(alias string, query string, args ...interface{}) (dbexecuted *DBExecuted) {
		dbexecuted = DBMSManager().Execute(alias, query, args...)
		return
	}, "DBMSRegister", func(alias string, cnsettings ...string) {
		DBMSManager().RegisterDbms(alias, cnsettings...)
		return
	}, "DBMS", func(alias string) (dbcn *DbConnection) {
		dbcn = DBMSManager().Dbms(alias)
		return
	})
}

//DBExecuted controller
type DBExecuted struct {
	LastInsertId int64
	RowsAffected int64
	Err          error
}

//Execute execute query for alias connection
//return a DbExecute controller that represents the outcome of the executed request
func (dbmngr *DbmsManager) Execute(alias string, query string, args ...interface{}) (dbexecuted *DBExecuted) {
	if cn := dbmngr.Dbms(alias); cn != nil {
		dbexecuted = &DBExecuted{}
		dbexecuted.LastInsertId, dbexecuted.RowsAffected, dbexecuted.Err = cn.Execute(query, args...)
	}
	return dbexecuted
}

//DBQuery DBQuery controller
type DBQuery struct {
	RSet         *DbResultSet
	readColsFunc ReadColumnsFunc
	readRwFunc   ReadRowFunc
	prcssFunc    ProcessingFunc
	Err          error
}

//ReadColumnsFunc definition
type ReadColumnsFunc = func(dbqry *DBQuery, columns []string, columntypes []*ColumnType)

//ReadRowFunc definition
type ReadRowFunc = func(dbqry *DBQuery, data []interface{}, firstRec bool, lastRec bool)

//ProcessingFunc definition
type ProcessingFunc = func(dbqry *DBQuery, stage QueryStage, a ...interface{})

//QueryStage stage
type QueryStage int

var qryStages = []string{"STARTED", "READING-COLUMNS", "COMPLETED-READING-COLUMNS", "READING-ROWS", "COMPLETED-READING-ROWS", "COMPLETED"}

func (qrystage QueryStage) String() (s string) {
	if qrystage > 0 && qrystage <= QueryStage(len(qryStages)) {
		s = qryStages[qrystage-1]
	} else {
		s = "UNKOWN"
	}
	return
}

//Process reading Columns then reading rows one by one till eof and finally indicate done processing
func (dbqry *DBQuery) Process() (err error) {
	var didProcess = false

	var columns = dbqry.MetaData().Columns()
	var coltypes = dbqry.MetaData().ColumnTypes()
	if dbqry.prcssFunc != nil {
		if !didProcess {
			didProcess = true
		}
		dbqry.prcssFunc(dbqry, 1)
	}
	if dbqry.prcssFunc != nil {
		if !didProcess {
			didProcess = true
		}
		dbqry.prcssFunc(dbqry, 2)
	}
	if dbqry.readColsFunc != nil {
		if !didProcess {
			didProcess = true
		}
		dbqry.readColsFunc(dbqry, columns, coltypes)
	}

	if dbqry.prcssFunc != nil {
		if !didProcess {
			didProcess = true
		}
		dbqry.prcssFunc(dbqry, 3)
	}

	if dbqry.prcssFunc != nil {
		if !didProcess {
			didProcess = true
		}
		dbqry.prcssFunc(dbqry, 4)
	}
	if dbqry.readRwFunc != nil {
		if !didProcess {
			didProcess = true
		}

		var hasRows = dbqry.Next()
		var rdata []interface{}
		var firstRow = true
		for hasRows {
			rdata = dbqry.Data()
			hasRows = dbqry.Next()
			dbqry.readRwFunc(dbqry, rdata, firstRow, !hasRows)
			if firstRow {
				firstRow = false
			}
		}
	}
	if dbqry.prcssFunc != nil {
		if !didProcess {
			didProcess = true
		}
		dbqry.prcssFunc(dbqry, 5, columns, coltypes)
	}

	if columns != nil {
		columns = nil
	}

	if coltypes != nil {
		coltypes = nil
	}

	if dbqry.prcssFunc != nil {
		if !didProcess {
			didProcess = true
		}
		dbqry.prcssFunc(dbqry, 6)
	}

	if didProcess {
		if dbqry.RSet != nil {
			err = dbqry.RSet.Close()
			dbqry.RSet = nil
		}
	}
	return
}

//DbFormatColDelimSettings DbFormatColDelimSettings
func DbFormatColDelimSettings(coldelim ...string) (readsettings map[string]string) {
	readsettings["format-type"] = "csv"
	readsettings["text-par"] = "\""
	if coldelim != nil && len(coldelim) == 1 {
		readsettings["col-sep"] = coldelim[0]
	} else {
		readsettings["col-sep"] = ","
	}
	readsettings["row-sep"] = "\r\n"
	return
}

var dbReadFormats = map[string]func() map[string]string{}
var dbReadFormattingFuncs = map[string]func(map[string]string, *DbResultSet, io.Writer) error{}

var dbReadFormatsLock = &sync.RWMutex{}

//ReadAllCustom ReadAllCustom
func (dbqry *DBQuery) ReadAllCustom(w io.Writer, settings map[string]string, formatFunction func(map[string]string, *DbResultSet, io.Writer) error) {
	if formatFunction != nil && settings != nil && len(settings) > 0 && w != nil {
		formatFunction(settings, dbqry.RSet, w)
	}
}

//ReadAll ReadAll
func (dbqry *DBQuery) ReadAll(w io.Writer, format string) {
	if w == nil || format == "" {
		return
	}
	dbReadFormatsLock.RLock()
	var settings func() map[string]string
	if settings = dbReadFormats[format]; settings != nil {
		var formatFunction = dbReadFormattingFuncs[format]
		if formatFunction != nil {
			defer dbReadFormatsLock.RUnlock()
			formatFunction(settings(), dbqry.RSet, w)
			formatFunction = nil
		} else {
			dbReadFormatsLock.RUnlock()
		}
		settings = nil
	} else {
		dbReadFormatsLock.RUnlock()
	}
}

//RegisterDbReadFormat RegisterDbReadFormat
func RegisterDbReadFormat(formatname string, settings map[string]string, formatFunction func(map[string]string, *DbResultSet, io.Writer) error) {
	if formatname != "" && settings != nil && len(settings) > 0 && formatFunction != nil {
		dbReadFormatsLock.RLock()
		defer dbReadFormatsLock.RUnlock()
	}
}

//PrintResult [refer to OutputResultSet] - helper method that output res *DbResultSet to the following formats into a io.Writer
//contentext=.js => javascript
//contentext=.json => json
//contentext=.csv => .csv
func (dbqry *DBQuery) PrintResult(out iorw.Printing, name string, contentext string, setting ...string) {
	//OutputResultSet(out, name, contentext, dbqry.RSet, dbqry.Err, setting...)
}

//MetaData return a DbResultSetMetaData object of the resultset that is wrapped by this DBQuery controller
func (dbqry *DBQuery) MetaData() *DbResultSetMetaData {
	if dbqry.RSet == nil {
		return nil
	}
	return dbqry.RSet.MetaData()
}

//Data returns an array if data of the current row from the underlying resultset
func (dbqry *DBQuery) Data() []interface{} {
	if dbqry.RSet == nil {
		return nil
	}
	return dbqry.RSet.Data()
}

func (dbqry *DBQuery) Columns() []string {
	if dbqry.RSet == nil {
		return nil
	}
	return dbqry.RSet.MetaData().Columns()
}

//Next execute the Next record method of the underlying resultset
func (dbqry *DBQuery) Next() bool {
	if dbqry.RSet == nil {
		return false
	}
	next, err := dbqry.RSet.Next()
	if err != nil {
		dbqry.Err = err
		dbqry.RSet = nil
	}
	return next
}

//Query query aliased connection and returns a DBQuery controller for the underlying resultset
func (dbmngr *DbmsManager) Query(alias string, query string, args ...interface{}) (dbquery *DBQuery) {
	if cn := dbmngr.Dbms(alias); cn != nil {
		var rdColsFunc ReadColumnsFunc
		var rdRowFunc ReadRowFunc
		var prcessFunc ProcessingFunc
		var foundOk = false
		if len(args) == 1 {
			if ad, adok := args[0].([]interface{}); adok {
				args = ad[:]
			}
		}
		if len(args) > 0 {
			var n = 0
			for n < len(args) {
				if rdColsFunc == nil {
					if rdColsFunc, foundOk = args[n].(ReadColumnsFunc); foundOk {
						if len(args) > 1 {
							args = append(args[:n], args[n+1:]...)
						} else {
							args = nil
						}
					} else {
						n++
					}
				} else if rdRowFunc == nil {
					if rdRowFunc, foundOk = args[n].(ReadRowFunc); foundOk {
						if len(args) > 1 {
							args = append(args[:n], args[n+1:]...)
						} else {
							args = nil
						}
					} else {
						n++
					}
				} else if prcessFunc == nil {
					if prcessFunc, foundOk = args[n].(ProcessingFunc); foundOk {
						if len(args) > 1 {
							args = append(args[:n], args[n+1:]...)
						} else {
							args = nil
						}
					} else {
						n++
					}
				} else {
					n++
				}
			}
		}
		dbquery = &DBQuery{readColsFunc: rdColsFunc, readRwFunc: rdRowFunc, prcssFunc: prcessFunc}

		dbquery.RSet, dbquery.Err = cn.Query(query, args...)
	}
	return dbquery
}
