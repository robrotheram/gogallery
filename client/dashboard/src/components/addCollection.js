import React from 'react';
import { Modal, Input, TreeSelect, Form } from 'antd';
import {galleryActions, collectionActions} from '../store/actions'


const AddCollection = (props) => {
  const [form] = Form.useForm();

  const handleCancel = () => {
    props.dispatch(galleryActions.hideAdd())
  };

  const handleCreate = () => {    
    form.validateFields().then(values => {
        console.log('Received values of form: ', values);
        form.resetFields();
        props.dispatch(collectionActions.create(values))
        props.dispatch(galleryActions.hideAdd())
    })
    .catch(err => {
        console.log("error:",err)
    })
  };
  
  return (
    <Modal
      visible={props.addCollectionModalVisable}
      title="Create a new collection"
      okText="Create"
      onCancel={handleCancel}
      onOk={handleCreate}
    >
      <Form form={form} layout="vertical">
        <Form.Item label="Choose collection" hasFeedback name="id" rules={[{ required: true, message: 'Please select the collection to upload photos to!' }]}>
              <TreeSelect
                  treeData={props.collections}
                  placeholder="Select Collection"
                  //onChange={enableUpload}
                />
        </Form.Item>
        <Form.Item label="Collection Name" name="name" rules={[{ required: true, message: 'Please input the name of collection!' }]}>
          <Input />
        </Form.Item>
      </Form>
    </Modal>
  );
}
export default (AddCollection)
