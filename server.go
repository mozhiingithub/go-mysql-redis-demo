// server 本项目的后端
// 负责监听、接受和处理请求、从缓存或硬盘返回指定内容
// 本项目的大体思路是：
// 请求的url内包含欲浏览的资源标题，后端解析出标题，
// 会先查询redis缓存中是否有该资源，有则直接返回缓存，
// 若无，则查询redis中的titles集合，查看该标题是否存在于数据库中，
// 若有，则查询数据库，获取该标题对应的文本的硬盘存储地址，
// 读取文件，将内容返回给请求方的同时，写入缓存，并设置120秒的过期时间;
// 若redis的titles集合中没有请求的标题，则直接告诉请求方没有所求资源
// 综上，面对有缓存的资源，只需访问一次内存;
// 面对不存在的资源，需要访问两次内存;
// 面对无缓存的资源，需要访问两次硬盘、三次内存
// 后端端口号取8000.
// 后端运行时，程序会先清空redis中的titles集合，再重新查询数据库，
// 将所查到的title名目写入redis中
// 后端运行后，程序会启动一个协程对请求进行监听，同时主线程会
// 开启一个对os.Stdin读取的阻塞。当管理员对程序输入任意键后，
// 后端停止运行

package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	ec "./errorcheck"
	"./myredis"
	"./mysql"
	"github.com/gomodule/redigo/redis"
)

const (
	serverPort = "8000" // 端口号
	ex         = 120    // 缓存过期时间
)

func main() {
	defer myredis.Rs.Close()

	var (
		e    error
		rows *sql.Rows
	)

	// 更新redis标题集合
	myredis.Rs.Do("del", "titles")                       // 删除原有标题集合
	rows, e = mysql.DB.Query("select title from titles") // 从数据库中获取标题集合
	ec.ErrorExit(e)
	for rows.Next() { // 遍历标题集合
		var title string
		e = rows.Scan(&title)
		ec.ErrorExit(e)
		myredis.Rs.Do("sadd", "titles", title) // 将集合添加到redis中
	}
	http.HandleFunc("/", myHandler) // 注册handler
	server := &http.Server{         // 定义一个server指针
		Addr:    ":" + serverPort,
		Handler: http.DefaultServeMux,
	}
	go server.ListenAndServe() // 启动协程，监听请求
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n') // 等待输入任意键
	server.Close()          // 关闭后端
}

func myHandler(w http.ResponseWriter, r *http.Request) {
	var (
		e     error
		title string
		res   string
	)
	r.ParseForm()
	title = r.URL.String()[1:] // 获取请求的标题
	log.Println("Got title:" + title)
	res, e = redis.String(myredis.Rs.Do("get", title)) // 直接查询是否有缓存
	if nil != e {                                      // 没有缓存，判断此标题是否存在于数据库
		var isExist bool
		isExist, e = redis.Bool(myredis.Rs.Do("sismember", "titles", title))
		if isExist { // 此标题存在于数据库中，读取数据并写入缓存
			var (
				rows *sql.Rows
				dir  string
				bs   []byte
			)
			rows, e = mysql.DB.Query(fmt.Sprintf("select dir from titles where title = '%s'", title)) // 查找该标题对应的硬盘文件地址
			for rows.Next() {
				e = rows.Scan(&dir)
				ec.ErrorExit(e)
			}
			bs, e = ioutil.ReadFile(dir) // 从硬盘读取文件
			ec.ErrorExit(e)
			res = string(bs)
			log.Println("Got " + title + " from hard disk")
			_, e = myredis.Rs.Do("setex", title, ex, res) //将读取的内容写入缓存
		} else { // 此标题不存在于数据库
			log.Println("No result for " + title)
			res = "No result."
		}
	} else { // redis中有缓存，直接返回
		log.Println("Got the cache of " + title)
	}
	fmt.Fprintf(w, res)
}
