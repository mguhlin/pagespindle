package book

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
)

const MaxInput = 16 << 20
const MaxText = 2 << 20
const MaxBlocks = 12000

type Block struct {
	ID          int    `json:"id"`
	Source      string `json:"source"`
	Prepared    string `json:"prepared"`
	Reason      string `json:"reason"`
	Reviewed    bool   `json:"reviewed"`
	Destination string `json:"destination"`
}
type Document struct {
	Schema int     `json:"schema"`
	Title  string  `json:"title"`
	Author string  `json:"author"`
	Source string  `json:"source"`
	Digest string  `json:"digest"`
	Blocks []Block `json:"blocks"`
}

func New(text, title string) (Document, error) {
	if !utf8.ValidString(text) || len(text) > MaxText || strings.IndexByte(text, 0) >= 0 {
		return Document{}, errors.New("Use valid UTF-8 text up to 2 MiB without null characters")
	}
	if strings.TrimSpace(text) == "" {
		return Document{}, errors.New("Add some readable text first")
	}
	d := Document{Schema: 1, Title: CleanTitle(title), Source: text}
	sum := sha256.Sum256([]byte(text))
	d.Digest = hex.EncodeToString(sum[:])
	// Only the working passage view normalizes line endings. The source is retained byte-for-byte.
	paragraphs := regexp.MustCompile(`\n[\t ]*\n`).Split(strings.ReplaceAll(text, "\r\n", "\n"), -1)
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		for len([]rune(p)) > 1400 {
			r := []rune(p)
			cut := 1400
			for i := 1400; i > 700; i-- {
				if r[i] == ' ' || r[i] == '\n' {
					cut = i
					break
				}
			}
			d.Blocks = append(d.Blocks, Block{ID: len(d.Blocks) + 1, Source: string(r[:cut]), Prepared: string(r[:cut])})
			p = strings.TrimSpace(string(r[cut:]))
		}
		if p != "" {
			d.Blocks = append(d.Blocks, Block{ID: len(d.Blocks) + 1, Source: p, Prepared: p})
		}
	}
	if len(d.Blocks) > MaxBlocks {
		return Document{}, errors.New("Too many passages; split this text into smaller projects")
	}
	return d, nil
}
func CleanTitle(s string) string {
	s = strings.TrimSpace(s)
	if len([]rune(s)) > 120 {
		s = string([]rune(s)[:120])
	}
	if s == "" {
		return "Listening Edition"
	}
	return s
}
func SafeName(s string) string {
	s = regexp.MustCompile(`[^\pL\pN _-]`).ReplaceAllString(CleanTitle(s), "-")
	s = strings.Trim(s, " .-")
	if s == "" {
		s = "Listening-Edition"
	}
	reserved := strings.ToUpper(s)
	if reserved == "CON" || reserved == "PRN" || reserved == "AUX" || reserved == "NUL" || regexp.MustCompile(`^(COM|LPT)[1-9]$`).MatchString(reserved) {
		s = "Book-" + s
	}
	return s
}
func Validate(d Document) error {
	if len([]rune(d.Title)) > 120 || len([]rune(d.Author)) > 200 {
		return errors.New("Title or author is too long")
	}
	if d.Schema != 1 {
		return errors.New("Unsupported project format")
	}
	orig, err := New(d.Source, d.Title)
	if err != nil {
		return err
	}
	if d.Digest != orig.Digest || len(d.Blocks) != len(orig.Blocks) {
		return errors.New("Project source or passage inventory has changed")
	}
	total := 0
	for i, b := range d.Blocks {
		if b.ID != orig.Blocks[i].ID || b.Source != orig.Blocks[i].Source {
			return errors.New("Project passage does not match the preserved source")
		}
		if !utf8.ValidString(b.Prepared) || strings.IndexByte(b.Prepared, 0) >= 0 || len(b.Prepared) > 12000 || len(b.Reason) > 2000 {
			return errors.New("Invalid prepared passage")
		}
		if b.Destination != "" && b.Destination != "body" && b.Destination != "notes" && b.Destination != "omit" {
			return errors.New("Invalid passage destination")
		}
		total += len(b.Prepared)
	}
	if total > MaxText*2 {
		return errors.New("Prepared text exceeds the output limit")
	}
	return nil
}
func Load(data []byte) (Document, error) {
	var d Document
	if len(data) > MaxInput {
		return d, errors.New("Project file is too large")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&d); err != nil {
		return d, errors.New("This is not a valid PageSpindle project")
	}
	if dec.Decode(&struct{}{}) != io.EOF {
		return d, errors.New("Unexpected trailing project content")
	}
	return d, Validate(d)
}
func Master(d Document) string {
	var body []string
	type group struct {
		title string
		notes []string
	}
	var groups []group
	chapter := "Source notes"
	for _, p := range d.Blocks {
		first := strings.TrimSpace(strings.SplitN(p.Source, "\n", 2)[0])
		if strings.HasPrefix(first, "# ") {
			chapter = strings.TrimSpace(strings.TrimPrefix(first, "# "))
		}
		if p.Destination == "omit" || strings.TrimSpace(p.Prepared) == "" {
			continue
		}
		if p.Destination == "notes" {
			if len(groups) == 0 || groups[len(groups)-1].title != chapter {
				groups = append(groups, group{title: chapter})
			}
			g := &groups[len(groups)-1]
			g.notes = append(g.notes, p.Prepared)
		} else {
			body = append(body, p.Prepared)
		}
	}
	if len(groups) > 0 {
		var notes []string
		for _, g := range groups {
			notes = append(notes, "## "+g.title+"\n\n"+strings.Join(g.notes, "\n\n"))
		}
		body = append(body, "# Notes\n\n"+strings.Join(notes, "\n\n"))
	}
	return strings.TrimSpace(strings.Join(body, "\n\n")) + "\n"
}
func Ready(d Document) error {
	if err := Validate(d); err != nil {
		return err
	}
	for _, b := range d.Blocks {
		if b.Destination == "omit" && strings.TrimSpace(b.Reason) == "" {
			return fmt.Errorf("Record a reason for omitting passage %d", b.ID)
		}
	}
	if strings.TrimSpace(Master(d)) == "" {
		return errors.New("The narration text is empty")
	}
	return nil
}

type Chapter struct {
	Title string
	Text  string
}

func Chapters(text string) []Chapter {
	var out []Chapter
	title := "Listening Edition"
	var body strings.Builder
	flush := func() {
		if strings.TrimSpace(body.String()) != "" {
			out = append(out, Chapter{title, strings.TrimSpace(body.String())})
			body.Reset()
		}
	}
	for _, line := range strings.Split(text, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "# ") {
			flush()
			title = strings.TrimSpace(strings.TrimPrefix(trim, "# "))
			body.WriteString(title + ".\n\n")
		} else {
			body.WriteString(line + "\n")
		}
	}
	flush()
	return out
}

func EPUB(data []byte) (string, error) {
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", errors.New("This EPUB archive could not be opened")
	}
	if len(z.File) > 2000 {
		return "", errors.New("EPUB has too many entries")
	}
	files := map[string]*zip.File{}
	var total uint64
	for _, f := range z.File {
		if strings.Contains(f.Name, "\\") || strings.HasPrefix(f.Name, "/") || path.Clean(f.Name) != strings.TrimSuffix(f.Name, "/") || strings.HasPrefix(f.Name, "../") {
			return "", errors.New("Unsafe EPUB entry")
		}
		if files[f.Name] != nil {
			return "", errors.New("Duplicate EPUB entry")
		}
		total += f.UncompressedSize64
		if total > 64<<20 || f.UncompressedSize64 > 16<<20 {
			return "", errors.New("Expanded EPUB exceeds the import limit")
		}
		files[f.Name] = f
	}
	read := func(name string) ([]byte, error) {
		f := files[name]
		if f == nil {
			return nil, fmt.Errorf("EPUB is missing %s", name)
		}
		r, e := f.Open()
		if e != nil {
			return nil, e
		}
		defer r.Close()
		b, e := io.ReadAll(io.LimitReader(r, (16<<20)+1))
		if len(b) > 16<<20 {
			return nil, errors.New("EPUB entry exceeds the import limit")
		}
		return b, e
	}
	if files["META-INF/encryption.xml"] != nil {
		return "", errors.New("Encrypted or obfuscated EPUB resources need a separate authorized conversion; DRM removal is not supported")
	}
	raw, err := read("META-INF/container.xml")
	if err != nil {
		return "", err
	}
	var c struct {
		Rootfiles []struct {
			Path string `xml:"full-path,attr"`
		} `xml:"rootfiles>rootfile"`
	}
	if xml.Unmarshal(raw, &c) != nil || len(c.Rootfiles) == 0 {
		return "", errors.New("EPUB package location is invalid")
	}
	opf := path.Clean(c.Rootfiles[0].Path)
	raw, err = read(opf)
	if err != nil {
		return "", err
	}
	var pkg struct {
		Items []struct {
			ID   string `xml:"id,attr"`
			Href string `xml:"href,attr"`
		} `xml:"manifest>item"`
		Spine []struct {
			ID string `xml:"idref,attr"`
		} `xml:"spine>itemref"`
	}
	if xml.Unmarshal(raw, &pkg) != nil {
		return "", errors.New("EPUB reading order is invalid")
	}
	manifest := map[string]string{}
	for _, i := range pkg.Items {
		manifest[i.ID] = i.Href
	}
	var out strings.Builder
	for _, s := range pkg.Spine {
		href, ok := manifest[s.ID]
		if !ok || strings.Contains(href, ":") || strings.Contains(href, "?") {
			return "", errors.New("EPUB reading order references an unsupported resource")
		}
		name := path.Join(path.Dir(opf), strings.Split(href, "#")[0])
		raw, err = read(name)
		if err != nil {
			return "", err
		}
		text, err := HTML(raw)
		if err != nil {
			return "", err
		}
		out.WriteString(text)
		out.WriteString("\n\n")
		if out.Len() > MaxText {
			return "", errors.New("EPUB text exceeds 2 MiB; import smaller sections")
		}
	}
	if strings.TrimSpace(out.String()) == "" {
		return "", errors.New("No readable EPUB text was found")
	}
	return out.String(), nil
}
func HTML(raw []byte) (string, error) {
	n, e := html.Parse(bytes.NewReader(raw))
	if e != nil {
		return "", e
	}
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style" || n.Data == "head") {
			return
		}
		block := n.Type == html.ElementNode && strings.Contains(" p div section article li h1 h2 h3 h4 blockquote br ", " "+n.Data+" ")
		if block {
			b.WriteString("\n\n")
		}
		if n.Type == html.ElementNode && (n.Data == "h1" || n.Data == "h2") {
			b.WriteString("# ")
		}
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		if n.Type == html.ElementNode && n.Data == "img" {
			for _, a := range n.Attr {
				if a.Key == "alt" && a.Val != "" {
					b.WriteString("\nImage alternative text: " + a.Val + "\n")
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
		if block {
			b.WriteString("\n\n")
		}
	}
	walk(n)
	return strings.TrimSpace(b.String()), nil
}
