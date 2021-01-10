import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import * as serviceWorkerRegistration from './serviceWorkerRegistration';
import reportWebVitals from './reportWebVitals';


import {Route} from 'react-router-dom'
import { Provider } from 'react-redux';
import { connect } from 'react-redux';
import { ConnectedRouter } from 'connected-react-router'

import store, {history} from './store'
import { galleryActions } from './store/actions';
import Header from './components/header'
import IndexPage from './pages/Index';
import AlbumPhotoPage from './pages/AlbumPhoto';
import AlbumsPage from './pages/Albums';
import PhotoPage from './pages/Photo';
import ProfilePage from './pages/Profile';

history.listen((location, action) => {
  if (action === 'PUSH') {
    window.scrollTo(0, 0);
  }
});

class AppComponent extends React.Component {
    componentDidMount() {
      this.props.dispatch(galleryActions.getAllCollections());
      this.props.dispatch(galleryActions.getAllPhotos());
      this.props.dispatch(galleryActions.getProfile());
      this.props.dispatch(galleryActions.getConfig());
    }
    render() {
    return(
          <main>
            
            
            <ConnectedRouter history={history}>
              <Header/>
              <div style={{marginTop:"60px"}}>
                <Route path="/" component={IndexPage} exact />
                <Route path="/albums" component={AlbumsPage} />
                <Route path="/photo/:id" component={PhotoPage} />
                <Route path="/album/:id" component={AlbumPhotoPage} />
                <Route path="/about" component={ProfilePage} />
                <Route path="/img/:id" onEnter={() => window.location.reload()} />
              </div>
            </ConnectedRouter>
           
          </main >
      )
    }
  }
const mapStateToProps = (state) =>{return {state}}
let App = connect(mapStateToProps)(AppComponent)

ReactDOM.render(
  <React.StrictMode>
   <Provider store={store()}>
     <App />
   </Provider>
  </React.StrictMode>,
  document.getElementById('root')
);

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://cra.link/PWA
serviceWorkerRegistration.register();

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
