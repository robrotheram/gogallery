

const initial_state = {
    search: {}
}


export function SearchReducer(state = initial_state, action) {
    switch (action.type) {
        case 'UPDATE_SEARCH':
            return {
            ...state,
            search: {...state.search, ...action.search}
            };
        default:
            return state
    }
  }