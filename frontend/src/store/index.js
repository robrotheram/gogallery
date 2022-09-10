import { createStore, applyMiddleware } from 'redux'
import rootReducer from './reducers'
import thunkMiddleware from 'redux-thunk'
import { createLogger } from 'redux-logger'

import { createBrowserHistory } from 'history';

export const history = createBrowserHistory();

const Constants = {
  prod : {
    baseUrl: "http://localhost:8800/api/admin",
    imageUrl: "http://localhost:8800/img/"
   },
   dev : {
    baseUrl: "http://localhost:8800/api/admin",
    imageUrl: "http://localhost:8800/img/"
   }
}
export const config = {
  baseUrl: "http://localhost:8800/api/admin",
  imageUrl: "http://localhost:8800/img/"
 }

const loggerMiddleware = createLogger()
export default function configureStore(preloadedState) {
    return createStore(
      rootReducer,
      preloadedState,
      applyMiddleware(thunkMiddleware)
    )
  }


export function IDFromTree(collections, key){
  key = key.split("-")
  key.shift()
  let el = {children:collections}
  key.forEach(k => {
    el = el.children[parseInt(k)]
  })
  return el
}