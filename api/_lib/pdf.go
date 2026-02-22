package lib

import (
	"bytes"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
)

// ReadPDF parses text out of a PDF byte slice.
func ReadPDF(fileData []byte) (string, error) {
	reader, err := pdf.NewReader(bytes.NewReader(fileData), int64(len(fileData)))
	if err != nil {
		return "", err
	}

	b, err := reader.GetPlainText()
	if err != nil && err != io.EOF {
		return "", err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(b)

	cleaned := strings.ReplaceAll(buf.String(), "\n", " ")
	cleaned = strings.Join(strings.Fields(cleaned), " ")

	return cleaned, nil
}

// ChunkText splits a string into overlapping chunks by words.
func ChunkText(text string, chunkSize int, overlap int) []string {
	var chunks []string
	words := strings.Fields(text)

	for i := 0; i < len(words); {
		end := i + chunkSize
		if end > len(words) {
			end = len(words)
		}

		chunk := strings.Join(words[i:end], " ")
		chunks = append(chunks, chunk)

		if end == len(words) {
			break
		}

		i += (chunkSize - overlap)
		if i >= len(words) {
			break
		}
	}

	return chunks
}
