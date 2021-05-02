import React, { useEffect } from 'react';
import { Form }  from 'antd';
import { Input, Select, InputNumber, Divider, Button } from 'antd';
import EditableTagGroup from './EditableTagGroup';

import { connect } from 'react-redux';
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
      span: 19,
      offset: 5,
    },
  },
};



const SettingsForm = (props) => {
  const [form] = Form.useForm();


  useEffect(() => {
    console.log("SETTINGS:",props.settings);
		form.setFieldsValue({
      Name:  props.settings.Name,
      Basepath:  props.settings.Basepath,
      Url:  props.settings.Url,
      ImagesPerPage:  props.settings.ImagesPerPage,
      PictureBlacklist:  props.settings.PictureBlacklist || [],
      AlbumBlacklist:  props.settings.AlbumBlacklist || [],
      Renderer: props.settings.Renderer,
    });
	}, [form,props.settings]);

  const handleSubmit = e => {
    e.preventDefault();
    form.scrollToField((err, values) => {
      if (!err) {
        console.log('Received values of form: ', values);
        props.dispatch(settingsActions.setGallery(values))
      }
    });
  };

    return (
      <Form form={form} {...formItemLayout} onSubmit={handleSubmit}>
        <Form.Item label="Gallery Name" name='Name'><Input /></Form.Item>
        <Form.Item label="Gallery Url" name='Url'><Input />
        </Form.Item>
        <Form.Item label="Images Per Page" name='ImagesPerPage'><InputNumber min={1} max={100} style={{width:"100%"}}/>
        </Form.Item>
        <Form.Item label="Image Folder" name='Basepath'><Input />
        </Form.Item>
        <Form.Item label="Image Blacklist" name='PictureBlacklist'><EditableTagGroup  />
        </Form.Item>
        <Form.Item label="Collection Blacklist" name='AlbumBlacklist'><EditableTagGroup/>
        </Form.Item>
        <Form.Item label="Renderer" name='Renderer'> 
          <Select>
            <Option value="imagemagick">Imagemagick</Option>
            <Option value="navive">Native</Option>
          </Select>
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

const mapToProps = (state) =>{
  const settings = state.SettingsReducer.gallery
  return {
    settings
  };
}

export default connect(mapToProps)(SettingsForm)