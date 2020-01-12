import React from 'react';
import ReactDOM from 'react-dom';
import './index.css'
import * as serviceWorker from './serviceWorker';
import {
  Router,
  Route
} from 'react-router-dom'
import { Provider } from 'react-redux';
import { history } from './store'
import store from './store'


import Header from './components/header'
import IndexPage from './pages/Index';
import AlbumPhotoPage from './pages/AlbumPhoto';
import AlbumsPage from './pages/Albums';
import PhotoPage from './pages/Photo';
import ProfilePage from './pages/Profile';
import { connect } from 'react-redux';


import { galleryActions } from './store/actions';

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
    }
    render() {
    return(
          <main>
            
            
            <Router history={history}>
              <Header/>
              <div style={{marginTop:"60px"}}>
                <Route path="/" component={IndexPage} exact />
                <Route path="/albums" component={AlbumsPage} />
                <Route path="/photo/:id" component={PhotoPage} />
                <Route path="/album/:id" component={AlbumPhotoPage} />
                <Route path="/about" component={ProfilePage} />
                </div>
            </Router>
           
          </main >
      )
    }
  }
const mapStateToProps = (state) =>{return {state}}
let App = connect(mapStateToProps)(AppComponent)


ReactDOM.render(<Provider store={store()}><App /></Provider>, document.getElementById('root'));

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
