package book

import (
	"encoding/json"
	"strings"
	"testing"
)

func speechWords(chapters []Chapter) string {
	var text []string
	for _, ch := range chapters {
		text = append(text, ch.Text)
	}
	return strings.Join(strings.Fields(strings.Join(text, "\n")), " ")
}

func TestNarrationReadsMarkdownContent(t *testing.T) {
	master := "# **First chapter** #\n\n## A *quiet* heading\n\nRead **every substantive word**, _including_ ***nested emphasis***.\n\n" +
		"> A quoted line.\n\n- First item\n- Second item\n\nVisit [the library](https://example.invalid/a_(b) \"Unused title\") or [its guide][guide].\n\n" +
		"![Map supplied by the author](https://example.invalid/map.png)\n\n[guide]: https://example.invalid/guide\n\n---\n\n" +
		"# Second chapter\n\n### Details\n\n#### More\n\n##### Still here\n\n###### Last heading\n\n~~Retain these words~~. [12]"
	chapters := Chapters(master)
	if len(chapters) != 2 || chapters[0].Title != "First chapter" || chapters[1].Title != "Second chapter" {
		t.Fatalf("chapter structure changed: %#v", chapters)
	}
	want := "First chapter. A quiet heading Read every substantive word, including nested emphasis. A quoted line. First item Second item Visit the library or its guide. Map supplied by the author Second chapter. Details More Still here Last heading Retain these words. [12]"
	if got := speechWords(chapters); got != want {
		t.Fatalf("speech content\n got %q\nwant %q", got, want)
	}
}

func TestNarrationPreservesLiteralSymbolsAndCode(t *testing.T) {
	master := "C# and #learning, issue #7, 2 * 3 = 6, well-known, file_name, 9:00, x² and [12].\n\n" +
		"Escaped \\*literal\\* and \\# symbol. `# code * literal` and &amp; caf&eacute;.\n\n" +
		"```text\n# not a chapter\nvalue = 2 * 3\n```\n\n<https://example.invalid/path>"
	chapters := Chapters(master)
	if len(chapters) != 1 {
		t.Fatal("a heading inside a code fence became a chapter")
	}
	want := "C# and #learning, issue #7, 2 * 3 = 6, well-known, file_name, 9:00, x² and [12]. Escaped *literal* and # symbol. # code * literal and & café. # not a chapter value = 2 * 3 https://example.invalid/path"
	if got := speechWords(chapters); got != want {
		t.Fatalf("literal content changed: %q", got)
	}
}

func TestNarrationListsTablesAndSetext(t *testing.T) {
	master := "**A chapter**\n===========\n\nA subheading\n------------\n\n3. Third step\n4. Fourth step\n\n" +
		"- [x] Finished\n- [ ] Pending\n\n| Name | Count |\n| :--- | ---: |\n| Café | **12** |"
	chapters := Chapters(master)
	if len(chapters) != 1 || chapters[0].Title != "A chapter" {
		t.Fatal("setext heading was not recognized")
	}
	want := "A chapter. A subheading 3. Third step 4. Fourth step Completed: Finished Not completed: Pending Name; Count Café; 12"
	if got := speechWords(chapters); got != want {
		t.Fatalf("list/table content changed: %q", got)
	}
}

func TestNarrationParsesBeforePassageAndPreviewSplits(t *testing.T) {
	source := "## **" + strings.TrimSpace(strings.Repeat("A café word ", 160)) + "**\n\nSee [the guide][guide].\n\n[guide]: https://example.invalid/guide"
	d, err := New(source, "Test")
	if err != nil || len(d.Blocks) < 2 {
		t.Fatal("test must exercise more than one prepared passage")
	}
	before, _ := json.Marshal(d)
	chapters := Chapters(Master(d))
	if len(chapters) != 1 || strings.ContainsAny(chapters[0].Text, "*#") || !strings.Contains(chapters[0].Text, "See the guide.") {
		t.Fatal("Markdown survived passage boundaries")
	}
	preview := string([]rune(chapters[0].Text)[:350])
	if strings.ContainsAny(preview, "*#") {
		t.Fatal("preview truncation reintroduced formatting")
	}
	after, _ := json.Marshal(d)
	if string(after) != string(before) || d.Source != source || !strings.Contains(Master(d), "## **") {
		t.Fatal("speech conversion changed the original, proposal or editable master")
	}
	loaded, err := Load(after)
	if err != nil || loaded.Source != source {
		t.Fatal("Markdown project recovery changed")
	}
}

func TestFormattingOnlyIsNotNarratable(t *testing.T) {
	for _, source := range []string{"---", "***", "#", "```\n```", "[guide]: https://example.invalid"} {
		d, err := New(source, "Test")
		if err != nil {
			t.Fatal(err)
		}
		if Ready(d) == nil {
			t.Fatalf("formatting-only source accepted: %q", source)
		}
	}
}

func TestMasterPreservesLongParagraphSeparators(t *testing.T) {
	source := "# Title\r\n\r\n" + strings.Repeat("A long sentence stays together. ", 100) + "End.\r\n\r\nA new paragraph."
	d, err := New(source, "Test")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := Master(d), strings.ReplaceAll(source, "\r\n", "\n")+"\n"; got != want {
		t.Fatal("working-passage splits changed paragraph boundaries")
	}
}
