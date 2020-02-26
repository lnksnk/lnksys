package db

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	parameters "github.com/efjoubert/lnksys/parameters"
)

//DbStatement container representing the underlying DbConnection and allocated sql.Tx transaction
type DbStatement struct {
	cn       *DbConnection
	tx       *sql.Tx
	stmntlck *sync.Mutex
}

//NewDbStatement invoke a new DbStatement from the DbConnection
func NewDbStatement(cn *DbConnection) (stmnt *DbStatement, err error) {
	if err = cn.db.Ping(); err == nil {
		stmnt = &DbStatement{cn: cn}
	}
	return stmnt, err
}

//Begin Invoke a transaction, sql.Tx, from the for ths DbStatement
func (stmnt *DbStatement) Begin() (err error) {
	if tx, txerr := stmnt.cn.db.Begin(); txerr == nil {
		stmnt.tx = tx
	} else {
		err = txerr
	}
	return err
}

//Execute execute a none DbResultSet query
//Usually used for statement like update and insert, or executing db procedures that dont return a record
//In the case of Insert or Update if the underlying db driver
//return the lastInsertID and rowsAffected if supported
func (stmnt *DbStatement) Execute(query string, args ...interface{}) (lastInsertID int64, rowsAffected int64, err error) {
	stmnt.stmntlck.Lock()
	defer stmnt.stmntlck.Unlock()
	if stmnt.tx == nil {
		err = stmnt.Begin()
	}

	var validNames = []string{}

	var mappedVals = map[string]interface{}{}
	var ignoreCase = false
	if len(args) == 1 {
		if pargs, ispargs := args[0].(*parameters.Parameters); ispargs {
			ignoreCase = true
			for _, skey := range pargs.StandardKeys() {
				validNames = append(validNames, skey)
				mappedVals[skey] = strings.Join(pargs.Parameter(skey), "")
			}
		} else if pmargs, ispmargs := args[0].(map[string]interface{}); ispmargs {
			for pmk, pmv := range pmargs {
				if mpv, mpvok := pmv.(map[string]interface{}); mpvok && mpv != nil {

				} else {
					validNames = append(validNames, pmk)
					mappedVals[pmk] = fmt.Sprint(pmv)
				}
			}
		}
	}
	var parseQry = stmnt.cn.ParseQuery(query, ignoreCase, validNames...)
	var txtArgs = []interface{}{}
	var qry = ""

	var parseParam = func() {
		if stmnt.cn.schema == "sqlserver" {
			qry += ("@p" + fmt.Sprintf("%d", len(txtArgs)))
		} else if stmnt.cn.schema == "postgres" {
			qry += ("$" + fmt.Sprintf("%d", len(txtArgs)))
		} else {
			qry += "?"
		}
	}

	var foundPrm = false
	if len(parseQry) > 0 {
		if len(args) > 0 {
			for _, prm := range parseQry {
				if len(prm) > 2 && prm[0] == '@' && prm[len(prm)-1] == '@' {
					var prmtest = prm[1 : len(prm)-1]
					foundPrm = false
					var prmval interface{} = nil
					if ignoreCase {
						prmtest = strings.ToUpper(prmtest)
					}
					if mpdval, mapdvalok := mappedVals[prmtest]; mapdvalok {
						prmval = mpdval
						txtArgs = append(txtArgs, prmval)
						parseParam()

						foundPrm = true
					} else {
						for _, d := range args {
							if prms, prmsok := d.(*parameters.Parameters); prmsok {
								if prms.ContainsParameter(prmtest) {
									prmval = strings.Join(prms.Parameter(prmtest), "")
									mappedVals[prmtest] = prmval
									txtArgs = append(txtArgs, prmval)
									parseParam()
									foundPrm = true
									break
								}
							}
						}
					}
					if !foundPrm {
						txtArgs = append(txtArgs, prm)
					}
				} else {
					qry += prm
				}
			}
		} else {
			for _, prm := range parseQry {
				qry += prm
			}
		}
	}
	if err == nil {
		if r, rerr := stmnt.tx.Exec(qry, txtArgs...); rerr == nil {
			lastInsertID, err = r.LastInsertId()
			rowsAffected, err = r.RowsAffected()
			r = nil
			err = stmnt.tx.Commit()
		} else {
			err = rerr
		}
	}
	if err != nil {
		if rolerr := stmnt.tx.Rollback(); rolerr != nil {
			err = rolerr
		}
	}
	return lastInsertID, rowsAfdfected, err
}

//Query and return a DbResultSet
func (stmnt *DbStatement) Query(query string, args ...interface{}) (rset *DbResultSet, err error) {
	stmnt.stmntlck.Lock()
	defer stmnt.stmntlck.Unlock()
	if stmnt.tx == nil {
		err = stmnt.Begin()
	}

	var validNames = []string{}

	var mappedVals = map[string]interface{}{}
	var ignoreCase = false
	if len(args) == 1 {
		if pargs, ispargs := args[0].(*parameters.Parameters); ispargs {
			ignoreCase = true
			for _, skey := range pargs.StandardKeys() {
				validNames = append(validNames, skey)
				mappedVals[skey] = strings.Join(pargs.Parameter(skey), "")
			}
		} else if pmargs, ispmargs := args[0].(map[string]interface{}); ispmargs {
			for pmk, pmv := range pmargs {
				if mpv, mpvok := pmv.(map[string]interface{}); mpvok && mpv != nil {

				} else {
					validNames = append(validNames, pmk)
					mappedVals[pmk] = fmt.Sprint(pmv)
				}
			}
		}
	}
	var parseQry = stmnt.cn.ParseQuery(query, ignoreCase, validNames...)
	var txtArgs = []interface{}{}
	var qry = ""

	var parseParam = func() {
		if stmnt.cn.schema == "sqlserver" {
			qry += ("@p" + fmt.Sprintf("%d", len(txtArgs)))
		} else if stmnt.cn.schema == "postgres" {
			qry += ("$" + fmt.Sprintf("%d", len(txtArgs)))
		} else {
			qry += "?"
		}
	}

	var foundPrm = false
	if len(parseQry) > 0 {
		if len(args) > 0 {
			for _, prm := range parseQry {
				if len(prm) > 2 && prm[0] == '@' && prm[len(prm)-1] == '@' {
					var prmtest = prm[1 : len(prm)-1]
					foundPrm = false
					var prmval interface{} = nil
					if ignoreCase {
						prmtest = strings.ToUpper(prmtest)
					}
					if mpdval, mapdvalok := mappedVals[prmtest]; mapdvalok {
						prmval = mpdval
						txtArgs = append(txtArgs, prmval)
						parseParam()

						foundPrm = true
					} else {
						for _, d := range args {
							if prms, prmsok := d.(*parameters.Parameters); prmsok {
								if prms.ContainsParameter(prmtest) {
									prmval = strings.Join(prms.Parameter(prmtest), "")
									mappedVals[prmtest] = prmval
									txtArgs = append(txtArgs, prmval)
									parseParam()
									foundPrm = true
									break
								}
							}
						}
					}
					if !foundPrm {
						txtArgs = append(txtArgs, prm)
					}
				} else {
					qry += prm
				}
			}
		} else {
			for _, prm := range parseQry {
				qry += prm
			}
		}
	}
	if rs, rserr := stmnt.tx.Query(qry, txtArgs...); rserr == nil {
		if cols, colserr := rs.Columns(); colserr == nil {
			for n, col := range cols {
				if col == "" {
					cols[n] = "COLUMN" + fmt.Sprint(n+1)
				}
			}
			if coltypes, coltypeserr := rs.ColumnTypes(); coltypeserr == nil {
				rset = &DbResultSet{rset: rs, stmnt: stmnt, rsmetadata: &DbResultSetMetaData{cols: cols[:], colTypes: columnTypes(coltypes[:])}, dosomething: make(chan bool, 1)}
			} else {
				err = coltypeserr
			}
		} else {
			err = colserr
		}
	} else {
		if rolerr := stmnt.tx.Rollback(); rolerr != nil {
			rserr = rolerr
		}
		err = rserr
	}
	return rset, err
}

//Close the allocated transaction, sql.Tx associated to this DbStatement
//It will by default perform a commit before releasing the transaction reference
func (stmnt *DbStatement) Close() (err error) {
	if stmnt.tx != nil {
		err = stmnt.tx.Commit()
		stmnt.tx = nil
	}
	if stmnt.cn != nil {
		stmnt.cn = nil
	}
	return
}

func columnTypes(sqlcoltypes []*sql.ColumnType) (coltypes []*ColumnType) {
	coltypes = make([]*ColumnType, len(sqlcoltypes))
	for n, ctype := range sqlcoltypes {
		coltype := &ColumnType{}
		coltype.databaseType = ctype.DatabaseTypeName()
		coltype.length, coltype.hasLength = ctype.Length()
		coltype.name = ctype.Name()
		coltype.databaseType = ctype.DatabaseTypeName()
		coltype.nullable, coltype.hasNullable = ctype.Nullable()
		coltype.precision, coltype.scale, coltype.hasPrecisionScale = ctype.DecimalSize()
		coltype.scanType = ctype.ScanType()
		coltypes[n] = coltype
	}
	return coltypes
}
