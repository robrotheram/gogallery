import { combineReducers } from 'redux'

import {UserReducer} from './user'
import {PhotoReducer} from './photos'
import {CollectionsReducer} from './collections'
import {GalleryReducer} from './gallery'
import {SettingsReducer} from './settings'

export default combineReducers({
    UserReducer,
    CollectionsReducer,
    PhotoReducer,
    GalleryReducer,
    SettingsReducer
  })