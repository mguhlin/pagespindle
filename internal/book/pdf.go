package book

import (
	"bytes"
	"errors"
	"github.com/ledongthuc/pdf"
	"io"
)

func PDF(data []byte) (text string, err error) {
	defer func() {
		if recover() != nil {
			err = errors.New("This PDF could not be safely extracted")
		}
	}()
	r, e := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if e != nil {
		return "", e
	}
	if r.NumPage() > 1500 {
		return "", errors.New("PDF exceeds 1500 pages; import a smaller section")
	}
	plain, e := r.GetPlainText()
	if e != nil {
		return "", e
	}
	b, e := io.ReadAll(io.LimitReader(plain, MaxText+1))
	if e != nil || len(b) > MaxText {
		return "", errors.New("PDF text exceeds the import limit")
	}
	return string(b), nil
}
