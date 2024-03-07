import React, { useEffect } from 'react';
import { Form, Switch }  from 'antd';
import { Input, Select, InputNumber, Divider, Button } from 'antd';
import EditableTagGroup from './EditableTagGroup';

import { useDispatch, useSelector } from 'react-redux';
import { settingsActions } from '../../store/actions';

const { Option } = Select;
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


const DeploymentForm = () => {
  const [form] = Form.useForm();
  const deploy = useSelector(state => state.SettingsReducer.deploy);
  const dispatch = useDispatch();

  useEffect(() => {
    console.log("DEPLOY:",deploy);
		form.setFieldsValue({
      SiteId:  deploy.SiteId,
      AuthToken:  deploy.AuthToken,
      Draft: deploy.Draft
    });
	}, [form,deploy]);

  const handleSubmit = () => {
    
    form.validateFields().then(values => {
      console.log('Received values of form: ', values);
      dispatch(settingsActions.setDeploy(values))
    });
  };

    return (
      <Form form={form} {...formItemLayout} onFinish={handleSubmit}>
        <Form.Item label="Site Id" name='SiteId'><Input /></Form.Item>
        <Form.Item label="Auth Token" name='AuthToken'><Input /></Form.Item>
        <Form.Item label="Draft" name='Draft'><Switch /></Form.Item>
        <Divider/>
        <Form.Item {...tailFormItemLayout}>
          <Button type="primary" htmlType="submit" style={{width:"100%"}}>
            Save
          </Button>
        </Form.Item>
        
      </Form>
    )
}

export default DeploymentForm