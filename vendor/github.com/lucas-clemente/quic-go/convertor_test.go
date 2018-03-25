package quic

import (
	"fmt"
	"testing"
	crand "crypto/rand"
	"io"
)

func TestConv(t *testing.T) {
	c := &Convertor{}
	conf := &Config{}
	conf.AESIv[0] = 1
	c.Init(conf)

	for x := 0; x < 10; x++ {
		for i := 1; i < 1497; i++ {
			testCase(c, i, t)
		}
	}
}

func testCase(c *Convertor, l int, t *testing.T) {
	data := make([]byte, l)
	if _, err := io.ReadFull(crand.Reader, data); err != nil {
		panic(err)
	}
	fmt.Println(data)
	p2 := c.MixData(data)
	fmt.Println(len(p2), p2)
	p3 := c.UnMixData(p2, len(p2))

	if p3 != l {
		fmt.Println("l:", l, "p3:", p3)
		t.Fail() 
	}
	for i := 0; i < l; i++ {
		if data[i] != p2[i] {
			t.Fail()
		}

	}

}
