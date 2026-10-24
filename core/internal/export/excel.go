package export

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"

	"github.com/negitanu/etherscribe/internal/model"
)

// RenderExcel builds a minimal OOXML .xlsx without third-party libs.
func RenderExcel(cfg *model.Config, sections []string) ([]byte, error) {
	sheets, err := reportTables(cfg, sections)
	if err != nil {
		return nil, err
	}
	return writeXLSX(sheets)
}

type sheetData struct {
	Name string
	Rows [][]string
}

func writeXLSX(sheets []sheetData) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	must := func(name, body string) error {
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(body))
		return err
	}
	if err := must("[Content_Types].xml", contentTypesXML(len(sheets))); err != nil {
		return nil, err
	}
	if err := must("_rels/.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`); err != nil {
		return nil, err
	}
	if err := must("xl/workbook.xml", workbookXML(sheets)); err != nil {
		return nil, err
	}
	if err := must("xl/_rels/workbook.xml.rels", workbookRelsXML(len(sheets))); err != nil {
		return nil, err
	}
	if err := must("xl/styles.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <fonts count="2"><font><sz val="11"/><name val="Yu Gothic"/><color rgb="FF243650"/></font><font><b/><sz val="11"/><name val="Yu Gothic"/><color rgb="FFFFFFFF"/></font></fonts>
  <fills count="3"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill><fill><patternFill patternType="solid"><fgColor rgb="FF244B7A"/><bgColor indexed="64"/></patternFill></fill></fills>
  <borders count="1"><border/></borders>
  <cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>
  <cellXfs count="2"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0" applyAlignment="1"><alignment vertical="top" wrapText="1"/></xf><xf numFmtId="0" fontId="1" fillId="2" borderId="0" xfId="0" applyAlignment="1" applyFont="1" applyFill="1"><alignment vertical="center" wrapText="1"/></xf></cellXfs>
  <cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>
</styleSheet>`); err != nil {
		return nil, err
	}
	for i, sh := range sheets {
		if err := must(fmt.Sprintf("xl/worksheets/sheet%d.xml", i+1), sheetXML(sh.Rows)); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func contentTypesXML(n int) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`)
	b.WriteString(`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>`)
	b.WriteString(`<Default Extension="xml" ContentType="application/xml"/>`)
	b.WriteString(`<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>`)
	b.WriteString(`<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>`)
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&b, `<Override PartName="/xl/worksheets/sheet%d.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`, i)
	}
	b.WriteString(`</Types>`)
	return b.String()
}

func workbookXML(sheets []sheetData) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">`)
	b.WriteString(`<sheets>`)
	for i, sh := range sheets {
		fmt.Fprintf(&b, `<sheet name="%s" sheetId="%d" r:id="rId%d"/>`, xmlEscape(sh.Name), i+1, i+1)
	}
	b.WriteString(`</sheets></workbook>`)
	return b.String()
}

func workbookRelsXML(n int) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&b, `<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet%d.xml"/>`, i, i)
	}
	b.WriteString(`<Relationship Id="rIdStyles" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>`)
	b.WriteString(`</Relationships>`)
	return b.String()
}

func sheetXML(rows [][]string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetViews><sheetView workbookViewId="0"><pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/></sheetView></sheetViews><sheetFormatPr defaultRowHeight="30"/><cols>`)
	maxCols := 1
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}
	for c := 0; c < maxCols; c++ {
		width := 24
		for _, row := range rows {
			if c < len(row) && len([]rune(row[c])) > 32 {
				width = 48
			}
		}
		fmt.Fprintf(&b, `<col min="%d" max="%d" width="%d" customWidth="1"/>`, c+1, c+1, width)
	}
	b.WriteString(`</cols><sheetData>`)
	for r, row := range rows {
		fmt.Fprintf(&b, `<row r="%d">`, r+1)
		for c, val := range row {
			ref := cellRef(c, r)
			style := 0
			if r == 0 {
				style = 1
			}
			fmt.Fprintf(&b, `<c r="%s" s="%d" t="inlineStr"><is><t xml:space="preserve">%s</t></is></c>`, ref, style, xmlEscape(val))
		}
		b.WriteString(`</row>`)
	}
	b.WriteString(`</sheetData>`)
	if len(rows) > 0 {
		fmt.Fprintf(&b, `<autoFilter ref="A1:%s"/>`, cellRef(maxCols-1, len(rows)-1))
	}
	b.WriteString(`<pageSetup orientation="landscape" paperSize="9"/></worksheet>`)
	return b.String()
}

func cellRef(col, row int) string {
	// col 0 -> A
	name := ""
	c := col
	for {
		name = string(rune('A'+c%26)) + name
		c = c/26 - 1
		if c < 0 {
			break
		}
	}
	return fmt.Sprintf("%s%d", name, row+1)
}

func xmlEscape(s string) string {
	r := strings.NewReplacer(`&`, "&amp;", `<`, "&lt;", `>`, "&gt;", `"`, "&quot;")
	return r.Replace(s)
}
