import { combineReducers } from 'redux'

export default combineReducers({
  PhotosReducer,
  CollectionsReducer,
  ProfileReducer
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