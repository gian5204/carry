package filetransfer

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gian5204/carry/internal/managedpath"
)

var (
	ErrDestinationExists = errors.New("destination file already exists")
	ErrUnsafePath        = errors.New("unsafe file path")
)

type Destination struct {
	file      *os.File
	temporary string
	final     string
	committed bool
}

func OpenSource(repositoryRoot, relativePath string) (*os.File, int64, error) {
	_, target, err := resolve(repositoryRoot, relativePath)
	if err != nil {
		return nil, 0, err
	}
	if err := ensureResolvedInside(repositoryRoot, target); err != nil {
		return nil, 0, err
	}

	file, err := os.Open(target)
	if err != nil {
		return nil, 0, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, 0, err
	}
	if !info.Mode().IsRegular() {
		file.Close()
		return nil, 0, fmt.Errorf("managed path is not a regular file")
	}

	return file, info.Size(), nil
}

func PrepareDestination(repositoryRoot, relativePath string) (*Destination, error) {
	root, target, err := resolve(repositoryRoot, relativePath)
	if err != nil {
		return nil, err
	}

	parent, err := createSafeParents(root, filepath.Dir(target))
	if err != nil {
		return nil, err
	}
	if _, err := os.Lstat(target); err == nil {
		return nil, ErrDestinationExists
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	temporary, err := os.CreateTemp(parent, ".carry-transfer-*")
	if err != nil {
		return nil, err
	}

	return &Destination{
		file:      temporary,
		temporary: temporary.Name(),
		final:     target,
	}, nil
}

func (d *Destination) Write(data []byte) (int, error) {
	return d.file.Write(data)
}

func (d *Destination) Commit() error {
	if err := d.file.Sync(); err != nil {
		return err
	}
	if err := d.file.Close(); err != nil {
		return err
	}
	d.file = nil

	if _, err := os.Lstat(d.final); err == nil {
		return ErrDestinationExists
	} else if !os.IsNotExist(err) {
		return err
	}

	// A same-filesystem hard link publishes the completed temporary file
	// atomically without the overwrite behavior of os.Rename.
	if err := os.Link(d.temporary, d.final); err != nil {
		if _, statErr := os.Lstat(d.final); statErr == nil {
			return ErrDestinationExists
		}
		return err
	}
	if err := os.Remove(d.temporary); err != nil {
		return errors.Join(err, os.Remove(d.final))
	}

	d.committed = true
	return nil
}

func (d *Destination) Abort() error {
	if d.committed {
		return nil
	}
	var closeErr error
	if d.file != nil {
		closeErr = d.file.Close()
		d.file = nil
	}
	removeErr := os.Remove(d.temporary)
	if os.IsNotExist(removeErr) {
		removeErr = nil
	}
	return errors.Join(closeErr, removeErr)
}

func resolve(repositoryRoot, relativePath string) (string, string, error) {
	normalized, err := managedpath.Normalize(relativePath)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", ErrUnsafePath, err)
	}

	root, err := filepath.Abs(repositoryRoot)
	if err != nil {
		return "", "", err
	}
	target := filepath.Join(root, filepath.FromSlash(normalized))
	relative, err := filepath.Rel(root, target)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", ErrUnsafePath, err)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", "", ErrUnsafePath
	}

	return root, target, nil
}

func ensureResolvedInside(repositoryRoot, target string) error {
	root, err := filepath.EvalSymlinks(repositoryRoot)
	if err != nil {
		return err
	}
	target, err = filepath.EvalSymlinks(target)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(root, target)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnsafePath, err)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return ErrUnsafePath
	}
	return nil
}

func createSafeParents(root, parent string) (string, error) {
	relative, err := filepath.Rel(root, parent)
	if err != nil {
		return "", err
	}

	current := root
	if relative == "." {
		return current, nil
	}
	for _, segment := range strings.Split(relative, string(filepath.Separator)) {
		current = filepath.Join(current, segment)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			if err := os.Mkdir(current, 0755); err != nil && !os.IsExist(err) {
				return "", err
			}
			info, err = os.Lstat(current)
		}
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return "", ErrUnsafePath
		}
	}

	return current, nil
}

var _ io.Writer = (*Destination)(nil)
