import './index.less';
import React from 'react';
import ReactDOM from 'react-dom';

import * as serviceWorker from './serviceWorker';
import { Provider } from 'react-redux';

import { history } from './store'
import store from './store'
import { connect } from 'react-redux';
import {
  Router,
  Route,
  Redirect
} from 'react-router-dom'

import { userActions } from './store/actions';

import Main from './pages/Main';
import Login from './pages/Login';
import Settings from './pages/Settings'


const PrivateRoute = ({ component: Component, ...rest }) => (
  <Route {...rest} render={props => (
    localStorage.getItem('token')
      ? <Component {...props} />
      : <Redirect to={{ pathname: '/login', state: { from: props.location } }} />
  )} />
)


const NoMatch = ({ location }) => (
  <Redirect to="/" />
)

class AppComponent extends React.Component {
  state = {
    timer: null,
    counter: 0
  };
  componentDidMount() {
    let timer = setInterval(this.tick, 20000);
    this.setState({ timer });
  }
  componentWillUnmount() {
    clearInterval(this.state.timer);
  }
  tick = () => {
   
    if(localStorage.getItem('token')){
     // console.log("calling")
      this.props.dispatch(userActions.reauth());
    }
    
  }
  render() {
  return(
        <main>
          <Router history={history}>
            <PrivateRoute path="/" component={Main} exact />
            <Route path="/login" component={Login} />
            <PrivateRoute path="/settings" component={Settings} />
            <Route component={NoMatch} />
          </Router>
        </main >
    )
  }
}
const mapStateToProps = (state) =>{return {state}}
let App = connect(mapStateToProps)(AppComponent)


ReactDOM.render(<Provider store={store()}><App /></Provider>, document.getElementById('app_root'));

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
