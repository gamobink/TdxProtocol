package entity

import (
	"encoding/binary"
	"math"
	"errors"
	"compress/zlib"
	"bytes"
	"io"
)

const (
	BS_BUY = 0
	BS_SELL = 1
)

const (
	STOCK_CODE_LEN = 6
)

type Transaction struct {
	Date uint32
	Minute uint16
	Price uint32
	Volume uint32
	Count uint32
	BS byte
}

type InfoExItem struct {
	Date uint32
	Bonus float32
	DeliveredShares float32
	RationedSharePrice float32
	RationedShares float32
}

type Bid struct {
	StockCode string
	Close uint32
	YesterdayClose uint32
	Open uint32
	High uint32
	Low uint32

	Vol uint32
	InnerVol uint32
	OuterVol uint32

	BuyPrice1 uint32
	SellPrice1 uint32
	BuyVol1 uint32
	SellVol1 uint32

	BuyPrice2 uint32
	SellPrice2 uint32
	BuyVol2 uint32
	SellVol2 uint32

	BuyPrice3 uint32
	SellPrice3 uint32
	BuyVol3 uint32
	SellVol3 uint32

	BuyPrice4 uint32
	SellPrice4 uint32
	BuyVol4 uint32
	SellVol4 uint32

	BuyPrice5 uint32
	SellPrice5 uint32
	BuyVol5 uint32
	SellVol5 uint32
}

type parser struct {
	RawBuffer []byte
	Current int
	Data []byte
}

type InstantTransParser struct {
	parser
	Req *InstantTransReq
}

type HisTransParser struct {
	parser
	Req *HisTransReq
}

type InfoExParser struct {
	parser
	Req *InfoExReq
}

type StockListParser struct {
	parser
	Req *StockListReq
	Total uint16
}

func (this *parser) getCmd() uint16 {
	return binary.LittleEndian.Uint16(this.RawBuffer[10:12])
}

func (this *parser) getHeaderLen() int {
	return 16
}

func (this *parser) getLen() uint16 {
	return binary.LittleEndian.Uint16(this.RawBuffer[12:14])
}

func (this *parser) getLen1() uint16 {
	return binary.LittleEndian.Uint16(this.RawBuffer[14:16])
}

func (this *parser) getSeqId() uint32 {
	return binary.LittleEndian.Uint32(this.RawBuffer[5:9])
}

func (this *parser) skipByte(count int) {
	this.Current += count
}

func (this *parser) skipData(count int) {
	for count >= 0 {
		if this.Data[this.Current] < 0x80 {
			this.skipByte(1)
		} else if this.Data[this.Current + 1] < 0x80 {
			this.skipByte(2)
		} else {
			this.skipByte(3)
		}

		count--
	}
}

func (this *parser) getByte() byte {
	ret := this.Data[this.Current]
	this.Current++
	return ret
}

func (this *parser) getUint16() uint16 {
	ret := binary.LittleEndian.Uint16(this.Data[this.Current:this.Current + 2])
	this.Current += 2
	return ret
}

func (this *parser) getUint32() uint32 {
	ret := binary.LittleEndian.Uint32(this.Data[this.Current:this.Current + 4])
	this.Current += 4
	return ret
}

func (this *parser) getFloat32() float32 {
	bits := binary.LittleEndian.Uint32(this.Data[this.Current:this.Current + 4])
	ret := math.Float32frombits(bits)
	this.Current += 4
	return ret
}

func (this *parser) parseData() int {
	v := this.Data[this.Current]
	if v >= 0x40 && v < 0x80 || v >= 0xc0 {
		return 0x40 - this.parseData2()
	} else {
		return this.parseData2()
	}
}

func (this *parser) parseData2() int {
	 //8f ff ff ff 1f == -49
	 //bd ff ff ff 1f == -3
	 //b0 fe ff ff 1f == -80
	 //8c 01		 == 76
	 //a8 fb b6 01 == 1017 万
	 //a3 8e 11    == 14.02 万
	 //82 27         == 2498
	var v int
	var nBytes int = 0
	for this.Data[this.Current + nBytes] >= 0x80 {
		nBytes++
	}
	nBytes++

	switch(nBytes){
	case 1:
		v = int(this.Data[this.Current])
	case 2:
		v = int(this.Data[this.Current+1]) * 0x40 + int(this.Data[this.Current]) - 0x80;
	case 3:
		v = (int(this.Data[this.Current+2]) * 0x80 + int(this.Data[this.Current+1]) - 0x80) * 0x40 + int(this.Data[this.Current]) - 0x80;
	case 4:
		v = ((int(this.Data[this.Current+3]) * 0x80 + int(this.Data[this.Current+2]) - 0x80) * 0x80 + int(this.Data[this.Current+1] - 0x80)) * 0x40 + int(this.Data[this.Current]) - 0x80;
	case 5:
		// over flow, positive to negative
		v = (((int(this.Data[this.Current+4]) * 0x80 + int(this.Data[this.Current+3]) - 0x80) * 0x80 + int(this.Data[this.Current+2]) - 0x80) * 0x80 + int(this.Data[this.Current+1]) - 0x80)* 0x40 + int(this.Data[this.Current]) - 0x80;
	case 6:
		// over flow, positive to negative
		v = ((((int(this.Data[this.Current+5]) * 0x80 + int(this.Data[this.Current+4]) -0x80) * 0x80 +  int(this.Data[this.Current+3]) - 0x80) * 0x80 + int(this.Data[this.Current+2]) - 0x80) * 0x80 + int(this.Data[this.Current+1]) - 0x80) * 0x40 + int(this.Data[this.Current]) - 0x80;
	default:
		panic(errors.New("bad data"))
	}
	this.skipByte(nBytes)
	return v
}

func (this *parser) uncompressIf() {
	if this.getLen() == this.getLen1() {
		this.Data = this.RawBuffer[this.getHeaderLen():]
	} else {
		b := bytes.NewReader(this.RawBuffer[this.getHeaderLen():])
		var out bytes.Buffer
		r, _ := zlib.NewReader(b)
		io.Copy(&out, r)
		this.Data = out.Bytes()
	}

	this.Current = 0
}

func (this *InstantTransParser) Parse() []*Transaction {
	if int(this.getLen()) + this.getHeaderLen() > len(this.RawBuffer) {
		panic(errors.New("incomplete data"))
	}

	if this.getSeqId() != this.Req.Header.SeqId {
		panic(errors.New("bad seq id"))
	}

	if this.getCmd() != this.Req.Header.Cmd {
		panic(errors.New("bad cmd"))
	}

	this.uncompressIf()

	var result []*Transaction

	count := this.getUint16()

	first := true
	var priceBase int

	for ; count > 0; count-- {
		trans := &Transaction{}
		trans.Minute = this.getUint16()
		if first {
			priceBase = this.parseData2()
			trans.Price = uint32(priceBase)
			first = false
		} else {
			priceBase = this.parseData() + priceBase
			trans.Price = uint32(priceBase)
		}
		trans.Volume = uint32(this.parseData2())
		trans.Count = uint32(this.parseData2())
		trans.BS = this.getByte()
		this.skipByte(1)
		result = append(result, trans)
	}
	return result
}

func NewInstantTransParser(req *InstantTransReq, data []byte) *InstantTransParser {
	return &InstantTransParser{
		parser: parser{
			RawBuffer: data,
		},
		Req: req,
	}
}

func (this *HisTransParser) Parse() []*Transaction {
	if int(this.getLen()) + this.getHeaderLen() > len(this.RawBuffer) {
		panic(errors.New("incomplete data"))
	}

	if this.getSeqId() != this.Req.Header.SeqId {
		panic(errors.New("bad seq id"))
	}

	if this.getCmd() != this.Req.Header.Cmd {
		panic(errors.New("bad cmd"))
	}

	this.uncompressIf()

	var result []*Transaction

	count := this.getUint16()
	this.skipByte(4)

	first := true
	var priceBase int

	for ; count > 0; count-- {
		trans := &Transaction{Date: this.Req.Date}
		trans.Minute = this.getUint16()
		if first {
			priceBase = this.parseData2()
			trans.Price = uint32(priceBase)
			first = false
		} else {
			priceBase = this.parseData() + priceBase
			trans.Price = uint32(priceBase)
		}
		trans.Volume = uint32(this.parseData2())
		trans.BS = this.getByte()
		trans.Count = uint32(this.parseData2())
		result = append(result, trans)
	}
	return result
}

func NewHisTransParser(req *HisTransReq, data []byte) *HisTransParser {
	return &HisTransParser{
		parser: parser{
			RawBuffer: data,
		},
		Req: req,
	}
}

func (this *InfoExParser) Parse() map[string][]*InfoExItem {
	if int(this.getLen()) + this.getHeaderLen() > len(this.RawBuffer) {
		panic(errors.New("incomplete data"))
	}

	if this.getSeqId() != this.Req.Header.SeqId {
		panic(errors.New("bad seq id"))
	}

	if this.getCmd() != this.Req.Header.Cmd {
		panic(errors.New("bad cmd"))
	}

	this.uncompressIf()

	result := map[string][]*InfoExItem{}

	count := this.getUint16()

	for ; count > 0; count-- {
		this.skipByte(1)
		stockCode := string(this.Data[this.Current:this.Current + STOCK_CODE_LEN])
		this.skipByte(STOCK_CODE_LEN)
		recordCount := this.getUint16()

		result[stockCode] = []*InfoExItem{}

		for ; recordCount > 0; recordCount-- {
			this.skipByte(1)
			stockCode1 := string(this.Data[this.Current:this.Current + STOCK_CODE_LEN])
			this.skipByte(STOCK_CODE_LEN + 1)
			if stockCode != stockCode1 {
				panic(errors.New("bad stock code"))
			}
			date := this.getUint32()
			tp := this.getByte()
			if tp != 1 {
				//fmt.Println("tp:", tp, "date:", date, "data:", hex.EncodeToString(this.Data[this.Current:this.Current+16]))
				this.skipByte(16)
				continue
			}

			obj := &InfoExItem{}
			obj.Date = date
			obj.Bonus = this.getFloat32() / 10
			obj.RationedSharePrice = this.getFloat32()
			obj.DeliveredShares = this.getFloat32() / 10
			obj.RationedShares = this.getFloat32() / 10

			result[stockCode] = append(result[stockCode], obj)
		}
	}
	return result
}

func NewInfoExParser(req *InfoExReq, data []byte) *InfoExParser {
	return &InfoExParser{
		parser: parser{
			RawBuffer: data,
		},
		Req: req,
	}
}

func (this *StockListParser) isStockValid(s []byte) bool {
	if len(s) < STOCK_CODE_LEN {
		return false
	}

	for i := 0; i < STOCK_CODE_LEN; i++ {
		if s[i] < 0x30 || s[i] > 0x39 {
			return false
		}
	}
	return true
}

func (this *StockListParser) searchStockCode() int {
	for i := this.Current; i < len(this.Data); i++ {
		if this.isStockValid(this.Data[i:]) {
			return i - this.Current - 1
		}
	}
	panic(errors.New("no stock code found"))
}

func (this *StockListParser) Parse() map[string]*Bid {
	if int(this.getLen()) + this.getHeaderLen() > len(this.RawBuffer) {
		panic(errors.New("incomplete data"))
	}

	if this.getSeqId() != this.Req.Header.SeqId {
		panic(errors.New("bad seq id"))
	}

	if this.getCmd() != this.Req.Header.Cmd {
		panic(errors.New("bad cmd"))
	}

	this.uncompressIf()

	result := map[string]*Bid{}

	totalCount := this.getUint16()
	count := this.getUint16()



	for ; count > 0; count-- {
		this.skipByte(1)	// Location
		stockCode := string(this.Data[this.Current:this.Current + STOCK_CODE_LEN])
		this.skipByte(STOCK_CODE_LEN)
		this.skipByte(2) // 未知

		bid := &Bid{StockCode: stockCode}

		bid.Close = uint32(this.parseData2())
		bid.YesterdayClose = uint32(this.parseData() + int(bid.Close))
		bid.Open = uint32(this.parseData() + int(bid.Close))
		bid.High = uint32(this.parseData() + int(bid.Close))
		bid.Low = uint32(this.parseData() + int(bid.Close))

		this.parseData()
		this.parseData()

		bid.Vol = uint32(this.parseData2())
		this.parseData2()
		this.skipByte(4)
		bid.InnerVol = uint32(this.parseData2())
		bid.OuterVol = uint32(this.parseData2())

		this.parseData()
		this.skipByte(1)

		bid.BuyPrice1 = uint32(this.parseData() + int(bid.Close))
		bid.SellPrice1 = uint32(this.parseData() + int(bid.Close))
		bid.BuyVol1 = uint32(this.parseData2())
		bid.SellVol1 = uint32(this.parseData2())

		bid.BuyPrice2 = uint32(this.parseData() + int(bid.Close))
		bid.SellPrice2 = uint32(this.parseData() + int(bid.Close))
		bid.BuyVol2 = uint32(this.parseData2())
		bid.SellVol2 = uint32(this.parseData2())

		bid.BuyPrice3 = uint32(this.parseData() + int(bid.Close))
		bid.SellPrice3 = uint32(this.parseData() + int(bid.Close))
		bid.BuyVol3 = uint32(this.parseData2())
		bid.SellVol3 = uint32(this.parseData2())

		bid.BuyPrice4 = uint32(this.parseData() + int(bid.Close))
		bid.SellPrice4 = uint32(this.parseData() + int(bid.Close))
		bid.BuyVol4 = uint32(this.parseData2())
		bid.SellVol4 = uint32(this.parseData2())

		bid.BuyPrice5 = uint32(this.parseData() + int(bid.Close))
		bid.SellPrice5 = uint32(this.parseData() + int(bid.Close))
		bid.BuyVol5 = uint32(this.parseData2())
		bid.SellVol5 = uint32(this.parseData2())

		result[stockCode] = bid

		if count > 1 {
			n := this.searchStockCode()
			this.skipByte(n)
		}
	}
	this.Total = totalCount
	return result
}

func NewStockListParser(req *StockListReq, data []byte) *StockListParser {
	return &StockListParser{
		parser: parser{
			RawBuffer: data,
		},
		Req: req,
	}
}
