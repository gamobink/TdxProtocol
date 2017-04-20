package test

import (
	"github.com/TdxProtocol/entity"
	"fmt"
	"bytes"
	"net"
	"encoding/hex"
	"testing"
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

func BuildStockListBuffer() *bytes.Buffer {
	req := entity.NewStockListReq(1, 0, 0, 1)
	fmt.Println(req)
	buf := new(bytes.Buffer)
	req.Write(buf)
	fmt.Println(buf.Bytes())
	return buf
}

func _TestStockListReq(t *testing.T) {
	fmt.Println("TestStockListReq...")
	buf := BuildStockListBuffer()

	conn, err := net.Dial("tcp", HOST)
	chk(err)

	_, err = conn.Write(buf.Bytes())
	chk(err)

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	chk(err)
	fmt.Println(hex.EncodeToString(buffer[:n]))
}

func BuildInfoExBuffer() *bytes.Buffer {
	req := entity.NewInfoExReq(1)
	req.AddCode("600000")
	req.AddCode("600001")
	fmt.Println(req)
	buf := new(bytes.Buffer)
	req.Write(buf)
	fmt.Println(buf.Bytes())
	return buf
}

func _TestInfoExReq(t *testing.T) {
	fmt.Println("TestInfoExReq...")
	buf := BuildInfoExBuffer()

	conn, err := net.Dial("tcp", HOST)
	chk(err)

	_, err = conn.Write(buf.Bytes())
	chk(err)

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	chk(err)
	fmt.Println(hex.EncodeToString(buffer[:n]))
}

func BuildInstantTransBuffer() (*bytes.Buffer, *entity.InstantTransReq){
	req := entity.NewInstantTransReq(1, "300629", 1655, 300)
	buf := new(bytes.Buffer)
	req.Write(buf)
	return buf, req
}

func TestInstantTransReq(t *testing.T) {
	fmt.Println("TestInstantTransReq...")
	buf, req := BuildInstantTransBuffer()

	conn, err := net.Dial("tcp", HOST)
	chk(err)

	_, err = conn.Write(buf.Bytes())
	chk(err)

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	chk(err)

	parser := entity.NewInstantTransParser(req, buffer[:n])
	result := parser.Parse()
	fmt.Println(hex.EncodeToString(parser.Data))

	for _, t := range result {
		fmt.Println(t)
	}
}

func BuildHisTransBuffer() *bytes.Buffer {
	req := entity.NewHisTransReq(1, 20170414, "600000", 0, 100)
	buf := new(bytes.Buffer)
	req.Write(buf)
	return buf
}

func _TestHisTransReq(t *testing.T) {
	fmt.Println("TestHisTransReq...")
	buf := BuildHisTransBuffer()

	conn, err := net.Dial("tcp", HOST)
	chk(err)

	_, err = conn.Write(buf.Bytes())
	chk(err)

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	chk(err)
	fmt.Println(hex.EncodeToString(buffer[:n]))
}
