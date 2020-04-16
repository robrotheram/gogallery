import { createStore, applyMiddleware } from 'redux'
import rootReducer from './reducers'
import thunkMiddleware from 'redux-thunk'
import { createLogger } from 'redux-logger'

import { createBrowserHistory } from 'history';

export const history = createBrowserHistory();

const Constants = {
  prod : {
    baseUrl: "/api",
    imageUrl: "/img/"
   },
   dev : {
    baseUrl: "http://localhost:8800/api",
    imageUrl: "http://localhost:8800/img/"
   }
}

export const config = process.env.NODE_ENV === 'development' ? Constants["dev"] :Constants["prod"];

export function searchTree(element, id){
  const keys = Object.keys(element)
  var result = null
  for (const key of keys) {
    console.log(element[key].id)
    if (element[key].id === id ){
      return element[key]
    } else {
      result = searchTree(element[key].children, id)
      if (result.id !== undefined) {
        return result
      }
    }
  }
  return {}
}


const loggerMiddleware = createLogger()
export default function configureStore(preloadedState) {
    return createStore(
      rootReducer,
      preloadedState,
      applyMiddleware(thunkMiddleware, loggerMiddleware)
    )
  }