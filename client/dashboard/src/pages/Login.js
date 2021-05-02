import React from 'react';
import './Login.css'
import { LockOutlined, UserOutlined } from '@ant-design/icons';
import { Input, Button, Alert, Form } from 'antd';
import {  Card } from 'antd';
import { Row, Col } from 'antd';
import { withRouter } from 'react-router-dom'
import { connect } from 'react-redux';
import { userActions } from '../store/actions';

const Login = (props) => {
  const [form] = Form.useForm();
  const changeRoute = (values) => {
    console.log("Stopping Reater")
    props.dispatch(userActions.login(values.username, values.password));
  }
  const handleSubmit = () => {
    form.validateFields().then(values => {
        console.log('Received values of form: ', values);
        if (values.username && values.password) {
          console.log('sending: ', values);
          changeRoute(values)
        }
    });
  };
  return (
    <Row type="flex" justify="center" align="middle" style={{minHeight: '100vh'}}>
      <Col span={6}>
        <Card  title={"GoGallery Dashbard"} style={{textAlign: "center"}}> 
          <Form form={form} className="login-form" onFinish={handleSubmit}>
            <Form.Item name="username" rules={[{ required: true, message: 'Please input your username!' }]}>
                <Input
                  prefix={<UserOutlined style={{ color: 'rgba(255,255,255,.25)' }} />}
                  placeholder="Username"
                />
            </Form.Item>
            <Form.Item name="password" rules={[{ required: true, message: 'Please input your Password!' }]}>
                <Input
                  prefix={<LockOutlined style={{ color: 'rgba(255,255,255,.25)' }} />}
                  type="password"
                  placeholder="Password"
                />
            </Form.Item>
            {props.loginFailed && (
              <div>
                <br/>
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
        </Card>
      </Col>
    </Row>
  );
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
