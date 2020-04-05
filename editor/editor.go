// editor模块，负责后端运行后，对整个项目内的资源进行增删改工作
// 包括：增加资源、修改资源内容、修改资源文件地址、删除资源
// 其中：
// 增加资源时，向数据库写入资源名称及文件地址，
// 将资源名写入redis的titles集合中，但不会将资源内容写入redis
// 因为写入数据库后，新资源属于“无缓存资源”的情况，
// 该资源被请求时，后端会自动从硬盘中读取内容并写入redis缓存
// 修改资源的情况也非常类似，在修改完数据库的内容后，同步删除redis混存，
// 使修改后的资源变成“无缓存资源”，交由后端完成redis缓存的更新工作
// 删除资源的步骤则是同步删除数据库内容和redis缓存，但注意，
// 删除资源只是指将该资源从项目名录中去除，而非删除硬盘中的资源文件
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	ec "../errorcheck"
	"../myredis"
	"../mysql"
)

func main() {
	defer myredis.Rs.Close()
	s := readStr()
	switch s {
	case "insert": // 插入
		{
			title, dir := getTitleDir()
			stmt, e1 := mysql.DB.Prepare("insert into titles (title,dir) values (?,?)")
			ec.ErrorExit(e1)
			stmt.Exec(title, dir)                  // 将资源名和资源文件地址写入数据库
			myredis.Rs.Do("sadd", "titles", title) // 将资源名写入redis的titles集合中

		}
	case "update": // 更新
		{
			title, dir := getTitleDir()
			if "" != dir { // 更新资源文件地址
				stmt, e1 := mysql.DB.Prepare("update titles set dir = ? where title = ?")
				ec.ErrorExit(e1)
				stmt.Exec(dir, title) // 更新数据库中的地址
			}
			myredis.Rs.Do("del", title) // 无论是否更新过文件地址，都清除redis对应的资源缓存
		}
	case "delete": // 删除
		{
			fmt.Print("title:")
			title := readStr()
			stmt, e1 := mysql.DB.Prepare("delete from titles where title = ?")
			ec.ErrorExit(e1)
			stmt.Exec(title)                       // 删除数据库中的该资源名目
			myredis.Rs.Do("del", title)            // 清除redis对应的资源缓存
			myredis.Rs.Do("srem", "titles", title) // 清除redis titles集合中的名目
		}
	default:
		{
			log.Println("invalid command.")
		}
	}
}

func getTitleDir() (title string, dir string) {
	fmt.Print("title:")
	title = readStr()
	fmt.Print("dir:")
	dir = readStr()
	return
}

func readStr() (s string) {
	reader := bufio.NewReader(os.Stdin)
	s, e := reader.ReadString('\n')
	ec.ErrorExit(e)
	s = s[:len(s)-1]
	return
}
