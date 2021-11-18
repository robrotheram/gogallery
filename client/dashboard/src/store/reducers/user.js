
const initialState = { 
  loggedIn: true, 
  loginFailed: false,
  token: localStorage.getItem('token') || "", 
  email: localStorage.getItem('email') || "", 
  username: localStorage.getItem('username') || "" 
};

export function UserReducer(state = initialState, action) {
  switch (action.type) {
    case 'LOGIN_SUCCESS':
      console.log("UDPATING USER SETTINGS")
      return {
        loggingIn: true,
        token: action.token,
        username:  action.username,
        email:  action.email,
        loginFailed: false,
        auth: true
      };
    case 'LOGOUT_SUCCESS':
      return {
        auth: false,
        loginFailed: false
      };
    case 'LOGOUT_FAILED':
    return {
      auth: false,
      loginFailed: true
    };
    default:
      return state
  }
}