// import 'antd/dist/antd.compact.min.css';
// import 'antd/dist/antd.dark.min.css';
// import 'antd/dist/reset.css';
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
import { Button, ConfigProvider, Layout, theme } from 'antd';

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
const { Header, Footer, Sider, Content } = Layout;
root.render(
  <ConfigProvider
    theme={{
      algorithm: [theme.darkAlgorithm],
      components: { 
        Layout:{
          colorBgHeader: "rgb(20, 20, 20)",
          colorBgTrigger: "rgb(20, 20, 20)"
        },
        Menu:{
          colorItemBg: "rgb(20, 20, 20)"
        }
        
      }
    }}
  >
     <Provider store={store()}>
     <AppComponent />
   </Provider>
  </ConfigProvider>
);