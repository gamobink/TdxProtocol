package main

import (
	"github.com/TdxProtocol/entity"
	"fmt"
	"bytes"
	"net"
	"encoding/hex"
)

const (
	HOST = "125.39.80.98:7709"
)

func chk(err error) {
	if err == nil {
		return
	}

	fmt.Println(err)
	panic(err)
}

func main() {
	req := entity.NewInfoExReq(1)
	req.AddCode("600000")
	req.AddCode("600001")
	fmt.Println(req)
	buf := new(bytes.Buffer)
	req.Write(buf)
	fmt.Println(buf.Bytes())

	conn, err := net.Dial("tcp", HOST)
	chk(err)

	_, err = conn.Write(buf.Bytes())
	chk(err)

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	chk(err)
	fmt.Println(hex.EncodeToString(buffer[:n]))
}
