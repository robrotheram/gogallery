package pages

import (
	"path/filepath"
	"sort"
	"strings"

	"gogallery/pkg/datastore"
)

type albumCardData struct {
	album             datastore.Album
	picture           datastore.Picture
	parentID          string
	count, childCount int
}

// Match actual folder paths, never display titles or the Parent basename:
// albums may be renamed, and different branches can contain identical names.
// Missing intermediate folders promote their children to the nearest known
// ancestor, keeping older/incomplete libraries navigable.
func albumParents(albums []datastore.Album) map[string]string {
	paths := make(map[string]string, len(albums))
	for _, album := range albums {
		if album.Path != "" {
			paths[filepath.Clean(album.Path)] = album.Id
		}
	}
	parents := make(map[string]string, len(albums))
	for _, album := range albums {
		parent := album.ParentPath
		if album.Path != "" {
			parent = filepath.Dir(filepath.Clean(album.Path))
		}
		if parent == "" {
			continue
		}
		for {
			if id, ok := paths[parent]; ok && id != album.Id {
				parents[album.Id] = id
				break
			}
			next := filepath.Dir(parent)
			if next == parent {
				break
			}
			parent = next
		}
	}
	return parents
}

// Counts include descendants. A parent without its own photos inherits a
// cover from its children instead of looking like an empty album.
func albumCards(albums []datastore.Album, pictures []datastore.Picture) []albumCardData {
	parents := albumParents(albums)
	children := make(map[string][]string)
	byID := make(map[string]datastore.Picture, len(pictures))
	newest := make(map[string]datastore.Picture)
	counts := make(map[string]int)
	for _, pic := range pictures {
		byID[pic.Id] = pic
		counts[pic.Album]++
		old, exists := newest[pic.Album]
		if !exists || pic.DateTaken.After(old.DateTaken) || (pic.DateTaken.Equal(old.DateTaken) && pic.Id < old.Id) {
			newest[pic.Album] = pic
		}
	}
	cards := make(map[string]albumCardData, len(albums))
	for _, album := range albums {
		parentID := parents[album.Id]
		if parentID != "" {
			children[parentID] = append(children[parentID], album.Id)
		}
		cards[album.Id] = albumCardData{album: album, parentID: parentID, count: counts[album.Id]}
	}
	belongsTo := func(albumID, parentID string) bool {
		for albumID != "" {
			if albumID == parentID {
				return true
			}
			albumID = parents[albumID]
		}
		return false
	}
	visited := make(map[string]bool)
	var collect func(string) albumCardData
	collect = func(id string) albumCardData {
		card := cards[id]
		if visited[id] {
			return card
		}
		visited[id] = true
		card.childCount = len(children[id])
		childCover := datastore.Picture{}
		for _, childID := range children[id] {
			child := collect(childID)
			card.count += child.count
			pic := child.picture
			if pic.Id != "" && (childCover.Id == "" || pic.DateTaken.After(childCover.DateTaken) || (pic.DateTaken.Equal(childCover.DateTaken) && pic.Id < childCover.Id)) {
				childCover = pic
			}
		}
		cover, ok := byID[card.album.ProfileId]
		if !ok || !belongsTo(cover.Album, id) {
			cover = newest[id]
		}
		if cover.Id == "" {
			cover = childCover
		}
		card.picture = cover
		cards[id] = card
		return card
	}
	result := make([]albumCardData, 0, len(albums))
	for _, album := range albums {
		result = append(result, collect(album.Id))
	}
	sort.SliceStable(result, func(i, j int) bool {
		a, b := strings.ToLower(result[i].album.Name), strings.ToLower(result[j].album.Name)
		if a == b {
			return result[i].album.Path < result[j].album.Path
		}
		return a < b
	})
	return result
}
