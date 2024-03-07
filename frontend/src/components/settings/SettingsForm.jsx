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


const SettingsForm = () => {
  const [form] = Form.useForm();
  const settings = useSelector(state => state.SettingsReducer.gallery);
  const dispatch = useDispatch();

  useEffect(() => {
    console.log("SETTINGS:",settings);
		form.setFieldsValue({
      Name:  settings.Name,
      Basepath:  settings.Basepath,
      Url:  settings.Url,
      Theme: settings.Theme,
      Destpath: settings.Destpath,
      ImagesPerPage:  settings.ImagesPerPage,
      PictureBlacklist:  settings.PictureBlacklist || [],
      AlbumBlacklist:  settings.AlbumBlacklist || [],
      Renderer: settings.Renderer,
    });
	}, [form,settings]);

  const handleSubmit = () => {
    
    form.validateFields().then(values => {
      console.log('Received values of form: ', values);
      dispatch(settingsActions.setGallery(values))
    });
  };

    return (
      <Form form={form} {...formItemLayout} onFinish={handleSubmit}>
        <Form.Item label="Gallery Name" name='Name'><Input /></Form.Item>
        <Form.Item label="Gallery Url" name='Url'><Input />
        </Form.Item>
        <Form.Item label="Path to build site" name='Destpath'><Input />
        </Form.Item>
        <Form.Item label="Theme Path" name='Theme'><Input />
        </Form.Item>
        <Form.Item label="Image Folder" name='Basepath'><Input />
        </Form.Item>
        <Form.Item label="Use Orginal Image" name='UseOriginal'><Switch />
        </Form.Item>
        <Form.Item label="Image Blacklist" name='PictureBlacklist'><EditableTagGroup  />
        </Form.Item>
        <Form.Item label="Collection Blacklist" name='AlbumBlacklist'><EditableTagGroup/>
        </Form.Item>
        <Divider/>
        <Form.Item {...tailFormItemLayout}>
          <Button type="primary" htmlType="submit" style={{width:"100%"}}>
            Save
          </Button>
        </Form.Item>
        
      </Form>
    )
}

export default SettingsForm