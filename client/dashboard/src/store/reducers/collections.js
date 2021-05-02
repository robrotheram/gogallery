const initialState = {
    collections: [],
    dates: [],
    uploadDates: [],
    isUpdating: false
};

const convertToTree = (tree) => {
    const proceesNode = (node) => {
        node.key = node.id
        node.value = node.id
        node.title = node.name
        node.children = Object.values(node.children)
        console.log("covertToTree start", node.children)
        if (node.children.length === 0 || node.children === undefined ) {
            node["key"] = node.id
            console.log("covertToTree", node.children)
            return node
        }
        console.log("covertToTree search", node)
        node.children = node.children.map(n => proceesNode(n))
        return node
    }
    tree = Object.values(tree)
    console.log("test", tree)
    return tree.map(node => proceesNode(node))
}

export function CollectionsReducer(state = initialState, action) {
    switch (action.type) {
        case 'COLLECTIONS_FETCHING':
            return {
            ...state,
            isUpdating: true
            };
        case 'COLLECTIONS_RECEIVED':
            return {
                ...state,
                isUpdating: false,
                collections: convertToTree(action.collections),
                dates: action.dates,
                uploadDates: action.uploadDates
            };
        default:
            return state
    }
  }