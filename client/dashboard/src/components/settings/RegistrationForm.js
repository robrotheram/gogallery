import React, {useState, useEffect } from 'react';
import { Form } from 'antd';
import { Input, Divider, Button } from 'antd';
import { connect } from 'react-redux';
import { userActions } from '../../store/actions';

const formItemLayout = {
  labelCol: {
    xs: { span: 24 },
    sm: { span: 5 },
  },
  wrapperCol: {
    xs: { span: 24 },
    sm: { span: 19 },
  },
};
const tailFormItemLayout = {
  wrapperCol: {
    xs: {
      span: 24,
      offset: 0,
    },
    sm: {
      span: 8,
      offset: 8,
    },
  },
};


const RegistrationForm = (props) => {
  const [confirmDirty, setConfirmDirty] = useState(false)
  const [form] = Form.useForm();

  useEffect(() => {
    console.log("SETTINGS:", props.auth);
    form.setFieldsValue({
      username: props.auth.username,
      email: props.auth.email,
    });
  }, [form, props.auth]);

  const handleSubmit = () => {
  
    form.validateFields().then(values => {
      console.log('Received values of form: ', values);
      if (values.password !== undefined) {
        props.dispatch(userActions.update({
          username: values.username,
          email: values.email,
          password: values.password
        }))
      } else {
        props.dispatch(userActions.update({
          username: values.username,
          email: values.email
        }))
      }
    });
  };

  const handleConfirmBlur = e => {
    const { value } = e.target;
    setConfirmDirty(confirmDirty || !!value)
  };




return (
  <Form form={form} {...formItemLayout} onFinish={handleSubmit}>
    < Divider>{props.auth.username}</Divider>
    <Form.Item label="Username"
      name='username'
      rules={[
        {
          required: true,
          message: 'Please input your username',
        },
      ]}
    ><Input />
    </Form.Item>
    <Form.Item label="E-mail" name='email' rules={[
      {
        type: 'email',
        message: 'The input is not valid E-mail!',
      },
      {
        required: true,
        message: 'Please input your E-mail!',
      },
    ]}><Input />
    </Form.Item>
    <Form.Item label="Password" hasFeedback
      name='password'
      rules={[
        {
          required: false,
          message: 'Please input your password!',
        }
      ]}><Input.Password />
    </Form.Item>
    <Form.Item label="Confirm Password" hasFeedback
      name='confirm'
      dependencies={['password']}
      rules={[
        {
          required: false,
          message: 'Please confirm your password!',
        },
        ({ getFieldValue }) => ({
          validator(_, value) {
            if (!value || getFieldValue('password') === value) {
              return Promise.resolve();
            }
            return Promise.reject(new Error('The two passwords that you entered do not match!'));
          },
        }),
      ]
      }><Input.Password onBlur={handleConfirmBlur} />
    </Form.Item>
    <Divider />
    <Form.Item {...tailFormItemLayout}>
      <Button type="primary" htmlType="submit" style={{ width: "100%" }}>
        Save
          </Button>
    </Form.Item>
  </Form>
)

}

const mapToProps = (state) => {
  console.log("REG", state.UserReducer);
  const auth = state.UserReducer;
  const photos = state.PhotoReducer.photos;
  return {
    auth,
    photos
  };
}

export default connect(mapToProps)(RegistrationForm)
