package service

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

// PlaygroundDocument is a user-uploaded PDF or text file to attach to a chat turn.
type PlaygroundDocument struct {
	Filename string `json:"filename"`
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`
}

type PlaygroundExtractedDocument struct {
	Filename  string `json:"filename"`
	Text      string `json:"text"`
	Truncated bool   `json:"truncated"`
}

var pdfTextShowOp = regexp.MustCompile(`\((?:\\.|[^\\])*\)\s*Tj|\[(?:[^\[\]]|\((?:\\.|[^\\])*\))*\]\s*TJ`)

func ExtractPlaygroundDocuments(docs []PlaygroundDocument) ([]PlaygroundExtractedDocument, error) {
	if len(docs) == 0 {
		return []PlaygroundExtractedDocument{}, nil
	}
	if len(docs) > PlaygroundMaxDocuments {
		return nil, playgroundError(400, "DOCUMENT_LIMIT", fmt.Sprintf("at most %d documents can be attached", PlaygroundMaxDocuments))
	}
	out := make([]PlaygroundExtractedDocument, 0, len(docs))
	totalRunes := 0
	for i, doc := range docs {
		extracted, err := extractPlaygroundDocument(doc, i)
		if err != nil {
			return nil, err
		}
		runes := utf8.RuneCountInString(extracted.Text)
		if totalRunes+runes > PlaygroundMaxDocumentRunes {
			remain := PlaygroundMaxDocumentRunes - totalRunes
			if remain < 0 {
				remain = 0
			}
			extracted.Text = truncateRunes(extracted.Text, remain)
			extracted.Truncated = true
			if extracted.Text != "" {
				out = append(out, extracted)
			}
			for j := i + 1; j < len(docs); j++ {
				name := strings.TrimSpace(docs[j].Filename)
				if name == "" {
					name = fmt.Sprintf("document-%d", j+1)
				}
				out = append(out, PlaygroundExtractedDocument{Filename: name, Truncated: true})
			}
			return out, nil
		}
		totalRunes += runes
		out = append(out, extracted)
	}
	return out, nil
}

func FormatPlaygroundDocumentPrompt(docs []PlaygroundExtractedDocument) string {
	if len(docs) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("The user attached the following document(s) as context. Use them when answering.\n")
	for _, doc := range docs {
		name := strings.TrimSpace(doc.Filename)
		if name == "" {
			name = "document"
		}
		b.WriteString("\n----- ")
		b.WriteString(name)
		if doc.Truncated {
			b.WriteString(" (truncated)")
		}
		b.WriteString(" -----\n")
		b.WriteString(doc.Text)
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}

func extractPlaygroundDocument(doc PlaygroundDocument, index int) (PlaygroundExtractedDocument, error) {
	name := strings.TrimSpace(doc.Filename)
	if name == "" {
		name = fmt.Sprintf("document-%d", index+1)
	}
	base := filepath.Base(name)
	ext := strings.ToLower(filepath.Ext(base))
	text := strings.TrimSpace(doc.Text)
	var raw []byte
	if strings.TrimSpace(doc.Data) != "" {
		decoded, err := decodePlaygroundDocumentData(doc.Data)
		if err != nil {
			return PlaygroundExtractedDocument{}, playgroundError(400, "DOCUMENT_INVALID", fmt.Sprintf("%s: invalid document data", base))
		}
		raw = decoded
	}
	if len(raw) > PlaygroundMaxDocumentBytes {
		return PlaygroundExtractedDocument{}, playgroundError(400, "DOCUMENT_TOO_LARGE", fmt.Sprintf("%s exceeds the %d MB limit", base, PlaygroundMaxDocumentBytes>>20))
	}
	switch ext {
	case ".txt", ".md", "":
		if text == "" && len(raw) > 0 {
			if !utf8.Valid(raw) {
				return PlaygroundExtractedDocument{}, playgroundError(400, "DOCUMENT_INVALID", fmt.Sprintf("%s is not valid UTF-8 text", base))
			}
			text = string(raw)
		}
		if ext == "" && looksLikePDF(raw) {
			extracted, truncated, err := extractPDFText(raw)
			if err != nil {
				return PlaygroundExtractedDocument{}, playgroundError(400, "DOCUMENT_INVALID", fmt.Sprintf("%s: %s", base, err.Error()))
			}
			return PlaygroundExtractedDocument{Filename: base, Text: strings.TrimSpace(extracted), Truncated: truncated}, nil
		}
		if strings.TrimSpace(text) == "" && len(raw) == 0 {
			return PlaygroundExtractedDocument{}, playgroundError(400, "DOCUMENT_EMPTY", fmt.Sprintf("%s is empty", base))
		}
		return PlaygroundExtractedDocument{Filename: base, Text: strings.TrimSpace(text)}, nil
	case ".pdf":
		if len(raw) == 0 {
			if text != "" {
				return PlaygroundExtractedDocument{Filename: base, Text: text}, nil
			}
			return PlaygroundExtractedDocument{}, playgroundError(400, "DOCUMENT_INVALID", fmt.Sprintf("%s is missing PDF data", base))
		}
		extracted, truncated, err := extractPDFText(raw)
		if err != nil {
			return PlaygroundExtractedDocument{}, playgroundError(400, "DOCUMENT_INVALID", fmt.Sprintf("%s: %s", base, err.Error()))
		}
		return PlaygroundExtractedDocument{Filename: base, Text: strings.TrimSpace(extracted), Truncated: truncated}, nil
	default:
		return PlaygroundExtractedDocument{}, playgroundError(400, "DOCUMENT_TYPE", fmt.Sprintf("%s is not a supported document type (PDF, TXT, or Markdown)", base))
	}
}

func decodePlaygroundDocumentData(data string) ([]byte, error) {
	data = strings.TrimSpace(data)
	if i := strings.Index(data, "base64,"); i >= 0 {
		data = data[i+len("base64,"):]
	}
	data = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == ' ' {
			return -1
		}
		return r
	}, data)
	return base64.StdEncoding.DecodeString(data)
}

func looksLikePDF(raw []byte) bool {
	return bytes.HasPrefix(bytes.TrimSpace(raw), []byte("%PDF"))
}

func extractPDFText(raw []byte) (string, bool, error) {
	if !looksLikePDF(raw) {
		return "", false, fmt.Errorf("file is not a PDF")
	}
	var texts []string
	streams := findPDFStreams(raw)
	for _, stream := range streams {
		content := stream
		if decoded, err := zlib.NewReader(bytes.NewReader(stream)); err == nil {
			if body, readErr := io.ReadAll(io.LimitReader(decoded, PlaygroundMaxDocumentBytes)); readErr == nil {
				_ = decoded.Close()
				content = body
			} else {
				_ = decoded.Close()
			}
		}
		texts = append(texts, extractPDFShowText(content)...)
	}
	if len(texts) == 0 {
		texts = extractPDFShowText(raw)
	}
	joined := strings.TrimSpace(strings.Join(texts, " "))
	if joined == "" {
		return "", false, fmt.Errorf("could not extract text from this PDF")
	}
	if utf8.RuneCountInString(joined) > PlaygroundMaxDocumentRunes {
		return truncateRunes(joined, PlaygroundMaxDocumentRunes), true, nil
	}
	return joined, false, nil
}

func findPDFStreams(raw []byte) [][]byte {
	var out [][]byte
	data := raw
	for {
		start := bytes.Index(data, []byte("stream"))
		if start < 0 {
			return out
		}
		payload := data[start+len("stream"):]
		if len(payload) >= 2 && payload[0] == '\r' && payload[1] == '\n' {
			payload = payload[2:]
		} else if len(payload) >= 1 && (payload[0] == '\n' || payload[0] == '\r') {
			payload = payload[1:]
		}
		end := bytes.Index(payload, []byte("endstream"))
		if end < 0 {
			return out
		}
		out = append(out, payload[:end])
		data = payload[end+len("endstream"):]
	}
}

func extractPDFShowText(content []byte) []string {
	matches := pdfTextShowOp.FindAll(content, -1)
	var parts []string
	for _, match := range matches {
		parts = append(parts, parsePDFTextOperators(match)...)
	}
	return parts
}

func parsePDFTextOperators(op []byte) []string {
	var parts []string
	data := op
	for {
		start := bytes.IndexByte(data, '(')
		if start < 0 {
			return parts
		}
		text, rest, ok := readPDFLiteral(data[start+1:])
		if !ok {
			return parts
		}
		if strings.TrimSpace(text) != "" {
			parts = append(parts, text)
		}
		data = rest
	}
}

func readPDFLiteral(data []byte) (string, []byte, bool) {
	var b strings.Builder
	for i := 0; i < len(data); i++ {
		ch := data[i]
		if ch == ')' {
			return b.String(), data[i+1:], true
		}
		if ch != '\\' {
			b.WriteByte(ch)
			continue
		}
		if i+1 >= len(data) {
			break
		}
		i++
		switch data[i] {
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case 't':
			b.WriteByte('\t')
		case 'b':
			b.WriteByte('\b')
		case 'f':
			b.WriteByte('\f')
		case '(', ')', '\\':
			b.WriteByte(data[i])
		case '\n', '\r':
			// line continuation
		default:
			if data[i] >= '0' && data[i] <= '7' {
				oct := data[i:min(i+3, len(data))]
				n := 0
				used := 0
				for _, d := range oct {
					if d < '0' || d > '7' {
						break
					}
					n = n*8 + int(d-'0')
					used++
				}
				if used > 0 {
					b.WriteByte(byte(n))
					i += used - 1
				}
			}
		}
	}
	return "", data, false
}

func truncateRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	var b strings.Builder
	count := 0
	for _, r := range s {
		if count >= n {
			break
		}
		b.WriteRune(r)
		count++
	}
	return b.String()
}
