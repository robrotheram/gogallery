import React from 'react';
import {
    Form,
    Input,
    Select,
    InputNumber,
    Divider,
    Button,
  } from 'antd';
  import EditableTagGroup from './EditableTagGroup';
 
  import { connect } from 'react-redux';
  import { settingsActions } from '../../store/actions';

  const { Option } = Select;

  class SettingsForm extends React.Component {
    state = {
      confirmDirty: false,
      autoCompleteResult: [],
    };
  
    handleSubmit = e => {
      e.preventDefault();
      this.props.form.validateFieldsAndScroll((err, values) => {
        if (!err) {
          console.log('Received values of form: ', values);
          this.props.dispatch(settingsActions.setGallery(values))
        }
      });
    };
  
    handleConfirmBlur = e => {
      const { value } = e.target;
      this.setState({ confirmDirty: this.state.confirmDirty || !!value });
    };
  
    compareToFirstPassword = (rule, value, callback) => {
      const { form } = this.props;
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
      callback();
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
            span: 19,
            offset: 5,
          },
        },
      };
      return (
        <Form {...formItemLayout} onSubmit={this.handleSubmit}>
          <Form.Item label="Gallery Name">
            {getFieldDecorator('Name', {
            })(<Input />)}
          </Form.Item>
          <Form.Item label="Gallery Url">
            {getFieldDecorator('Url', {
            })(<Input />)}
          </Form.Item>
          <Form.Item label="Images Per Page">
            {getFieldDecorator('ImagesPerPage', {
            })(<InputNumber min={1} max={100} style={{width:"100%"}}/>)}
          </Form.Item>
          <Form.Item label="Image Folder">
            {getFieldDecorator('Basepath', {
            })(<Input />)}
          </Form.Item>
          <Form.Item label="Image Blacklist">
            {getFieldDecorator('PictureBlacklist', {
            })(<EditableTagGroup  />)}
          </Form.Item>
          <Form.Item label="Collection Blacklist">
            {getFieldDecorator('AlbumBlacklist', {
            })(<EditableTagGroup/>)}
          </Form.Item>
          <Form.Item label="Renderer">
            {getFieldDecorator('Renderer', {
            })( <Select>
            <Option value="imagemagick">Imagemagick</Option>
            <Option value="navive">Native</Option>
          </Select>)}
          </Form.Item>
          <Divider/>
          <Form.Item {...tailFormItemLayout}>
            <Button type="primary" htmlType="submit" style={{width:"100%"}}>
              Save
            </Button>
          </Form.Item>
          
        </Form>
      );
    }
  }

const mapToProps = (state) =>{
  console.log("SETTINGS:",state.SettingsReducer);
  const settings = state.SettingsReducer.gallery
  return {
    settings
  };
}
export default connect(mapToProps)(Form.create({ name: 'settings', mapPropsToFields(props) {
  return {
    Name: Form.createFormField({...props.username, value: props.settings.Name}),
    Basepath: Form.createFormField({...props.username, value: props.settings.Basepath}),
    Url: Form.createFormField({...props.username, value: props.settings.Url}),
    ImagesPerPage: Form.createFormField({...props.username, value: props.settings.ImagesPerPage}),
    PictureBlacklist: Form.createFormField({...props.username, value: props.settings.PictureBlacklist || []}),
    AlbumBlacklist: Form.createFormField({...props.username, value: props.settings.AlbumBlacklist || []}),
    Renderer: Form.createFormField({...props.email, value: props.settings.Renderer}),
  };
}})(SettingsForm));