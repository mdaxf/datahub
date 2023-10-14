// Copyright 2023 IAC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dbconn

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/mdaxf/iac/logger"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

var (
	/*
		mysql,sqlserver, goracle
	*/
	DatabaseType = "mysql"

	/*
		user:password@tcp(localhost:3306)/mydb
		server=%s;port=%d;user id=%s;password=%s;database=%s
	*/
	//DatabaseConnection = "server=xxx;user id=xx;password=xxx;database=xxx"  //sqlserver
	DatabaseConnection = "user:iacf12345678@tcp(localhost:3306)/iac"

	once sync.Once
	err  error
)
var iLog logger.Log

func Initialize(databaseconfig map[string]interface{}) error {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "Database"}
	iLog.Info(fmt.Sprintf("Connect Database with: %s %s", DatabaseType, DatabaseConnection))

	DatabaseType = databaseconfig["type"].(string)
	DatabaseConnection = databaseconfig["connection"].(string)

	err := ConnectDB()
	if err != nil {
		iLog.Error(fmt.Sprintf("initialize Database error: %s", err.Error()))
		return err
	}
	return nil
}
func ConnectDB() error {

	iLog.Info(fmt.Sprintf("Connect Database: %s %s", DatabaseType, DatabaseConnection))

	once.Do(func() {
		DB, err = sql.Open(DatabaseType, DatabaseConnection)
		if err != nil {
			iLog.Error(fmt.Sprintf("initialize Database error: %s", err.Error()))
			panic(err.Error())
		}
		DB.SetMaxIdleConns(1000)
		DB.SetMaxOpenConns(10000)
	})
	iLog.Info("Database connected")
	return nil
}

func DBPing() error {

	return DB.Ping()

}
