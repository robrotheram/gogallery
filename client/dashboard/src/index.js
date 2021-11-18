import './index.less';
import React from 'react';
import ReactDOM from 'react-dom';

import * as serviceWorker from './serviceWorker';
import { Provider, useSelector } from 'react-redux';
import { BrowserRouter } from 'react-router-dom';
import store from './store'
import {
  Routes,
  Route,
  Navigate
} from 'react-router-dom'

import Main from './pages/Main';
import Login from './pages/Login';
import Settings from './pages/Settings'

const NoMatch = ({ location }) => (
  <Navigate to="/" />
)

const Protect = (element) => {
  const {auth} = useSelector(state => state.UserReducer)
  if (!auth) {
    return <Navigate to="/login" />
  }
  return element
}

const  AppComponent = () => {
  return(
      <BrowserRouter>
          <Routes>
            <Route path="/" element={Protect(<Main/>)} />
            <Route path="/settings" element={Protect(<Settings/>)} />
            <Route path="/login" element={<Login />} /> 
            <Route element={NoMatch} />
          </Routes>
      </BrowserRouter>
    )
}

ReactDOM.render(
  <Provider store={store()}>
    <AppComponent />
  </Provider>, 
  
document.getElementById('app_root'));

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
