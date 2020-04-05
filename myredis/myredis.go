// myredis模块
// 负责对redis的连接初始化
// Rs是一个全局变量，经init函数初始化后，
// 供其他模块进行redis缓存的增删改查
package myredis

import (
	"log"
	"os"

	"github.com/gomodule/redigo/redis"
)

var Rs redis.Conn

func init() {
	var e error
	Rs, e = redis.Dial("tcp", "127.0.0.1:6379")
	if nil != e {
		Rs.Close()
		log.Println(e)
		os.Exit(1)
	}
}
