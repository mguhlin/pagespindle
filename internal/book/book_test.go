package book

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestSourceAndReviewRoundtrip(t *testing.T) {
	source := "# Test\r\n\r\nA well-known café opens at 9:00. [12] x² remains.\r\n\r\nA unique sidebar."
	d, e := New(source, "../Test")
	if e != nil {
		t.Fatal(e)
	}
	if d.Source != source {
		t.Fatal("original bytes changed")
	}
	if Ready(d) == nil {
		t.Fatal("unreviewed narration accepted")
	}
	for i := range d.Blocks {
		d.Blocks[i].Reviewed = true
	}
	d.Blocks[2].Destination = "notes"
	b, _ := json.Marshal(d)
	copy, e := Load(b)
	if e != nil {
		t.Fatal(e)
	}
	if copy.Source != source || !strings.Contains(Master(copy), "# Notes\n\n## Test\n\nA unique sidebar.") {
		t.Fatal("source/notes lost")
	}
	copy.Blocks[1].Source = "substituted"
	if Validate(copy) == nil {
		t.Fatal("tampered source accepted")
	}
	copy = d
	copy.Blocks = append([]Block(nil), d.Blocks...)
	copy.Blocks[1].Destination = "omit"
	if Ready(copy) == nil {
		t.Fatal("silent omission accepted")
	}
	if strings.Contains(SafeName("../CON / title"), "/") {
		t.Fatal("unsafe filename")
	}
}
func TestLongUnicodePassagesAccountedFor(t *testing.T) {
	text := strings.Repeat("猫 café well-known 12. ", 700)
	d, e := New(text, "Test")
	if e != nil {
		t.Fatal(e)
	}
	var words []string
	for _, b := range d.Blocks {
		words = append(words, strings.Fields(b.Source)...)
	}
	if strings.Join(words, " ") != strings.Join(strings.Fields(text), " ") {
		t.Fatal("passage split lost text")
	}
}
func epub(entries map[string]string) []byte {
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	for n, s := range entries {
		w, _ := z.Create(n)
		w.Write([]byte(s))
	}
	z.Close()
	return b.Bytes()
}
func TestEPUBUsesSpineAndRejectsTraversal(t *testing.T) {
	files := map[string]string{"META-INF/container.xml": `<container><rootfiles><rootfile full-path="OEBPS/book.opf"/></rootfiles></container>`, "OEBPS/book.opf": `<package><manifest><item id="a" href="a.xhtml"/><item id="b" href="b.xhtml"/></manifest><spine><itemref idref="b"/><itemref idref="a"/></spine></package>`, "OEBPS/a.xhtml": `<html><body><p>Last chapter. [1]</p><img alt="Map supplied by author"/></body></html>`, "OEBPS/b.xhtml": `<html><head><style>Not narrative</style></head><body><h1>First chapter</h1><p>Keep 12 and x².</p></body></html>`}
	text, e := EPUB(epub(files))
	if e != nil {
		t.Fatal(e)
	}
	if strings.Index(text, "First chapter") > strings.Index(text, "Last chapter") || !strings.Contains(text, "Map supplied by author") || strings.Contains(text, "Not narrative") {
		t.Fatal("reading order/content failure")
	}
	files["../outside"] = "bad"
	if _, e := EPUB(epub(files)); e == nil {
		t.Fatal("traversal accepted")
	}
}
func TestProjectRejectsExtraFieldsAndInvalidTypes(t *testing.T) {
	d, _ := New("Keep original.", "Test")
	b, _ := json.Marshal(d)
	b = bytes.Replace(b, []byte(`"schema":1`), []byte(`"schema":1,"command":"execute"`), 1)
	if _, e := Load(b); e == nil {
		t.Fatal("unknown authority field accepted")
	}
	if _, e := New("\x00bad", "Test"); e == nil {
		t.Fatal("null bytes accepted")
	}
	if _, e := New(strings.Repeat("x", MaxText+1), "Test"); e == nil {
		t.Fatal("oversized text accepted")
	}
}
