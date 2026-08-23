package assets

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"io/fs"
	"sort"
)

// Digest identifies every regular file in a source by its slash-separated
// path and bytes. File modes and directory entries do not affect the result.
func Digest(source Source) (string, error) {
	var paths []string
	if err := source.WalkDir(".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type().IsRegular() {
			paths = append(paths, name)
		}
		return nil
	}); err != nil {
		return "", err
	}
	sort.Strings(paths)

	digest := sha256.New()
	for _, name := range paths {
		content, err := source.ReadFile(name)
		if err != nil {
			return "", err
		}
		writeDigestField(digest, []byte(name))
		writeDigestField(digest, content)
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func writeDigestField(digest interface{ Write([]byte) (int, error) }, value []byte) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = digest.Write(length[:])
	_, _ = digest.Write(value)
}
