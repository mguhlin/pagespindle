//go:build js && wasm

package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/mguhlin/pagespindle/internal/book"
	"strings"
	"syscall/js"
)

type request struct {
	Operation string        `json:"operation"`
	Text      string        `json:"text"`
	Title     string        `json:"title"`
	Data      string        `json:"data"`
	Kind      string        `json:"kind"`
	Document  book.Document `json:"document"`
}

func main() {
	js.Global().Set("pagespindleCore", js.FuncOf(func(this js.Value, args []js.Value) any {
		result := map[string]any{}
		var req request
		var e error
		if len(args) != 1 || len(args[0].String()) > 24<<20 {
			e = errors.New("Browser core request exceeds its limit")
		} else {
			e = json.Unmarshal([]byte(args[0].String()), &req)
		}
		if e == nil {
			switch req.Operation {
			case "new":
				var d book.Document
				d, e = book.New(req.Text, req.Title)
				result["document"] = d
			case "import":
				var data []byte
				data, e = base64.StdEncoding.DecodeString(req.Data)
				if e == nil && len(data) > book.MaxInput {
					e = errors.New("File exceeds 16 MiB")
				}
				if e == nil {
					var d book.Document
					var text string
					switch req.Kind {
					case ".json":
						d, e = book.Load(data)
					case ".epub":
						text, e = book.EPUB(data)
					case ".pdf":
						text, e = book.PDF(data)
					case ".txt", ".md":
						text = string(data)
					default:
						e = errors.New("Unsupported file type")
					}
					if e == nil && req.Kind != ".json" {
						d, e = book.New(text, req.Title)
					}
					result["document"] = d
				}
			case "ready":
				e = book.Ready(req.Document)
				result["text"] = book.Master(req.Document)
			case "master":
				e = book.Validate(req.Document)
				result["text"] = book.Master(req.Document)
			case "chapters":
				e = book.Ready(req.Document)
				result["chapters"] = book.Chapters(book.Master(req.Document))
			case "validate":
				e = book.Validate(req.Document)
			default:
				e = errors.New("Unknown browser core operation")
			}
		}
		if e != nil {
			result = map[string]any{"error": strings.TrimSpace(e.Error())}
		}
		b, _ := json.Marshal(result)
		return string(b)
	}))
	select {}
}
