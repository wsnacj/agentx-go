package assetfs

import (
	"io"
	"io/fs"
	"path"
	"sort"
	"time"
)

type snapshotFS struct {
	nodes map[string]snapshotNode
}

type snapshotNode struct {
	name string
	mode fs.FileMode
	data []byte
}

func captureSnapshot(source fs.FS) (*snapshotFS, error) {
	nodes := map[string]snapshotNode{}
	err := fs.WalkDir(source, ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry == nil || !fs.ValidPath(name) {
			return &fs.PathError{Op: "snapshot", Path: name, Err: fs.ErrInvalid}
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		node := snapshotNode{name: entry.Name(), mode: info.Mode()}
		if entry.IsDir() {
			node.mode |= fs.ModeDir
			nodes[name] = node
			return nil
		}
		if !info.Mode().IsRegular() {
			return &fs.PathError{Op: "snapshot", Path: name, Err: fs.ErrInvalid}
		}
		content, err := fs.ReadFile(source, name)
		if err != nil {
			return err
		}
		if info.Size() >= 0 && int64(len(content)) != info.Size() {
			return &fs.PathError{Op: "snapshot", Path: name, Err: fs.ErrInvalid}
		}
		node.data = append([]byte(nil), content...)
		nodes[name] = node
		return nil
	})
	if err != nil {
		return nil, err
	}
	if _, ok := nodes["."]; !ok {
		nodes["."] = snapshotNode{name: ".", mode: fs.ModeDir | 0o555}
	}
	return &snapshotFS{nodes: nodes}, nil
}

func (s *snapshotFS) Open(name string) (fs.File, error) {
	if s == nil || !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	node, ok := s.nodes[name]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	file := &snapshotFile{
		fs:   s,
		path: name,
		node: node,
	}
	if node.mode.IsDir() {
		file.entries = s.childEntries(name)
	}
	return file, nil
}

func (s *snapshotFS) ReadFile(name string) ([]byte, error) {
	if s == nil || !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "read", Path: name, Err: fs.ErrInvalid}
	}
	node, ok := s.nodes[name]
	if !ok {
		return nil, &fs.PathError{Op: "read", Path: name, Err: fs.ErrNotExist}
	}
	if node.mode.IsDir() {
		return nil, &fs.PathError{Op: "read", Path: name, Err: fs.ErrInvalid}
	}
	return append([]byte(nil), node.data...), nil
}

func (s *snapshotFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if s == nil || !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrInvalid}
	}
	node, ok := s.nodes[name]
	if !ok {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
	}
	if !node.mode.IsDir() {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrInvalid}
	}
	return append([]fs.DirEntry(nil), s.childEntries(name)...), nil
}

func (s *snapshotFS) childEntries(dir string) []fs.DirEntry {
	prefix := dir + "/"
	if dir == "." {
		prefix = ""
	}
	entries := make([]fs.DirEntry, 0)
	for name, node := range s.nodes {
		if name == dir || path.Dir(name) != dir {
			continue
		}
		if dir == "." && path.Dir(name) != "." {
			continue
		}
		if dir != "." && len(name) <= len(prefix) {
			continue
		}
		entries = append(entries, snapshotInfo{node: node})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	return entries
}

type snapshotFile struct {
	fs       *snapshotFS
	path     string
	node     snapshotNode
	offset   int
	entries  []fs.DirEntry
	dirIndex int
	closed   bool
}

func (f *snapshotFile) Stat() (fs.FileInfo, error) {
	if f == nil || f.closed {
		return nil, fs.ErrClosed
	}
	return snapshotInfo{node: f.node}, nil
}

func (f *snapshotFile) Read(p []byte) (int, error) {
	if f == nil || f.closed {
		return 0, fs.ErrClosed
	}
	if f.node.mode.IsDir() {
		return 0, &fs.PathError{Op: "read", Path: f.path, Err: fs.ErrInvalid}
	}
	if f.offset >= len(f.node.data) {
		return 0, io.EOF
	}
	n := copy(p, f.node.data[f.offset:])
	f.offset += n
	return n, nil
}

func (f *snapshotFile) ReadDir(n int) ([]fs.DirEntry, error) {
	if f == nil || f.closed {
		return nil, fs.ErrClosed
	}
	if !f.node.mode.IsDir() {
		return nil, &fs.PathError{Op: "readdir", Path: f.path, Err: fs.ErrInvalid}
	}
	if f.dirIndex >= len(f.entries) && n > 0 {
		return nil, io.EOF
	}
	end := len(f.entries)
	if n > 0 && f.dirIndex+n < end {
		end = f.dirIndex + n
	}
	out := append([]fs.DirEntry(nil), f.entries[f.dirIndex:end]...)
	f.dirIndex = end
	return out, nil
}

func (f *snapshotFile) Close() error {
	if f == nil {
		return nil
	}
	f.closed = true
	return nil
}

type snapshotInfo struct {
	node snapshotNode
}

func (i snapshotInfo) Name() string       { return i.node.name }
func (i snapshotInfo) Size() int64        { return int64(len(i.node.data)) }
func (i snapshotInfo) Mode() fs.FileMode  { return i.node.mode }
func (i snapshotInfo) ModTime() time.Time { return time.Time{} }
func (i snapshotInfo) IsDir() bool        { return i.node.mode.IsDir() }
func (i snapshotInfo) Sys() any           { return nil }
func (i snapshotInfo) Type() fs.FileMode  { return i.node.mode.Type() }
func (i snapshotInfo) Info() (fs.FileInfo, error) {
	return i, nil
}
