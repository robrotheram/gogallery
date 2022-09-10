
const initialState = {
    photos: [],
    isUpdating: false
};

export function PhotoReducer(state = initialState, action) {
    switch (action.type) {
        case 'PHOTO_FETCHING':
            return {
            ...state,
            isUpdating: true
            };
        case 'PHOTO_RECEIVED':
            return {
                ...state,
                isUpdating: false,
                photos: action.photos
            };
        default:
            return state
    }
  }