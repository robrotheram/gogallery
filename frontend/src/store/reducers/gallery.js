
const initialState = {
    addCollectionModalVisable: false,
    uploadModalVisable: false,
    imageSize: "small"
};

export function GalleryReducer(state = initialState, action) {
    switch (action.type) {
        case 'SHOW_ADD_MODAL':
            return {...state, addCollectionModalVisable: true };
        case 'HIDE_ADD_MODAL':
                return {...state, addCollectionModalVisable: false };
        case 'SHOW_UPLOAD_MODAL':
            return {...state, uploadModalVisable: true };
        case 'HIDE_UPLOAD_MODAL':
                return {...state, uploadModalVisable: false };
        case 'CHANGE_IMAGE_SIZE':
            return {...state, imageSize: action.size};
        default:
            return state
    }
  }