import React from 'react';
import {
    Form,
    Input,
    Tooltip,
    Icon,
    Cascader,
    Select,
    Row,
    Col,
    Divider,
    Button,
    AutoComplete,
  } from 'antd';
  import { connect } from 'react-redux';
import { settingsActions } from '../../store/actions';


  const { Option } = Select;
  const AutoCompleteOption = AutoComplete.Option;

  

  class ProfileForm extends React.Component {
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
          this.props.dispatch(settingsActions.setProfile(values))
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
      const { autoCompleteResult } = this.state;
  
    

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
      
      const prefixSelector = getFieldDecorator('prefix', {
        initialValue: '86',
      })(
        <Select style={{ width: 70 }}>
          <Option value="86">+86</Option>
          <Option value="87">+87</Option>
        </Select>,
      );
  
      const websiteOptions = autoCompleteResult.map(website => (
        <AutoCompleteOption key={website}>{website}</AutoCompleteOption>
      ));
      console.log("REGISTRATION STATWE",this.props)
      return (
        <div>
          <Form {...formItemLayout} onSubmit={this.handleSubmit}>
            <Divider>About</Divider>
            <Form.Item label="Profile Photo">
            {getFieldDecorator('ProfilePhoto', {
            })(<Input />)}
            </Form.Item>
            <Form.Item label="Background About Photo">
            {getFieldDecorator('BackgroundPhoto', {
            })(<Input />)}
            </Form.Item>
            <Form.Item label="Description">
            {getFieldDecorator('Description', {
            })(<Input />)}
            </Form.Item>
            <Form.Item label="Footer Text">
            {getFieldDecorator('Footer', {
            })(<Input />)}
            </Form.Item>
            <Divider>Social</Divider>
            <p style={{textAlign:"center"}}>Leave black to disable</p>
            <Form.Item label="Twitter">
            {getFieldDecorator('Twitter', {
            })(<Input />)}
            </Form.Item>
            <Form.Item label="Instagram">
            {getFieldDecorator('Instagram', {
            })(<Input />)}
            </Form.Item>
            <Form.Item label="Website">
            {getFieldDecorator('Website', {
            })(<Input />)}
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
  console.log("SETTINGS:",state.SettingsReducer);
  const settings = state.SettingsReducer.profile
  return {
    settings
  };
}

export default connect(mapToProps)(Form.create({ name: 'register', mapPropsToFields(props) {
  return {
    ProfilePhoto: Form.createFormField({...props.username, value: props.settings.ProfilePhoto}),
    BackgroundPhoto: Form.createFormField({...props.username, value: props.settings.BackgroundPhoto}),
    Description: Form.createFormField({...props.username, value: props.settings.Description}),
    Footer: Form.createFormField({...props.username, value: props.settings.Footer}),
    Twitter: Form.createFormField({...props.username, value: props.settings.Twitter}),
    Instagram: Form.createFormField({...props.username, value: props.settings.Instagram}),
    Website: Form.createFormField({...props.email, value: props.settings.Website}),
  };
}})(ProfileForm));
