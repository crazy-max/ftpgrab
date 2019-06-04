package utl

import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
)

var (
	gbk      = simplifiedchinese.GBK.NewDecoder()
	gbk18030 = simplifiedchinese.GB18030.NewDecoder()
	hzgb2312 = simplifiedchinese.HZGB2312.NewDecoder()
)

func Encoding(encoding string, value string) string {
	switch encoding {
	case "GBK":
		return encodingValue(gbk, value)
	case "GB18030":
		return encodingValue(gbk18030, value)
	case "HZGB2312":
		return encodingValue(hzgb2312, value)
	default:
		return value
	}
}

func encodingValue(enc *encoding.Decoder, value string) string {
	encValue, err := enc.String(value)
	if err != nil {
		return value
	}
	return encValue
}
