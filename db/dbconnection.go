package db

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"sync"
)

type DbConnection struct {
	dbms       *DbmsManager
	driverName string
	cnsettings map[string]string
	query      url.Values
	username   string
	password   string
	path       string
	host       string
	schema     string
	cnurl      *url.URL
	db         *sql.DB
	dbcnlck    *sync.Mutex
}

func NewDbConnection(cnsettings ...string) (dbcn *DbConnection) {
	dbcn = &DbConnection{cnsettings: map[string]string{}, dbcnlck: &sync.Mutex{}}

	dbcn.LoadCnSettings(cnsettings...)
	return
}

func (dbcn *DbConnection) LockDBCN() {
	dbcn.dbcnlck.Lock()
}

func (dbcn *DbConnection) UnlockDBCN() {
	dbcn.dbcnlck.Unlock()
}

func (dbcn *DbConnection) Driver() string {
	return dbcn.driverName
}

//query e.g 'SELECT :TEST-PARAM AS param' where :TEST-PARAM is name of parameter
func (dbcn *DbConnection) ParseQuery(query string, ignoreCase bool, validNames ...string) (parsedquery []string) {
	startParam := false
	startText := false
	pname := ""

	var valNames = map[string]bool{}

	if len(validNames) > 0 {
		for _, vnme := range validNames {
			if ignoreCase {
				vnme = strings.ToUpper(vnme)
			}
			if isvnme, vnmeok := valNames[vnme]; vnmeok {
				if !isvnme {
					valNames[vnme] = true
				}
			} else {
				valNames[vnme] = true
			}
		}
	}
	var tmpquery = ""

	for _, c := range query {
		if c == '\'' {
			if startText {
				startText = false
			} else {
				startText = true
			}
		}
		if startParam {
			if c == '@' {
				if pname != "" {
					testName := pname
					if ignoreCase {
						testName = strings.ToUpper(testName)
					}
					if len(valNames) > 0 {

						if isvnme, vnmeok := valNames[testName]; vnmeok && isvnme {
							if tmpquery != "" {
								parsedquery = append(parsedquery, tmpquery)
								tmpquery = ""
							}
							parsedquery = append(parsedquery, "@"+testName+"@")
							pname = ""
						} else {
							tmpquery = tmpquery + string(c) + pname + string(c)
						}
					} else {
						if tmpquery != "" {
							parsedquery = append(parsedquery, tmpquery)
							tmpquery = ""
						}
						parsedquery = append(parsedquery, "@"+testName+"@")
						pname = ""
					}
				} else {
					tmpquery = tmpquery + string(c) + pname + string(c)
				}
				pname = ""
				startParam = false
			} else {
				pname += string(c)
			}
		} else {
			if c == '@' {
				startParam = true
			} else {
				tmpquery += string(c)
			}
		}
	}

	if startParam {
		if pname != "" {
			tmpquery = tmpquery + "@" + pname
		}
	}

	if tmpquery != "" {
		parsedquery = append(parsedquery, tmpquery)
	}

	return parsedquery
}

func (dbcn *DbConnection) Schema() string {
	return dbcn.schema
}

func (dbcn *DbConnection) Path() string {
	return dbcn.path
}

func (dbcn *DbConnection) DbName() string {
	if dbcn.path != "" {
		return dbcn.path
	} else if dbnme, dbnmeok := dbcn.cnsettings["database"]; dbnmeok {
		return dbnme
	}
	return ""
}

func (dbcn *DbConnection) User() string {
	return dbcn.username
}

func (dbcn *DbConnection) AltUser() string {
	if altuser, altuserok := dbcn.cnsettings["alt-user"]; altuserok {
		return altuser
	}
	return dbcn.username
}

func (dbcn *DbConnection) LoadCnSettings(cnsettings ...string) {
	var changedSettings = []string{}
	for _, cns := range cnsettings {
		if strings.Index(cns, "=") > -1 {
			var sname = strings.TrimSpace(cns[0:strings.Index(cns, "=")])
			var sval = strings.TrimSpace(cns[strings.Index(cns, "=")+1:])
			if sname != "" && sval != "" {
				if cnval, cnok := dbcn.cnsettings[sname]; cnok && cnval != sval {
					dbcn.cnsettings[sname] = sval
				} else {
					dbcn.cnsettings[sname] = sval
					changedSettings = append(changedSettings, sname)
				}
			}
		}
	}
	if len(changedSettings) > 0 {
		if dbcn.cnurl != nil {
			dbcn.cnurl = nil
		}
		dbcn.cnurl = &url.URL{}
		var additionalsettings = map[string]string{}
		for _, cname := range changedSettings {
			if cname == "user id" || cname == "username" {
				dbcn.username = dbcn.cnsettings[cname]
			} else if cname == "password" {
				dbcn.password = dbcn.cnsettings[cname]
			} else if cname == "host" {
				var host = dbcn.cnsettings[cname]
				if strings.Index(host, "/") > -1 {
					dbcn.host = host[0:strings.Index(host, "/")]
					dbcn.path = host[strings.Index(host, "/"):]
				} else {
					dbcn.host = host
				}

			} else if cname == "instance" {
				dbcn.path = dbcn.cnsettings[cname]
			} else if cname == "driver" {
				dbcn.driverName = dbcn.cnsettings[cname]
			} else if cname == "schema" {
				dbcn.schema = dbcn.cnsettings[cname]
			} else {
				additionalsettings[cname] = dbcn.cnsettings[cname]
			}
		}
		if dbcn.driverName != "" {
			if dbcn.schema == "" {
				if dbcn.driverName == "pgx" || dbcn.driverName == "pq" {
					dbcn.schema = "postgres"
				} else {
					dbcn.schema = dbcn.driverName
				}
			}
			dbcn.cnurl.Scheme = dbcn.schema
			if dbcn.username != "" /*&& dbcn.password != ""*/ {
				dbcn.cnurl.User = url.UserPassword(dbcn.username, dbcn.password)
			}
			if dbcn.host != "" {
				dbcn.cnurl.Host = dbcn.host
				if dbcn.path != "" {
					dbcn.cnurl.Path = dbcn.path
				}
			}

		}
		if len(additionalsettings) > 0 {
			var addquery = url.Values{}
			for addkey, addv := range additionalsettings {
				addquery.Add(addkey, addv)
			}
			dbcn.cnurl.RawQuery = addquery.Encode()
			addquery = nil
		}
		var cnstring = dbcn.cnurl.String()

		if dbcn.driverName == "mysql" {
			cnstring = dbcn.cnurl.User.String() + "@tcp(" + dbcn.host + ")/" + dbcn.Path() + dbcn.cnurl.RawQuery

			// user:password@tcp(localhost:5555)/dbname?tls=skip-verify&autocommit=true
		}

		if db, err := sql.Open(dbcn.driverName, cnstring); err == nil {
			if err = db.Ping(); err == nil {
				if dbcn.db != nil {
					dbcn.db.Close()
					dbcn.db = nil
				}
				dbcn.db = db
			} else {
				fmt.Println(err)
			}
		} else {
			fmt.Println(err)
		}

	}
}

//Execute - Execute (query) refer to golang sql connection Execute method
func (dbcn *DbConnection) Execute(query string, args ...interface{}) (lastInsertID int64, rowsAffected int64, err error) {
	dbcn.LockDBCN()
	defer func() {
		dbcn.UnlockDBCN()
	}()
	stmnt := &DbStatement{cn: dbcn}
	lastInsertID, rowsAffected, err = stmnt.Execute(query, args...)
	stmnt = nil
	return
}

//Query - Query (query) refer to golang sql connection Query method
//except that it returns and DbResultSet that extends standard resultset functionality
func (dbcn *DbConnection) Query(query string, args ...interface{}) (rset *DbResultSet, err error) {
	dbcn.LockDBCN()
	defer func() {
		dbcn.UnlockDBCN()
	}()
	stmnt := &DbStatement{cn: cn}
	rset, err = stmnt.Query(query, args...)
	return rset, err
}
