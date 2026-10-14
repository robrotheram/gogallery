package pages

import (
	"gogallery/pkg/datastore"
	"testing"
	"time"
)

func TestAlbumHierarchyUsesPathsAndIncludesDescendantPhotos(t *testing.T) {
	albums := []datastore.Album{
		{Id: "root", Name: "Renamed holiday", Path: "/photos/2026"},
		{Id: "child", Name: "Garden", Path: "/photos/2026/Garden", Parent: "2026"},
		{Id: "leaf", Name: "Close-ups", Path: "/photos/2026/Garden/Macro"},
		{Id: "other", Name: "Garden", Path: "/photos/2025/Garden"},
		{Id: "empty", Name: "Empty", Path: "/photos/2026/Empty"},
		{Id: "prefix", Name: "Unrelated", Path: "/photos/20260/Other"},
	}
	pictures := []datastore.Picture{
		{Id: "direct", Album: "child", DateTaken: time.Unix(1, 0)},
		{Id: "nested", Album: "leaf", DateTaken: time.Unix(2, 0)},
		{Id: "unrelated", Album: "other", DateTaken: time.Unix(3, 0)},
	}
	cards := map[string]albumCardData{}
	for _, card := range albumCards(albums, pictures) {
		cards[card.album.Id] = card
	}
	if cards["child"].parentID != "root" || cards["leaf"].parentID != "child" {
		t.Fatal("renamed parent broke nesting")
	}
	if cards["other"].parentID != "" || cards["prefix"].parentID != "" {
		t.Fatal("names or path prefixes merged unrelated albums")
	}
	if root := cards["root"]; root.childCount != 2 || root.count != 2 || root.picture.Id != "direct" {
		t.Fatalf("parent summary/cover: %#v", root)
	}
	if child := cards["child"]; child.childCount != 1 || child.count != 2 || child.picture.Id != "direct" {
		t.Fatalf("mixed album summary: %#v", child)
	}
	if cards["empty"].count != 0 || cards["empty"].picture.Id != "" {
		t.Fatal("empty child must remain empty")
	}
	// A saved cover from a descendant is valid, but a sibling's photo is not.
	albums[0].ProfileId = "nested"
	for _, card := range albumCards(albums, pictures) {
		if card.album.Id == "root" && card.picture.Id != "nested" {
			t.Fatal("valid descendant cover was discarded")
		}
	}
	albums[0].ProfileId = "unrelated"
	for _, card := range albumCards(albums, pictures) {
		if card.album.Id == "root" && card.picture.Id == "unrelated" {
			t.Fatal("unrelated cover was accepted")
		}
	}
}
func TestAlbumHierarchyKeepsOrphansReachable(t *testing.T) {
	parents := albumParents([]datastore.Album{
		{Id: "root", Name: "Travel", Path: "/photos/Travel"},
		{Id: "leaf", Name: "Leaf", Path: "/photos/Travel/Missing/Leaf"},
		{Id: "orphan", Name: "Orphan", Path: "/photos/Unknown/Orphan"},
	})
	if parents["leaf"] != "root" {
		t.Fatal("missing intermediate folder should use nearest known ancestor")
	}
	if parents["orphan"] != "" {
		t.Fatal("orphan should remain accessible at the top level")
	}
}
