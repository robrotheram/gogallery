import React, { useState } from 'react';
import { ContainerOutlined, DeleteOutlined } from '@ant-design/icons';
import { Modal, Button, TreeSelect,Form } from 'antd';

import { connect } from 'react-redux';
import { collectionActions } from '../store/actions';

const { confirm } = Modal;
const ButtonGroup = Button.Group;



const MoveModal = (props) => {
  const [visible, setVisable] = useState(false)
  const [form] = Form.useForm();

  const showModal = () => {
    setVisable(true)
  };

  const handleCancel = () => {
    setVisable(true)
  };

  const showDeleteConfirm = () =>{
    confirm({
      title: 'Are you sure delete these photos?',
      content: '',
      okText: 'Yes',
      okType: 'danger',
      cancelText: 'No',
      onOk() {
        props.selectedPhotos.forEach(photo => {
          props.dispatch(collectionActions.remove(photo.id))
        })
        
        console.log('OK');
      },
      onCancel() {
        console.log('Cancel');
      },
    });
  }

  const handleCreate = () => {
    form.validateFields().then(values => {
      values["photos"] = props.selectedPhotos
      console.log('Received values of Move form: ', values);
      props.dispatch(collectionActions.move(values))
      setVisable(false)
    })
    .catch(err => {})
  };

  return (
    <div>
      <ButtonGroup style={{ float: "left" }}>
          <Button onClick={showModal}><ContainerOutlined />move</Button>
          <Button onClick={showDeleteConfirm} ><DeleteOutlined />delete</Button>
      </ButtonGroup>
      <Modal
        visible={visible}
        title="Move photos to collections"
        okText="Move Photos"
        onCancel={handleCancel}
        onOk={handleCreate}
      >
        <Form form={form} layout="vertical">
          <Form.Item label="Choose collection" 
            hasFeedback 
            name="album"
            rules={[{ required: true, message: 'Please select the collection to upload photos to!' }]}
          >
              <TreeSelect
                  treeData={props.collections}
                  placeholder="Select Collection"
                />
          </Form.Item>
        </Form>
      </Modal>
    </div>
    );
}

const mapToProps = (state) =>{
  const photos = state.PhotoReducer.photos;
  const dates = state.CollectionsReducer.dates
  const uploadDates = state.CollectionsReducer.uploadDates
  const collections = state.CollectionsReducer.collections
  return {
    photos,
    dates,
    collections,
    uploadDates
  };
}
export default connect(mapToProps)(MoveModal)
