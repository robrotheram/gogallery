import React from 'react';
import {
    Form,
    Input,
    Divider,
    Button,
  } from 'antd';
  import { connect } from 'react-redux';
  import { userActions } from '../../store/actions';

  class RegistrationForm extends React.Component {
    state = {
      confirmDirty: false,
      autoCompleteResult: [],
      auth: {username:""}
    };

    // componentDidUpdate(prevProps, prevState) {
    //   if (prevState.auth.username !== this.state.auth.username) {
    //     let auth = prevState.auth
    //     auth.username = nextProps.auth.username
    //     this.setState({auth});
        
    //   }
    // }

    // static getDerivedStateFromProps(nextProps, prevState){
    //   const { auth } = nextProps
    //   console.log("REGISTRATION", auth.username ,  prevState.auth.username)
    //   if (auth.username !== prevState.auth.username ){
    //     return { auth }
    //   }
    //   return null
    // }



  
    handleSubmit = e => {
      e.preventDefault();
      this.props.form.validateFieldsAndScroll((err, values) => {
        if (!err) {
          console.log('Received values of form: ', values);
          if(values.password !== undefined){
            this.props.dispatch(userActions.update({
              username: values.username,
              email: values.email,
              password: values.password
            }))
          }else{
            this.props.dispatch(userActions.update({
              username: values.username,
              email: values.email
            }))
          }
        }
      });
    };
  
    handleConfirmBlur = e => {
      const { value } = e.target;
      this.setState({ confirmDirty: this.state.confirmDirty || !!value });
    };
  
    compareToFirstPassword = (rule, value, callback) => {
      const { form } = this.props;
      
      //if (value && this.state.confirmDirty) {
        form.validateFields(['password'], { force: true });
      
      if (value && value !== form.getFieldValue('password')) {
        callback('Two passwords that you enter is inconsistent!');
      } else {
        callback();
      }
    };
  
    validateToNextPassword = (rule, value, callback) => {
      const { form } = this.props;
      if (value && this.state.confirmDirty) {
        form.validateFields(['confirm'], { force: true });
      }
      if (value && value !== form.getFieldValue('confirm')) {
        callback('Two passwords that you enter is inconsistent!');
      } else {
        callback();
      }
    };
  
    handleWebsiteChange = value => {
      let autoCompleteResult;
      if (!value) {
        autoCompleteResult = [];
      } else {
        autoCompleteResult = ['.com', '.org', '.net'].map(domain => `${value}${domain}`);
      }
      this.setState({ autoCompleteResult });
    };
  
    render() {
      const { getFieldDecorator } = this.props.form;
  
    

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
      
      console.log("REGISTRATION STATWE",this.props)
      return (
        <div>
        <Form {...formItemLayout} onSubmit={this.handleSubmit}>
        < Divider>{this.state.auth.username}</Divider>
          <Form.Item label="Username">
            {getFieldDecorator('username', {
              rules: [
                {
                  required: true,
                  message: 'Please input your username',
                },
              ],
            })(<Input/>)}
          </Form.Item>
          <Form.Item label="E-mail">
            {getFieldDecorator('email', {
              rules: [
                {
                  type: 'email',
                  message: 'The input is not valid E-mail!',
                },
                {
                  required: true,
                  message: 'Please input your E-mail!',
                },
              ],
            })(<Input />)}
          </Form.Item>
          <Form.Item label="Password" hasFeedback>
            {getFieldDecorator('password', {
              rules: [
                {
                  required: false,
                  message: 'Please input your password!',
                },
                {
                  validator: this.validateToNextPassword,
                },
              ],
            })(<Input.Password />)}
          </Form.Item>
          <Form.Item label="Confirm Password" hasFeedback>
            {getFieldDecorator('confirm', {
              rules: [
                {
                  required: false,
                  message: 'Please confirm your password!',
                },
                {
                  validator: this.compareToFirstPassword,
                },
              ],
            })(<Input.Password onBlur={this.handleConfirmBlur} />)}
          </Form.Item>
          <Divider/>
          <Form.Item {...tailFormItemLayout}>
            <Button type="primary" htmlType="submit" style={{width:"100%"}}>
              Save
            </Button>
          </Form.Item>
          </Form>
        </div>

      );
    }
  }
const mapToProps = (state) =>{
  console.log("REG",state.UserReducer);
  const auth = state.UserReducer;
  const photos = state.PhotoReducer.photos;
  return {
    auth,
    photos
  };
}

export default connect(mapToProps)(Form.create({ name: 'register', mapPropsToFields(props) {
  return {
    username: Form.createFormField({...props.username, value: props.auth.username}),
    email: Form.createFormField({...props.email, value: props.auth.email}),
  };
}})(RegistrationForm));
