import { createStore, applyMiddleware } from 'redux'
import rootReducer from './reducers'

import thunkMiddleware from 'redux-thunk'
import { createLogger } from 'redux-logger'

import { createBrowserHistory } from 'history';

export const history = createBrowserHistory();

export const config = {
    baseUrl: "http://localhost:8800"
}


const loggerMiddleware = createLogger()



export default function configureStore(preloadedState) {
    return createStore(
      rootReducer,
      preloadedState,
      applyMiddleware(thunkMiddleware, loggerMiddleware)
    )
  }