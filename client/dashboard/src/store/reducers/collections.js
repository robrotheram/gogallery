const initialState = {
    collections: [],
    dates: [],
    uploadDates: [],
    isUpdating: false
};

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
                collections: action.collections,
                dates: action.dates,
                uploadDates: action.uploadDates
            };
        default:
            return state
    }
  }