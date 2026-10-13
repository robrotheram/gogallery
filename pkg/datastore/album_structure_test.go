package datastore

import "testing"

func TestGetParentAlbumFromStructure(t *testing.T) {
	root := Album{Id: "root", Name: "Travel"}.ToAlbumNode()
	child := Album{Id: "child", Name: "Italy"}.ToAlbumNode()
	child.Children["Rome"] = Album{Id: "grandchild", Name: "Rome"}.ToAlbumNode()
	root.Children["Italy"] = child
	tree := AlbumStructure{"Travel": root}
	for _, tc := range []struct{ id, parent string }{
		{"root", ""}, {"child", "root"}, {"grandchild", "child"}, {"missing", ""},
	} {
		t.Run(tc.id, func(t *testing.T) {
			if got := GetParentAlbumFromStructure(tree, tc.id); got.Id != tc.parent {
				t.Fatalf("parent = %q, want %q", got.Id, tc.parent)
			}
		})
	}
}
