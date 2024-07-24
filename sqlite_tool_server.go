package sqlite_tool_server

import (
	"fmt"

	"github.com/panjf2000/gnet/v2"
	"github.com/yudaiyan/go-sqlite-tool-server/server"
	"github.com/yudaiyan/go-sync/sync/future"
	"gorm.io/gorm"
)

func StartServer(db *gorm.DB, port int) future.IFuture[error] {
	f := future.New[error]()

	ss := &server.Server{
		Network:   "tcp",
		Addr:      fmt.Sprintf("0.0.0.0:%d", port),
		Multicore: false,
		DB:        db,
	}
	go func() {
		err := gnet.Run(ss, fmt.Sprintf("%s://%s", ss.Network, ss.Addr), gnet.WithMulticore(false), gnet.WithReusePort(true), gnet.WithReuseAddr(true))
		f.Set(err)
	}()
	//启动tcp服务器
	return f
}
