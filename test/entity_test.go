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

func BuildStockListBuffer() (*bytes.Buffer, *entity.StockListReq) {
	req := entity.NewStockListReq(1, 0, 0, 2)
	buf := new(bytes.Buffer)
	req.Write(buf)
	return buf, req
}

func TestStockListReq(t *testing.T) {
	fmt.Println("TestStockListReq...")
	buf, req := BuildStockListBuffer()

	conn, err := net.Dial("tcp", HOST)
	chk(err)

	_, err = conn.Write(buf.Bytes())
	chk(err)

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	chk(err)

	parser := entity.NewStockListParser(req, buffer[:n])
	result := parser.Parse()
	fmt.Println(hex.EncodeToString(parser.Data))

	fmt.Println("total:", parser.Total)
	for _, b := range result {
		fmt.Println(b)
	}
}

func BuildInfoExBuffer() (*bytes.Buffer, *entity.InfoExReq) {
	req := entity.NewInfoExReq(1)
	req.AddCode("000099")
	fmt.Println(req)
	buf := new(bytes.Buffer)
	req.Write(buf)
	fmt.Println(buf.Bytes())
	return buf, req
}

func _TestInfoExReq(t *testing.T) {
	fmt.Println("TestInfoExReq...")
	buf, req := BuildInfoExBuffer()

	conn, err := net.Dial("tcp", HOST)
	chk(err)

	_, err = conn.Write(buf.Bytes())
	chk(err)

	buffer := make([]byte, 1024 * 1024)
	n, err := conn.Read(buffer)
	chk(err)

	parser := entity.NewInfoExParser(req, buffer[:n])
	result := parser.Parse()
	fmt.Println(hex.EncodeToString(parser.Data))

	for k, l := range result {
		fmt.Println(k)
		for _, t := range l {
			fmt.Println(t)
		}
	}
}

func BuildInstantTransBuffer() (*bytes.Buffer, *entity.InstantTransReq){
	req := entity.NewInstantTransReq(1, "300629", 1655, 300)
	buf := new(bytes.Buffer)
	req.Write(buf)
	return buf, req
}

func _TestInstantTransReq(t *testing.T) {
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

func BuildHisTransBuffer() (*bytes.Buffer, *entity.HisTransReq) {
	req := entity.NewHisTransReq(1, 20170414, "600000", 3800, 200)
	buf := new(bytes.Buffer)
	req.Write(buf)
	return buf, req
}

func _TestHisTransReq(t *testing.T) {
	fmt.Println("TestHisTransReq...")
	buf, req := BuildHisTransBuffer()

	conn, err := net.Dial("tcp", HOST)
	chk(err)

	_, err = conn.Write(buf.Bytes())
	chk(err)

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	chk(err)

	parser := entity.NewHisTransParser(req, buffer[:n])
	result := parser.Parse()
	fmt.Println(hex.EncodeToString(parser.Data))

	for _, t := range result {
		fmt.Println(t)
	}
}

func Test(t *testing.T) {
}