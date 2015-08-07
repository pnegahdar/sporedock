const initialState = {
  apps: []
}

export function webapp(state = initialState, action) {
  console.log('webapp', state, action)
  switch (action.type) {
  case 'APP_LIST':
    return {
      ...state,
      apps: action.params
    }
  }
  return state
}
