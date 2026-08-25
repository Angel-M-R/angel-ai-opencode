package assets

import (
	"testing"
	"testing/fstest"
)

func TestDigestIsDeterministicOverPathsAndBytes(t *testing.T) {
	first := New(fstest.MapFS{
		"z/file.txt": &fstest.MapFile{Data: []byte("last\n")},
		"a/file.txt": &fstest.MapFile{Data: []byte("first\n")},
	}, "first")
	second := New(fstest.MapFS{
		"a/file.txt": &fstest.MapFile{Data: []byte("first\n")},
		"z/file.txt": &fstest.MapFile{Data: []byte("last\n")},
	}, "second")

	firstDigest, err := Digest(first)
	if err != nil {
		t.Fatal(err)
	}
	secondDigest, err := Digest(second)
	if err != nil {
		t.Fatal(err)
	}
	if firstDigest != secondDigest {
		t.Fatalf("digest differs by map iteration: %q != %q", firstDigest, secondDigest)
	}

	changed := New(fstest.MapFS{
		"a/file.txt": &fstest.MapFile{Data: []byte("first\n")},
		"z/file.txt": &fstest.MapFile{Data: []byte("changed\n")},
	}, "changed")
	changedDigest, err := Digest(changed)
	if err != nil {
		t.Fatal(err)
	}
	if changedDigest == firstDigest {
		t.Fatal("digest did not change with file content")
	}

	renamed := New(fstest.MapFS{
		"a/file.txt":       &fstest.MapFile{Data: []byte("first\n")},
		"renamed/file.txt": &fstest.MapFile{Data: []byte("last\n")},
	}, "renamed")
	renamedDigest, err := Digest(renamed)
	if err != nil {
		t.Fatal(err)
	}
	if renamedDigest == firstDigest {
		t.Fatal("digest did not change with file path")
	}
}

func TestDigestFramesPathsAndContentUnambiguously(t *testing.T) {
	first := New(fstest.MapFS{
		"a": &fstest.MapFile{Data: []byte("x\x00b")},
		"c": &fstest.MapFile{Data: []byte("y")},
	}, "first")
	second := New(fstest.MapFS{
		"a": &fstest.MapFile{Data: []byte("x")},
		"b": &fstest.MapFile{Data: []byte("c\x00y")},
	}, "second")

	firstDigest, err := Digest(first)
	if err != nil {
		t.Fatal(err)
	}
	secondDigest, err := Digest(second)
	if err != nil {
		t.Fatal(err)
	}
	if firstDigest == secondDigest {
		t.Fatal("different path and content boundaries produced the same digest")
	}
}
