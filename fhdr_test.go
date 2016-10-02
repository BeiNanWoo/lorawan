package lorawan

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDevAddr(t *testing.T) {
	Convey("Given an empty DevAddr", t, func() {
		var a DevAddr
		Convey("Then MarshalBinary returns ", func() {
			b, err := a.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0, 0, 0, 0})
		})

		Convey("Given the DevAddr{255, 1, 1, 1}", func() {
			a = DevAddr{255, 1, 1, 1}
			Convey("Then NwkID returns byte(127)", func() {
				So(a.NwkID(), ShouldEqual, byte(127))
			})
		})

		Convey("Given the DevAddr{1, 2, 3, 4}", func() {
			a = DevAddr{1, 2, 3, 4}
			Convey("Then MarshalBinary returns []byte{4, 3, 2, 1}", func() {
				b, err := a.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{4, 3, 2, 1})
			})

			Convey("Then MarshalText returns 01020304", func() {
				b, err := a.MarshalText()
				So(err, ShouldBeNil)
				So(string(b), ShouldEqual, "01020304")
			})
		})

		Convey("Given the slice []byte{4, 3, 2, 1}", func() {
			b := []byte{4, 3, 2, 1}
			Convey("Then UnmarshalBinary returns DevAddr{1, 2, 3, 4}", func() {
				err := a.UnmarshalBinary(b)
				So(err, ShouldBeNil)
				So(a, ShouldResemble, DevAddr{1, 2, 3, 4})
			})
		})

		Convey("Given the string 01020304", func() {
			str := "01020304"
			Convey("Then UnmarshalText returns DevAddr{1, 2, 3, 4}", func() {
				err := a.UnmarshalText([]byte(str))
				So(err, ShouldBeNil)
				So(a, ShouldResemble, DevAddr{1, 2, 3, 4})
			})
		})

		Convey("Given []byte{1, 2, 3, 4}", func() {
			b := []byte{1, 2, 3, 4}
			Convey("Then Scan scans the value correctly", func() {
				So(a.Scan(b), ShouldBeNil)
				So(a[:], ShouldResemble, b)
			})
		})
	})
}

func TestFCtrl(t *testing.T) {
	Convey("Given an empty FCtrl", t, func() {
		var c FCtrl
		Convey("Then MarshalBinary returns []byte{0}", func() {
			b, err := c.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0})
		})

		Convey("Given FOptsLen > 15", func() {
			c.fOptsLen = 16
			Convey("Then MarshalBinary returns an error", func() {
				_, err := c.MarshalBinary()
				So(err, ShouldNotBeNil)
			})
		})

		testTable := []struct {
			ADR       bool
			ADRACKReq bool
			ACK       bool
			FPending  bool
			FOptsLen  uint8
			Bytes     []byte
		}{
			{true, false, false, false, 2, []byte{130}},
			{false, true, false, false, 3, []byte{67}},
			{false, false, true, false, 4, []byte{36}},
			{false, false, false, true, 5, []byte{21}},
			{true, true, true, true, 6, []byte{246}},
		}

		for _, test := range testTable {
			Convey(fmt.Sprintf("Given ADR=%v, ADRACKReq=%v, ACK=%v, FPending=%v, fOptsLen=%d", test.ADR, test.ADRACKReq, test.ACK, test.FPending, test.FOptsLen), func() {
				c.ADR = test.ADR
				c.ADRACKReq = test.ADRACKReq
				c.ACK = test.ACK
				c.FPending = test.FPending
				c.fOptsLen = test.FOptsLen
				Convey(fmt.Sprintf("Then MarshalBinary returns %v", test.Bytes), func() {
					b, err := c.MarshalBinary()
					So(err, ShouldBeNil)
					So(b, ShouldResemble, test.Bytes)
				})
			})

			Convey(fmt.Sprintf("Given the slice %v", test.Bytes), func() {
				b := test.Bytes
				Convey(fmt.Sprintf("Then UnmarshalBinary returns a FCtrl with ADR=%v, ADRACKReq=%v, ACK=%v, FPending=%v, fOptsLen=%d", test.ADR, test.ADRACKReq, test.ACK, test.FPending, test.FOptsLen), func() {
					err := c.UnmarshalBinary(b)
					So(err, ShouldBeNil)
					So(c, ShouldResemble, FCtrl{ADR: test.ADR, ADRACKReq: test.ADRACKReq, ACK: test.ACK, FPending: test.FPending, fOptsLen: test.FOptsLen})
				})
			})
		}
	})
}

func TestFHDR(t *testing.T) {
	Convey("Given an empty FHDR", t, func() {
		KeepUnknownMACCommandRemainder = false
		var h FHDR

		Convey("Then MarshalBinary returns []byte{0, 0, 0, 0, 0, 0, 0}", func() {
			b, err := h.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0, 0, 0, 0, 0, 0, 0})
		})

		Convey("Given the FCnt contains a value > 16 bits", func() {
			h.FCnt = 65795

			Convey("Then only the least-significant 16 bits are marshalled", func() {
				b, err := h.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{0, 0, 0, 0, 0, 3, 1})
			})
		})

		Convey("Given DevAddr=67305985, FCtrl=FCtrl(ADR=true, ADRACKReq=false, ACK=true, FPending=true), Fcnt=5, FOpts=[]MACCommand{(CID=LinkCheckAns, Payload=LinkCheckAnsPayload(Margin=7, GwCnt=9))}", func() {
			h.DevAddr = DevAddr([4]byte{1, 2, 3, 4})
			h.FCtrl = FCtrl{ADR: true, ADRACKReq: false, ACK: true, FPending: true}
			h.FCnt = 5
			h.FOpts = []MACCommand{
				{CID: LinkCheckAns, Payload: &LinkCheckAnsPayload{Margin: 7, GwCnt: 9}},
			}
			Convey("Then MarshalBinary returns []byte{4, 3, 2, 1, 179, 5, 0, 2, 7, 9}", func() {
				b, err := h.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{4, 3, 2, 1, 179, 5, 0, 2, 7, 9})
			})
		})

		Convey("Given FOpts contains 5 times MACCommand{(CID=LinkCheckAns, Payload=LinkCheckAnsPayload(Margin=7, GwCnt=9))}", func() {
			for i := 0; i < 5; i++ {
				h.FOpts = append(h.FOpts, MACCommand{CID: LinkCheckAns, Payload: &LinkCheckAnsPayload{Margin: 7, GwCnt: 9}})
			}
			Convey("Then MarshalBinary does not return an error", func() {
				_, err := h.MarshalBinary()
				So(err, ShouldBeNil)
			})
		})

		Convey("Given FOpts contains 6 times MACCommand{(CID=LinkCheckAns, Payload=LinkCheckAnsPayload(Margin=7, GwCnt=9))}", func() {
			for i := 0; i < 6; i++ {
				h.FOpts = append(h.FOpts, MACCommand{CID: LinkCheckAns, Payload: &LinkCheckAnsPayload{Margin: 7, GwCnt: 9}})
			}
			Convey("Then MarshalBinary does return an error", func() {
				_, err := h.MarshalBinary()
				So(err, ShouldResemble, errors.New("lorawan: max number of FOpts bytes is 15"))
			})
		})

		Convey("Given uplink=false and slice []byte{4, 3, 2, 1, 179, 5, 0, 2, 7, 9}", func() {
			b := []byte{4, 3, 2, 1, 179, 5, 0, 2, 7, 9}
			Convey("Then UnmarshalBinary does not return an error", func() {
				err := h.UnmarshalBinary(false, b)
				So(err, ShouldBeNil)

				Convey("Then DevAddr=[4]{1, 2, 3, 4}", func() {
					So(h.DevAddr, ShouldEqual, DevAddr([4]byte{1, 2, 3, 4}))
				})

				Convey("Then FCtrl=FCtrl(ADR=true, ADRACKReq=false, ACK=true, FPending=true, fOptsLen=3)", func() {
					So(h.FCtrl, ShouldResemble, FCtrl{ADR: true, ADRACKReq: false, ACK: true, FPending: true, fOptsLen: 3})
				})

				Convey("Then len(FOpts)=1", func() {
					So(h.FOpts, ShouldHaveLength, 1)
					Convey("Then CID=LinkCheckAns", func() {
						So(h.FOpts[0].CID, ShouldEqual, LinkCheckAns)
					})

				})

				Convey("Then Payload=LinkCheckAnsPayload(Margin=7, GwCnt=9)", func() {
					p, ok := h.FOpts[0].Payload.(*LinkCheckAnsPayload)
					So(ok, ShouldBeTrue)
					So(p, ShouldResemble, &LinkCheckAnsPayload{Margin: 7, GwCnt: 9})
				})
			})
		})

		Convey("Given uplink=false and slice []byte{4, 3, 2, 1, 181, 5, 0, 2, 7, 9, 78, 79} (one known mac-command and some unknown data)", func() {
			b := []byte{4, 3, 2, 1, 181, 5, 0, 2, 7, 9, 78, 79}
			var logBytes bytes.Buffer
			log.SetOutput(&logBytes)

			Convey("Then UnmarshalBinary does not return an error", func() {
				err := h.UnmarshalBinary(false, b)
				So(err, ShouldBeNil)

				Convey("Then DevAddr=[4]{1, 2, 3, 4}", func() {
					So(h.DevAddr, ShouldEqual, DevAddr([4]byte{1, 2, 3, 4}))
				})

				Convey("Then FCtrl=FCtrl(ADR=true, ADRACKReq=false, ACK=true, FPending=true, fOptsLen=5)", func() {
					So(h.FCtrl, ShouldResemble, FCtrl{ADR: true, ADRACKReq: false, ACK: true, FPending: true, fOptsLen: 5})
				})

				Convey("Then len(FOpts)=1", func() {
					So(h.FOpts, ShouldHaveLength, 1)
					Convey("Then CID=LinkCheckAns", func() {
						So(h.FOpts[0].CID, ShouldEqual, LinkCheckAns)
					})

				})

				Convey("Then Payload=LinkCheckAnsPayload(Margin=7, GwCnt=9)", func() {
					p, ok := h.FOpts[0].Payload.(*LinkCheckAnsPayload)
					So(ok, ShouldBeTrue)
					So(p, ShouldResemble, &LinkCheckAnsPayload{Margin: 7, GwCnt: 9})
				})

				Convey("Then a warning was printed", func() {
					So(logBytes.String(), ShouldEndWith, "warning: unmarshal mac-command error (skipping remaining mac-command bytes): lorawan: invalid CID=4e for uplink=false\n")
				})
			})

			Convey("Given KeepUnknownMACCommandRemainder=true", func() {
				KeepUnknownMACCommandRemainder = true

				Convey("Then UnmarshalBinary does not return an error", func() {
					err := h.UnmarshalBinary(false, b)
					So(err, ShouldBeNil)

					Convey("Then len(Fopts)=2", func() {
						So(h.FOpts, ShouldHaveLength, 2)
					})

					Convey("Then the second MACCommand contains the UnknownMACCommandRemainder", func() {
						pl := UnknownMACCommandRemainder{79}
						So(h.FOpts[1], ShouldResemble, MACCommand{
							CID:     CID(78),
							Payload: &pl,
						})
					})
				})
			})
		})

		Convey("Given uplink=false and slice []byte{1, 2, 3, 4, 179, 5, 0, 2, 7}", func() {
			b := []byte{1, 2, 3, 4, 179, 5, 0, 2, 7}
			Convey("Then UnmarshalBinary returns an error", func() {
				err := h.UnmarshalBinary(false, b)
				So(err, ShouldResemble, errors.New("lorawan: not enough remaining bytes"))
			})
		})

		Convey("Given FOpts with a MACCommand with non-empty NewChannelReqPayload", func() {
			m := MACCommand{
				CID: NewChannelReq,
				Payload: &NewChannelReqPayload{
					ChIndex: 2,
					Freq:    1234567,
					MaxDR:   5,
					MinDR:   1,
				},
			}
			h.FOpts = []MACCommand{m}
			Convey("When it is transformed into binary", func() {
				b, err := h.MarshalBinary()
				So(err, ShouldBeNil)
				Convey("Then it can be converted back to the original payload", func() {
					actual := FHDR{}
					So(actual.UnmarshalBinary(false, b), ShouldBeNil)
					So(actual.FOpts, ShouldResemble, []MACCommand{m})
				})
			})
		})
	})
}
