package lsp

import (
	"strings"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

// document represents a document that is open in the editor.
// The Content and lines are parsed when the document is opened or changed to avoid having to perform
// the same operation multiple times and to keep the interaction snappy for hover operations.
type document struct {
	uri     protocol.DocumentUri
	Content string
	lines   []string
}

// documentStore is a simple in-memory store for documents that are open in the editor.
// Its primary purpose it to provide quick access to open files for operations like hover.
// Rather than reading the file from disk, we can get the document from the store.
type documentStore struct {
	documents map[protocol.DocumentUri]*document
}

func newDocumentStore() *documentStore {
	return &documentStore{
		documents: make(map[protocol.DocumentUri]*document),
	}
}

func (ds *documentStore) get(uri protocol.DocumentUri) (*document, bool) {
	doc, ok := ds.documents[uri]
	return doc, ok
}

func (ds *documentStore) set(uri protocol.DocumentUri, content string) {
	ds.documents[uri] = &document{
		uri:     uri,
		Content: content,
		lines:   strings.Split(content, "\n"),
	}
}

func (ds *documentStore) delete(uri protocol.DocumentUri) {
	delete(ds.documents, uri)
}

func (d *document) update(changes []interface{}) {
	for _, change := range changes {
		switch c := change.(type) {
		case protocol.TextDocumentContentChangeEvent:
			startIndex, endIndex := c.Range.IndexesIn(d.Content)
			d.Content = d.Content[:startIndex] + c.Text + d.Content[endIndex:]
		case protocol.TextDocumentContentChangeEventWhole:
			d.Content = c.Text
		}
	}

	d.lines = strings.Split(d.Content, "\n")
}
