package conn

import (
	crand "crypto/rand"
	mrand "math/rand"
	"net"
	"strconv"
	"io"
	"time"

	"encoding/hex"
	. "github.com/stealpipe/lib/debug"
	"strings"
)

const (
	max_header_length = 64 * 1024
	max_chunk_length  = 20
	no_Mode           = -1
	chunk_Mode        = 100
	content_Mode      = 200
	READBUFF          = 1024 * 4
)

const (
	READ_HEADER      = 100
	READ_CHUNK       = 200
	READ_CHUNKDATA   = 300
	READ_CONTENTDATA = 400
)

type ReaderStatus struct {
	readstatus     int
	contentdatalen int
	readedlen      int
	chunkLength    int
	X_readeddata   []byte
	readedbuf      []byte
	header         []byte
}

type WriteStatus struct {
	datatypes int // chunk or contentlength
	datalen   int //
	sended    int //
}

type HttpSocket struct {
	conn Pipe
	w    WriteStatus
	r    ReaderStatus
}

func getmrandCreater() *mrand.Rand {
	s2 := mrand.NewSource(int64(time.Now().Nanosecond()))
	return mrand.New(s2)
}

func (p *HttpSocket) Init(n Pipe, k []byte) {
	p.conn = n
	p.w.datatypes = no_Mode
	p.r.readstatus = READ_HEADER
}

func (p *HttpSocket) CreateIv() bool {
	return true
}

func (p *HttpSocket) ReadFakeHead() bool {
	return true
}

func (p *HttpSocket) ReadPackageFake() bool {
	return true
}

func (p *HttpSocket) GetConn() net.Conn {
	return p.conn
}

func (p *HttpSocket) ReadIv() bool {
	return true
}

func (p *HttpSocket) ReadyRead() bool {
	return true
}

func (p *HttpSocket) ReadyWrite() bool {
	return true
}

func (p *HttpSocket) Close() {
	p.conn.Close()
}

func (p *HttpSocket) SetSpeed(speed int64) {
}

func (p *HttpSocket) SetFakeHeaderLength(len int64) {
}
func (p *HttpSocket) SetFakeHeaderPaddingIndex(int64) {

}
func (p *HttpSocket) SetPackageFakeLength(int64) {
}

func (p *HttpSocket) Read() (bool, []byte) {

	if p.r.X_readeddata != nil {
		pdata := p.r.X_readeddata
		p.r.X_readeddata = nil
		return true, pdata
	}
	return p.ReadData()
}

func (p *HttpSocket) ReadData() (bool, []byte) {

	for {

		Info("begin ReadData")
		if p.r.readstatus == READ_HEADER {

			Info("Read header")
			b, _ := p.Read_Header()

			if !b {
				panic("not a header")
				return false, nil
			}
			continue
		}
		if p.r.readstatus == READ_CHUNK {

			Info("Read Chunk")
			b, _ := p.Read_Chunk()
			if !b {
				return false, nil
			}
		}
		if p.r.readstatus == READ_CHUNKDATA {

			Info("Read ChunkData")
			b, p := p.Read_ChunkData()
			if !b {
				return false, nil
			}
			return b, p
		}
		if p.r.readstatus == READ_CONTENTDATA {

			Info("Read Content")
			b, p := p.Read_Content()
			if !b {
				return false, nil
			}
			return b, p
		}
		panic("p.r.readstatus")
	}
}

func (p *HttpSocket) Read_Header() (bool, []byte) {

	if p.r.header != nil {
		panic(p.r.header)
	}
	for {
		if p.r.readedbuf == nil {
			newdata := make([]byte, READBUFF)
			n, err := p.conn.Read(newdata)
			if err != nil {
				panic("read conn, no data")
				return false, nil
			}
			p.r.readedbuf = newdata[0:n]
		}
		if p.r.header == nil {
			p.r.header = p.r.readedbuf
		} else {
			p.r.header = append(p.r.header, p.r.readedbuf...)
		}
		p.r.readedbuf = nil
		s := string(p.r.header)
		Debug("http header ", len(s))
		head_length := strings.Index(s, "\r\n\r\n")
		Debug("http header len ", head_length)
		if head_length == -1 {
			if len(s) > max_header_length {
				Info(s)
				panic("len(s) > max_header_length")
				return false, nil
			}
			continue
		} else {
			head_length += 4
			Info("drop http header length :", head_length)
			if len(s) == head_length {
				p.r.readedbuf = nil
			} else {
				p.r.readedbuf = []byte(s[head_length:])
			}
		}
		s = s[:head_length]
		if strings.Index(s, "Transfer-Encoding") != -1 {
			p.r.readstatus = READ_CHUNK
			p.r.header = nil
			return true, nil
		}

		v := strings.Split(s, "\r\n")
		for _, i := range v {
			linev := strings.Split(i, ":")
			if len(linev) == 2 {

				if strings.Trim(linev[0], " ") == "Content-Length" {
					contentlen, err := strconv.Atoi(linev[1])
					if err != nil {
						panic("2222")
						return false, nil
					} else {
						Info("parse content length :", contentlen)
						p.r.contentdatalen = contentlen
						p.r.readedlen = 0
						p.r.readstatus = READ_CONTENTDATA
						p.r.header = nil
						return true, nil
					}
				} else {
					continue
				}
			} else {
				//panic(i)
			}
		}
		return false, nil
	}

}

func (p *HttpSocket) Read_Chunk() (bool, []byte) {

	for {
		Info("read Read_Chunk")
		if p.r.readedbuf == nil {
			Info("read chunk p.r.readedbuf == nil")
			newdata := make([]byte, READBUFF)
			n, err := p.conn.Read(newdata)
			if err != nil {
				return false, nil
			}
			p.r.readedbuf = newdata[0:n]

		}
		if p.r.header == nil {
			p.r.header = p.r.readedbuf
		} else {
			Info("p.r.header is not nil", p.r.header)
			p.r.header = append(p.r.header, p.r.readedbuf...)
		}

		s := string(p.r.header)
		head_length := strings.Index(s, "\r\n")

		if head_length == -1 {
			if len(s) > max_chunk_length {
				Info("chunk header", []byte(s[0:10]))
				return false, nil
			}
			continue
		}
		head_length += 2
		if head_length > max_chunk_length {
			return false, nil
		}
		Info("drop http chunk length :", head_length)
		p.r.readedbuf = []byte(s[head_length:])
		chunklen := Hex2Int(s[0 : head_length-2])
		Info("read chunk len ", s[0:head_length-2], []byte(s[0:head_length-2]), chunklen)
		p.r.chunkLength = chunklen
		p.r.readedlen = 0
		p.r.readstatus = READ_CHUNKDATA
		p.r.header = nil
		if p.r.chunkLength == 0 {
			return false, nil
			p.r.readstatus = READ_HEADER
		}
		return true, nil
	}

}

func (p *HttpSocket) Read_ChunkData() (bool, []byte) {

	for {

		pdata := p.r.readedbuf
		p.r.readedbuf = nil
		if len(pdata) == 0 {
			pdata = nil
		}

		if pdata == nil {
			newdata := make([]byte, READBUFF)
			n, err := p.conn.Read(newdata)
			if err != nil {
				return false, nil
			}
			pdata = newdata[0:n]
		}

		max_read := p.r.chunkLength - p.r.readedlen

		Info("read chunk data chunk readed buf", len(pdata), p.r.chunkLength, p.r.readedlen, max_read)

		if max_read > len(pdata) { //
			p.r.readedlen += len(pdata)
			Info("read chunk data len ", p.r.chunkLength, p.r.readedlen, len(pdata))
			return true, pdata
		} else if max_read <= len(pdata) {
			p.r.readedlen += max_read

			Info("read chunk data len ", p.r.chunkLength, p.r.readedlen, max_read)
			rpdata := pdata[0:max_read]

			if len(pdata) == max_read {
				p.r.readedbuf = nil
			} else {
				p.r.readedbuf = pdata[max_read:]
			}
			p.r.readstatus = READ_CHUNK
			return true, rpdata
		}
	}
}

func (p *HttpSocket) Read_Content() (bool, []byte) {
	for {

		pdata := p.r.readedbuf
		p.r.readedbuf = nil
		if len(pdata) == 0 {
			pdata = nil
		}

		if pdata == nil {
			newdata := make([]byte, READBUFF)
			n, err := p.conn.Read(newdata)
			if err != nil {
				return false, nil
			}
			pdata = newdata[0:n]
		}

		max_read := p.r.contentdatalen - p.r.readedlen

		Info("read content data content readed buf", len(pdata), p.r.contentdatalen, p.r.readedlen, max_read)

		if max_read > len(pdata) { //chunk 很大
			p.r.readedlen += len(pdata)
			Info("read content data len ", p.r.contentdatalen, p.r.readedlen, len(pdata))
			return true, pdata
		} else if max_read <= len(pdata) {
			p.r.readedlen += max_read

			Info("read content data len ", p.r.contentdatalen, p.r.readedlen, max_read)
			rpdata := pdata[0:max_read]

			if len(pdata) == max_read {
				p.r.readedbuf = nil
			} else {
				p.r.readedbuf = pdata[max_read:]
			}
			p.r.readstatus = READ_HEADER
			return true, rpdata
		}
	}

}

func (p *HttpSocket) Readn(l int) (bool, []byte) {

	pdata := p.r.X_readeddata
	p.r.X_readeddata = nil
	for {
		if pdata != nil && len(pdata) >= l {

			break
		}

		if p.r.X_readeddata == nil {
			b, pdata2 := p.Read()
			if !b {

				panic(p)
				return b, nil
			}

			if pdata == nil {
				pdata = pdata2

			} else {
				pdata = append(pdata, pdata2...)
			}

		} else {

			pdata = append(pdata, p.r.X_readeddata...)
			p.r.X_readeddata = nil

		}
		Info("Readn read once", len(pdata), l)
		if len(pdata) >= l {
			break
		}
	}

	rpdata := pdata[0:l]
	if len(pdata) > l {
		p.r.X_readeddata = pdata[l:]
	} else {
		p.r.X_readeddata = nil
	}
	return true, rpdata

}
func (p *HttpSocket) Write(data []byte, datalen int) bool {

	sendedlen := 0
	if p.w.datatypes == no_Mode {
		paddingMode := GetRandomMode()

		s, _, datalen, httpheaderlen := GetRequest(GetRandomURL(), GetRandomHost(), paddingMode)
		Debug(s)
		p.w.datatypes = paddingMode
		p.w.datalen = datalen
		Info("write padding len: ", paddingMode, datalen, Int2Hex(datalen), []byte(Int2Hex(datalen)))
		Info("write http header length : ", httpheaderlen, len(s), s, datalen, []byte(s))
		_, err := p.conn.Write([]byte(s))
		if err != nil {
			return false
		}

	}

	for {
		Debug("begin write Data", "sendedlen: ", sendedlen, " datalen: ", datalen)
		if sendedlen >= datalen {
			Debug("send success", sendedlen)
			return true
		}

		sendlen := 0

		l := p.w.datalen - p.w.sended

		if l <= 0 {
			p.w.datalen = 0
			p.w.sended = 0
			if p.w.datatypes == chunk_Mode {
				datalen2 := GetRandomChunkSize()
				Info("write chunk len: ", datalen2, Int2Hex(datalen2), []byte(Int2Hex(datalen2)))
				_, err := p.conn.Write([]byte(Int2Hex(datalen2) + "\r\n"))
				if err != nil {
					return false
				}
				p.w.datalen = datalen2
				p.w.sended = 0
			} else if p.w.datatypes == content_Mode {
				paddingMode := GetRandomMode()

				s, _, datalen, httpheaderlen := GetRequest(GetRandomURL(), GetRandomHost(), paddingMode)
				Debug(s)
				p.w.datatypes = paddingMode
				p.w.datalen = datalen
				Info("write padding len: ", paddingMode, datalen, Int2Hex(datalen), []byte(Int2Hex(datalen)))
				Info("write http header length : ", httpheaderlen, len(s), s, datalen, []byte(s))
				_, err := p.conn.Write([]byte(s))
				if err != nil {
					return false
				}

			}
		}

		l = p.w.datalen - p.w.sended

		if l > 0 {
			if datalen-sendedlen > l {
				sendlen = l
			} else {
				sendlen = datalen - sendedlen
			}
			Debug("Send ", sendlen, len(data), datalen, " bytes")
			_, err := p.conn.Write(data[sendedlen : sendedlen+sendlen])
			if err != nil {
				Debug("send failed", sendlen)
				return false
			}
			p.w.sended += sendlen
			sendedlen += sendlen
		} else {
			panic("l > 0")
		}
	}
	return true
}

func InitContent() []string {

	s := `Accept: text/html
	            Accept: image/*
	                            Accept: text/html, application/xhtml+xml, application/xml;q=0.9, */*;q=0.8
	Accept-Charset: iso-8859-1
	Accept-Charset: utf-8, iso-8859-1;q=0.5
	Accept-Language: utf-8, iso-8859-1;q=0.5, *;q=0.1
	ccept-Encoding: gzip
	Accept-Encoding: compress
	Accept-Encoding: deflate
	Accept-Encoding: br
	Accept-Encoding: identity
	Accept-Encoding: *
	Accept-Encoding: deflate, gzip;q=1.0, *;q=0.5
	Accept-Language: zh-cn
	Access-Control-Allow-Credentials: true
	Access-Control-Max-Age: 600
	Age: 2400
	Cache-Control: must-revalidate
	Cache-Control: no-cache
	Cache-Control: no-store
	Cache-Control: no-transform
	Cache-Control: public
	Cache-Control: private
	Cache-Control: proxy-revalidate
	Cache-Control: max-age=10240
	Cache-Control: s-maxage=10240
	Connection: keep-alive
	Cookie: PHPSESSID=298zf09hf012fh2; csrftoken=u32t4o3tb3gg43; _gat=1;
	Content-Type: text/html; charset=utf-8
	Date: Wed, 21 Oct 2015 07:28:00 GMT
	Expires: Wed, 21 Oct 2015 07:28:00 GMT
	If-Match: "bfc13a64729c4290ef5b2c2730249c88ca92d82d"
	Server: Apache/2.4.1 (Unix)
    `

	for i := 0; i < 100; i++ {
		s += "Cookie: PHPSESSID=" + Randhex(8) + "; csrftoken=" + Randhex(6) + "; _gat=" + Randhex(3) + ";" + "\n"
		s += "Cache-Control: max-age=" + Rand(100000) + "\n"
		s += "Access-Control-Max-Age: " + Rand(100000) + "\n"
		s += "Server: Apache/" + Rand(5) + "." + Rand(20) + "." + Rand(30) + "\n"
		s += "Server: Nginx/" + Rand(5) + "." + Rand(20) + "." + Rand(30) + "\n"
		s += "Server: IIS/" + Rand(10) + "." + Rand(20) + "." + Rand(30) + "\n"
	}
	accept := splitStr(s)
	return accept
}

func GetUrl() {

}

func GetHost() []string {
	s := `baidu.com
  sina.com
  qq.com
  163.com
  263.com
  soso.com
  taobao.com
  alipay.com
  paypal.com
  youku.com
  ibm.com
  microsoft.com
  sap.com
  dell.com
  sohu.com
  `
	v := splitStr(s)
	v2 := make([]string, 0)
	v3 := make([]string, 0)
	for _, i := range v {
		v2 = append(v2, "www."+i)
		v2 = append(v2, "login."+i)
		v2 = append(v2, "api."+i)
		v2 = append(v2, "cdn."+i)
		v2 = append(v2, "us."+i)
		v2 = append(v2, "cn."+i)
	}
	for _, i := range v2 {
		v3 = append(v3, "Host: "+i+".cn")
		v3 = append(v3, "Host: "+i)
	}
	for _, i := range v3 {
		Debug(i)
	}
	return v3
}

//var randomkey int
func GetRandomChunkSize() int {
	return getmrandCreater().Intn(15000)*getmrandCreater().Intn(15000) + 1

}

var HackGetRandomMode int

func GetRandomMode() int {

	if HackGetRandomMode == 0 {
		if getmrandCreater().Intn(15000)*getmrandCreater().Intn(15000)%2 == 0 {

			return chunk_Mode
		} else {
			return content_Mode
		}
	} else {
		return HackGetRandomMode
	}
}

func InitUserAgent() []string {

	s := `Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_8; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50
                Mozilla/5.0 (Windows; U; Windows NT 6.1; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50
                Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0;
                Mozilla/5.0 (Macintosh; Intel Mac OS X 10.6; rv:2.0.1) Gecko/20100101 Firefox/4.0.1
                Mozilla/5.0 (Windows NT 6.1; rv:2.0.1) Gecko/20100101 Firefox/4.0.1
                Opera/9.80 (Macintosh; Intel Mac OS X 10.6.8; U; en) Presto/2.8.131 Version/11.11
                Opera/9.80 (Windows NT 6.1; U; en) Presto/2.8.131 Version/11.11
                Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_0) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11
                Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; .NET4.0C; .NET4.0E)
                Opera/9.80 (Windows NT 5.1; U; zh-cn) Presto/2.9.168 Version/11.50
                Mozilla/5.0 (Windows NT 5.1; rv:5.0) Gecko/20100101 Firefox/5.0
                Mozilla/5.0 (Windows NT 5.2) AppleWebKit/534.30 (KHTML, like Gecko) Chrome/12.0.742.122 Safari/534.30
                Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/536.11 (KHTML, like Gecko) Chrome/20.0.1132.11 TaoBrowser/2.0 Safari/536.11
                Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.1 (KHTML, like Gecko) Chrome/21.0.1180.71 Safari/537.1 LBBROWSER
                Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; .NET4.0C; .NET4.0E; LBBROWSER)
                Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; Trident/4.0; SV1; QQDownload 732; .NET4.0C; .NET4.0E; 360SE)
                Mozilla/5.0 (Windows NT 5.1) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.84 Safari/535.11 SE 2.X MetaSr 1.0
                Mozilla/5.0 (Windows NT 5.1) AppleWebKit/537.1 (KHTML, like Gecko) Chrome/21.0.1180.89 Safari/537.1
                Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; Trident/4.0; SV1; QQDownload 732; .NET4.0C; .NET4.0E; SE 2.X MetaSr 1.0)
                Opera/9.27 (Windows NT 5.2; U; zh-cn)
                Opera/8.0 (Macintosh; PPC Mac OS X; U; en)
                Mozilla/5.0 (Macintosh; PPC Mac OS X; U; en) Opera 8.0
                Mozilla/5.0 (Windows; U; Windows NT 5.2) Gecko/2008070208 Firefox/3.0.1
                Mozilla/5.0 (Windows; U; Windows NT 5.1) Gecko/20070309 Firefox/2.0.0.3
                Mozilla/5.0 (Windows; U; Windows NT 5.1) Gecko/20070803 Firefox/1.5.0.12
                Mozilla/4.0 (compatible; MSIE 12.0
                Mozilla/5.0 (Windows NT 5.1; rv:44.0) Gecko/20100101 Firefox/44.0
    Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/535.11 (KHTML, like Gecko) Ubuntu/11.10 Chromium/27.0.1453.93 Chrome/27.0.1453.93 Safari/537.36
    Mozilla/5.0 (Windows NT 6.2; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/27.0.1453.94 Safari/537.36
    Mozilla/5.0 (iPhone; CPU iPhone OS 6_1_4 like Mac OS X) AppleWebKit/536.26 (KHTML, like Gecko) CriOS/27.0.1453.10 Mobile/10B350 Safari/8536.25
    Mozilla/5.0 (compatible; WOW64; MSIE 10.0; Windows NT 6.2)
    Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.0)
    Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.0; Trident/4.0)
    Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0)
    Opera/9.80 (Macintosh; Intel Mac OS X 10.6.8; U; en) Presto/2.9.168 Version/11.52
    Opera/9.80 (Windows NT 6.1; WOW64; U; en) Presto/2.10.229 Version/11.62
    Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_6; en-US) AppleWebKit/533.20.25 (KHTML, like Gecko) Version/5.0.4 Safari/533.20.27
    Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US) AppleWebKit/533.20.25 (KHTML, like Gecko) Version/5.0.4 Safari/533.20.27
    Mozilla/5.0 (iPad; CPU OS 5_0 like Mac OS X) AppleWebKit/534.46 (KHTML, like Gecko) Version/5.1 Mobile/9A334 Safari/7534.48.3
    Mozilla/5.0 (iPhone; CPU iPhone OS 5_0 like Mac OS X) AppleWebKit/534.46 (KHTML, like Gecko) Version/5.1 Mobile/9A334 Safari/7534.48.3
    Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)
    Mozilla/5.0 (iPad; CPU OS 5_0 like Mac OS X) AppleWebKit/534.46 (KHTML, like Gecko) Version/5.1 Mobile/9A334 Safari/7534.48.3
    Mozilla/5.0 (iPhone; CPU iPhone OS 5_0 like Mac OS X) AppleWebKit/534.46 (KHTML, like Gecko) Version/5.1 Mobile/9A334 Safari/7534.48.3
    Mozilla/5.0 (iPod; U; CPU like Mac OS X; en) AppleWebKit/420.1 (KHTML, like Gecko) Version/3.0 Mobile/3A101a Safari/419.3
`

	v := splitStr(s)
	agent := make([]string, 0)
	for _, i := range v {
		agent = append(agent, "User-Agent: "+i)
	}
	return agent
}

func RandomPick(t []string) string {
	index := getmrandCreater().Intn(10000000) % (len(t))
	return t[index]

}

func splitStr(s string) []string {
	v := strings.Split(s, "\n")
	v2 := make([]string, 0)
	for _, i := range v {

		i2 := strings.Replace(i, "\n", "", -1)
		i2 = strings.Replace(i, "\t", "", -1)
		i2 = strings.Trim(i2, " ")
		if len(i2) != 0 {
			v2 = append(v2, i2)
		}

	}

	for _, i := range v2 {
		Debug(i)
	}
	return v2
}

func Randhex(w int) string {

	b := make([]byte, w)

	if _, err := io.ReadFull(crand.Reader, b); err != nil {
		panic(err)
	}
	s := hex.EncodeToString(b)

	Debug(s)

	return s
}

func Int2Hex(i int) string {

	s16 := strconv.FormatInt(int64(i), 16)
	return s16
}

func Hex2Int(s64 string) int {

	if s, err := strconv.ParseInt(s64, 16, 64); err == nil {
		Debug(s)
		return int(s)

	} else {
		return 0
	}

}

func RandHex(a string) {

	if s, err := strconv.ParseInt(a, 16, 64); err == nil {
		Debug("%T, %v\n", s, s)
	}

}

func GetRandomHost() string {
	x := getmrandCreater().Intn(15)
	if x%2 == 1 {
		return RandomPick(GetHost())

	} else {
		x := getmrandCreater().Intn(10) + 1
		s := "Host: www."
		v := make([]byte, x)
		for i := 0; i < x; i++ {
			v[i] = byte((getmrandCreater().Intn(100000) % 26) + 'a')
		}
		return s + string(v[0:x]) + ".com"
	}
}

func GetRandomURL() string {
	s := "GET / HTTP/1.1"
	return s

}

func GetRequest(url string, host string, idatatype int) (request string, datatype int, datalen int, httpheaderlen int) {

	v := make([]string, 0)

	header_agent := RandomPick(InitUserAgent())

	v = append(v, header_agent)
	header_host := host

	v = append(v, header_host)
	x := getmrandCreater().Intn(15) + 1
	for i := 0; i < x; i++ {
		header_agent := RandomPick(InitContent())
		v = append(v, header_agent)
	}
	datalen = GetRandomChunkSize()
	datatype = idatatype

	if datatype == content_Mode {
		v = append(v, "Content-Length :"+strconv.Itoa(datalen))
	} else if datatype == chunk_Mode {
		v = append(v, "Transfer-Encoding: chunked")
	} else {
		panic("wrong datatype")
	}
	list := mrand.Perm(len(v))
	request += url + "\r\n"
	for i, _ := range list {
		request += v[list[i]] + "\r\n"
	}
	request += "\r\n"
	httpheaderlen = len(request)
	Info("http header length :", len(request))
	if datatype == chunk_Mode {
		request += Int2Hex(datalen) + "\r\n"
	}
	Debug(request)
	return request, datatype, datalen, httpheaderlen
}

func Init() {

	GetHost()
	InitContent()
	InitUserAgent()

}

func Rand(s int) string {

	l := getmrandCreater().Intn(s)
	return strconv.Itoa(l)
}

func RandInt(s int) int {
	return getmrandCreater().Intn(s)
}

func (p *HttpSocket) LocalAddr() net.Addr {
	return p.conn.LocalAddr()
}

func (p *HttpSocket) RemoteAddr() net.Addr {
	return p.conn.RemoteAddr()
}
