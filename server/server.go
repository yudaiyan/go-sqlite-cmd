package server

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/go-errors/errors"
	"github.com/panjf2000/gnet/v2"
	"github.com/yudaiyan/go-sqlite-tool-server/pkg/db"
	"gorm.io/gorm"
)

// 用于转发的gnet server
type Server struct {
	gnet.BuiltinEventEngine
	eng       gnet.Engine
	Network   string
	Addr      string
	Multicore bool
	DB        *gorm.DB
}

func (s *Server) OnBoot(eng gnet.Engine) (action gnet.Action) {
	log.Printf("running server on %s with multi-core=%t",
		fmt.Sprintf("%s://%s", s.Network, s.Addr), s.Multicore)
	s.eng = eng
	return
}

func (s *Server) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	var outStr string
	defer func() {
		out = []byte(outStr)
	}()

	log.Printf("new connect %s", c.RemoteAddr())

	version, err := db.Version(s.DB)
	if err != nil {
		msg := err.(*errors.Error).ErrorStack()
		log.Println(msg)
		outStr += fmt.Sprintln(msg)
		return
	}

	outStr += fmt.Sprintf("SQLite version %s\n", version)
	outStr += "Enter \".help\" or \".h\" for usage hints.\n"
	outStr += "> "
	return
}

func (s *Server) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	if err != nil {
		log.Printf("error occurred on connection=%s, %v\n", c.RemoteAddr().String(), err)
	}
	log.Printf("删除连接[%s]", c.RemoteAddr().Network())
	return
}

func (s *Server) OnTraffic(c gnet.Conn) (action gnet.Action) {
	buf, err := c.Next(-1)
	if err != nil {
		log.Printf("invalid packet: %v", err)
		return gnet.Close
	}
	log.Println("收到数据:", buf)

	var out string
	cmd := strings.TrimSpace(string(buf))
	switch {
	case cmd == ".table":
		out, err = db.Table(s.DB)
	case regexp.MustCompile(`(?i)^[\s\t]*SELECT`).MatchString(cmd):
		out, err = db.Select(s.DB, cmd)
	case cmd == "q":
		return gnet.Close
	case cmd == "": // 由于trim，截取了\r\n，因此换行只剩下空字符串
	case cmd == ".help" || cmd == ".h":
		out = `.help                   Show this message
.table                  Show table name
SELECT                  Show table data
q                               Quit
`

	default:
		out = fmt.Sprintf("非法的命令 %s\n", cmd)
	}

	if err != nil {
		msg := err.(*errors.Error).ErrorStack()
		log.Println(msg)
		out += fmt.Sprintln(msg)
	}

	out += ("> ")
	if _, err = c.Write([]byte(out)); err != nil {
		log.Printf("invalid packet: %v", err)
		return gnet.Close
	}
	return
}
