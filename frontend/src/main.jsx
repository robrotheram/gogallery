import 'antd/dist/antd.compact.min.css';
import 'antd/dist/antd.dark.min.css';

import React from 'react';
import {createRoot} from 'react-dom/client';

import { Provider } from 'react-redux';
import { BrowserRouter } from 'react-router-dom';
import store from './store'
import {
  Routes,
  Route,
  Navigate
} from 'react-router-dom'

import Main from './pages/Main';
import Settings from './pages/Settings'

const NoMatch = ({ location }) => (
  <Navigate to="/" />
)

const  AppComponent = () => {
  return(
      <BrowserRouter>
          <Routes>
            <Route path="/" element={(<Main/>)} />
            <Route path="/settings" element={(<Settings/>)} />
            <Route element={NoMatch} />
          </Routes>
      </BrowserRouter>
    )
}


const container = document.getElementById('app_root');
const root = createRoot(container);
root.render(
    <Provider store={store()}>
    <AppComponent />
  </Provider>
);