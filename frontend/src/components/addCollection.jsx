import React from 'react';
import { Modal, Input, TreeSelect, Form } from 'antd';
import {galleryActions, collectionActions} from '../store/actions'
import { useDispatch, useSelector } from 'react-redux';

const AddCollection = () => {
  const  {addCollectionModalVisable} = useSelector(state => state.GalleryReducer);
  const {collections} = useSelector(state => state.CollectionsReducer)

  const [form] = Form.useForm();
  const dispatch = useDispatch()

  const handleCancel = () => {
    dispatch(galleryActions.hideAdd())
  };

  const handleCreate = () => {    
    form.validateFields().then(values => {
        console.log('Received values of form: ', values);
        form.resetFields();
        dispatch(collectionActions.create(values))
        dispatch(galleryActions.hideAdd())
    })
    .catch(err => {
        console.log("error:",err)
    })
  };
  
  return (
    <Modal
      open={addCollectionModalVisable}
      title="Create a new collection"
      okText="Create"
      onCancel={handleCancel}
      onOk={handleCreate}
    >
      <Form form={form} layout="vertical">
        <Form.Item label="Choose collection" hasFeedback name="id" rules={[{ required: false, message: 'Please select the collection to upload photos to!' }]}>
              <TreeSelect
                  treeData={collections}
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
