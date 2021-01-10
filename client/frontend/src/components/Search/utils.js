import Fuse from 'fuse.js';

export const fuzzySearch = (keys, data, searchTerm) => {

    console.log("SEARCH", searchTerm)
    const fuse = new Fuse(data, {
        isCaseSensitive: true,
        keys: keys
      })
    let results = fuse.search(searchTerm)
    return results.map(i => i.item)
}