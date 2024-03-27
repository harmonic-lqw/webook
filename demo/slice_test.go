package demo

import "fmt"

func removeWeightConn(conns *[]*weightConn, conn *weightConn) {
	for i, c := range *conns {
		if c == conn {
			// 找到要删除的元素，使用切片操作删除它
			*conns = append((*conns)[:i], (*conns)[i+1:]...)
			return
		}
	}
	// 如果元素不在切片中，可以选择打印一条消息或进行其他处理
	fmt.Println("weightConn not found in the slice")
}
