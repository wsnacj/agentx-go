// Package section defines the substrate-neutral section tree exchanged between
// the document pipeline and a host-provided section splitter.
package section

// Node is one matched document section. Rule details intentionally remain in
// the host because the portable pipeline only needs names, pages and topology.
type Node struct {
	Name     string
	Pages    []string
	Matched  []string
	Children []*Node
}

// SectionNode is kept as an experimental source-compatible spelling while the
// Developer Preview document surface is being measured.
type SectionNode = Node
