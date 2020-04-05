// mysql模块
// 负责对数据库的连接进行初始化
// DB是一个全局变量，由init函数进行初始化后，
// 供其他模块使用，进行数据库的增删改查
// 本项目的数据库，名为demo
// init函数中，设置的数据库最长连接时间为10秒，
// 最大连接数为20,最大闲时连接数为10
package mysql

import (
	"database/sql"
	"time"

	ec "../errorcheck"
	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func init() {
	var e error
	DB, e = sql.Open("mysql", "root:@tcp(localhost:3306)/demo")
	ec.ErrorExit(e)
	DB.SetConnMaxLifetime(10 * time.Second)
	DB.SetMaxOpenConns(20)
	DB.SetMaxIdleConns(10)
	e = DB.Ping()
	ec.ErrorExit(e)
	return
}
