// errorcheck模块
// 将错误检查函数独立成包，是因为很多模块都需要进行错误检查
// ErrorExit函数会先判断传入错误是否为空
// 若非空，则输出错误内容并强制终止程序
package errorcheck

import (
	"log"
	"os"
)

func ErrorExit(e error) {
	if nil != e {
		log.Println(e)
		os.Exit(1)
	}
}
