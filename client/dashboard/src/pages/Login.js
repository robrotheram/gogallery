import React from 'react';
import './Login.css'
import { LockOutlined, UserOutlined } from '@ant-design/icons';
import { Input, Button, Alert, Form } from 'antd';
import {  Card } from 'antd';
import { Row, Col } from 'antd';
import { userActions } from '../store/actions';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router';

const Login = () => {
  const [form] = Form.useForm();
  const dispatch = useDispatch()
  const navigate = useNavigate()

  const {loginFailed} = useSelector(state => state.UserReducer);
  
  const changeRoute = (values) => {
    console.log("Stopping Reater")
    dispatch(userActions.login(values.username, values.password, navigate))
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
                  value={"admin"}
                />
            </Form.Item>
            <Form.Item name="password" rules={[{ required: true, message: 'Please input your Password!' }]}>
                <Input
                  prefix={<LockOutlined style={{ color: 'rgba(255,255,255,.25)' }} />}
                  type="password"
                  placeholder="Password"
                  value={"hkLRgDJn"}
                />
            </Form.Item>
            {loginFailed && (
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

export default Login