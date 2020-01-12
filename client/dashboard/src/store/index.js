import { createStore, applyMiddleware } from 'redux'
import rootReducer from './reducers'

import thunkMiddleware from 'redux-thunk'
import { createLogger } from 'redux-logger'

import { createBrowserHistory } from 'history';

export const history = createBrowserHistory({
  basename: process.env.PUBLIC_URL
});

const Constants = {
  prod : {
    baseUrl: "/api/admin",
    imageUrl: "/img/"
   },
   dev : {
    baseUrl: "http://localhost:8800/api/admin",
    imageUrl: "http://localhost:8800/img/"
   }
}
export const config = process.env.NODE_ENV === 'development' ? Constants["dev"] :Constants["prod"];

const loggerMiddleware = createLogger()
export default function configureStore(preloadedState) {
    return createStore(
      rootReducer,
      preloadedState,
      applyMiddleware(thunkMiddleware, loggerMiddleware)
    )
  }