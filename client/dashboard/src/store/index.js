import { createStore, applyMiddleware } from 'redux'
import rootReducer from './reducers'
import {composeWithDevTools} from 'redux-devtools-extension'
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
      composeWithDevTools(applyMiddleware(thunkMiddleware, loggerMiddleware))
    )
  }

// export function formatTree(element){
//     const keys = Object.keys(element)
//     for (const key of keys) {
//       //console.log(element[key].id)
//       element[key].title = element[key].name
//       element[key].value = element[key].id
      
//       formatTree(element[key].children)
//       element[key].children = Object.values(element[key].children)
//     }
//   }

export function IDFromTree(collections, key){
  key = key.split("-")
  key.shift()
  let el = {children:collections}
  key.forEach(k => {
    el = el.children[parseInt(k)]
  })
  return el
}