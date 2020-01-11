import React from 'react';
import './Login.css'
import { Form, Icon, Input, Button, Alert, Divider } from 'antd';
import {  Card } from 'antd';
import { Row, Col } from 'antd';

import {
  BrowserRouter as Router,
  Route,
  Link,
  Redirect,
  withRouter
} from 'react-router-dom'

import { connect } from 'react-redux';
import { userActions } from '../store/actions';

const CollectionCreateForm = Form.create({ name: 'form_in_modal' })(
class extends React.Component {
  handleSubmit = e => {
    e.preventDefault();
    this.props.form.validateFields((err, values) => {
      if (!err) {
        console.log('Received values of form: ', values);
        const { dispatch } = this.props;
        if (values.username && values.password) {
          console.log('sending: ', values);
          this.props.onDone(values)
        }
      }
    });
  };

  render() {
    const { getFieldDecorator } = this.props.form;
    return (
      <Form onSubmit={this.handleSubmit} className="login-form">
        <Form.Item>
          {getFieldDecorator('username', {
            rules: [{ required: true, message: 'Please input your username!' }],
          })(
            <Input
              prefix={<Icon type="user" style={{ color: 'rgba(255,255,255,.25)' }} />}
              placeholder="Username"
            />,
          )}
        </Form.Item>
        <Form.Item>
          {getFieldDecorator('password', {
            rules: [{ required: true, message: 'Please input your Password!' }],
          })(
            <Input
              prefix={<Icon type="lock" style={{ color: 'rgba(255,255,255,.25)' }} />}
              type="password"
              placeholder="Password"
            />,
          )}
        </Form.Item>
        {this.props.loginFailed && (
         <div><br/>
         <Alert message="Login Error: Invalid username/password" type="error" showIcon closable style={{"backgroundColor":"#141414", textAlign:"left"}} />
         <br/>
         </div>
      )}
        <Form.Item>
          <Button type="primary" htmlType="submit" className="login-form-button">
            Log in
          </Button>
        </Form.Item>
      </Form>
    );
  }
}
);

class Login extends React.Component {
  state = {
    redirectToReferrer: false
  }

  changeRoute = (values) => {
    console.log("Stopping Reater")
    this.props.dispatch(userActions.login(values.username, values.password));
  }

  render() {
    return (
        <Row type="flex" justify="center" align="middle" style={{minHeight: '100vh'}}>
            <Col>
            <Card  title={"GoGallery Dashbard"} style={{textAlign: "center"}}> 
        <CollectionCreateForm
          login ={this.props.login}
          loginFailed = {this.props.loginFailed}
          onDone = {this.changeRoute}
        />
        </Card>
        </Col>
        </Row>
    );
  }
}

const mapStateToProps = (state) =>{
  console.log(state)
  const { loggingIn,  loginFailed} = state.UserReducer;
  return {
      loggingIn,
      loginFailed
  };
}

export default withRouter(connect(mapStateToProps)(Login));
