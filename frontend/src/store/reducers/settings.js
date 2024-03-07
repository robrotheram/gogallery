
const initialState = {
  stats: {},
  profile:{},
  gallery:{},
  deploy:{},
  isUpdating: false
};

export function SettingsReducer(state = initialState, action) {
  switch (action.type) {
    case 'STATS_FETCHING':
      return {
        ...state,
        isUpdating: true
      };
    case 'STATS_UPDATED':
      return {
        ...state,
        isUpdating: false,
        stats: action.stats
      };
    case 'PROFILE_UPDATED':
      return {
        ...state,
        isUpdating: false,
        profile: action.profile
      }
    case 'GALLERY_UPDATED':
      return {
        ...state,
        isUpdating: false,
        gallery: action.gallery
      };
      case 'DEPLOY_UPDATED':
        return {
          ...state,
          isUpdating: false,
          deploy: action.deploy
        };
    default:
      return state
  }
}