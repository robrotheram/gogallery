import { combineReducers } from 'redux'

export default combineReducers({
  PhotosReducer,
  CollectionsReducer,
  ProfileReducer,
  ConfigReducer
  })

export function PhotosReducer(state =  { photos: [], isUpdating: false }, action) {
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
export function ProfileReducer(state =  { profile: {}, isUpdating: false }, action) {
  switch (action.type) {
      case 'PROFILE_FETCHING':
          return {
          ...state,
          isUpdating: true
          };
      case 'PROFILE_RECEIVED':
          return {
              ...state,
              isUpdating: false,
              profile: action.profile
          };
      default:
          return state
  }
}
export function CollectionsReducer(state =  { collections: {}, isUpdating: false }, action) {
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
              collections: action.collections
          };
      default:
          return state
  }
}
export function ConfigReducer(state =  { config: {}, isUpdating: false }, action) {
    switch (action.type) {
        case 'CONFIG_FETCHING':
            return {
            ...state,
            isUpdating: true
            };
        case 'CONFIG_RECEIVED':
            return {
                ...state,
                isUpdating: false,
                config: action.config
            };
        default:
            return state
    }
  }