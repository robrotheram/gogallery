import React, {useEffect} from 'react';
import { Form } from "antd";
import { Input, Divider, Button } from 'antd';
import { connect } from 'react-redux';
import { settingsActions } from '../../store/actions';

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

const  ProfileForm = (props) =>  {
  const [form] = Form.useForm();
  useEffect(() => {
    console.log("SETTINGS:", props.settings);
    form.setFieldsValue({
      ProfilePhoto:  props.settings.ProfilePhoto,
      BackgroundPhoto:  props.settings.BackgroundPhoto,
      Description:  props.settings.Description,
      Footer:  props.settings.Footer,
      Twitter:  props.settings.Twitter,
      Instagram:  props.settings.Instagram,
      Website: props.settings.Website,
    });
  }, [form, props.settings]);




  const handleSubmit = e => {
    e.preventDefault();
    form.validateFieldsAndScroll((err, values) => {
      if (!err) {
        console.log('Received values of form: ', values);
        props.dispatch(settingsActions.setProfile(values))
      }
    });
  };



    return (
      <div>
        <Form form={form}  {...formItemLayout} onSubmit={handleSubmit}>
          <Divider>About</Divider>
          <Form.Item label="Profile Photo" name='ProfilePhoto'><Input /></Form.Item>
          <Form.Item label="Background About Photo" name='BackgroundPhoto'><Input /></Form.Item>
          <Form.Item label="Description" name='Description'><Input /></Form.Item>
          <Form.Item label="Footer Text" name='Footer'><Input /></Form.Item>
          <Divider>Social</Divider>
          <p style={{textAlign:"center"}}>Leave black to disable</p>
          <Form.Item label="Twitter" name='Twitter'><Input /></Form.Item>
          <Form.Item label="Instagram" name='Instagram'><Input /></Form.Item>
          <Form.Item label="Website" name='Website'><Input /></Form.Item>
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
const mapToProps = (state) =>{
  console.log("SETTINGS:",state.SettingsReducer);
  const settings = state.SettingsReducer.profile
  return {
    settings
  };
}

export default connect(mapToProps)(ProfileForm);
