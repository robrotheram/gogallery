

const initial_state = {
    photos: [], 
    isUpdating: false
}


export function PhotosReducer(state = initial_state, action) {
    switch (action.type) {
        case 'PHOTOS_FETCHING':
            return {
            ...state,
            isUpdating: true
            };
        case 'PHOTOS_RECEIVED':
            return {
                ...state,
                isUpdating: false,
                photos: action.photos
            };
        default:
            return state
    }
  }