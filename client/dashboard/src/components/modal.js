import React, { useState } from 'react';
import { ContainerOutlined, DeleteOutlined } from '@ant-design/icons';
import { Modal, Button, TreeSelect,Form } from 'antd';

import { useDispatch, useSelector } from 'react-redux';
import { collectionActions } from '../store/actions';

const { confirm } = Modal;
const ButtonGroup = Button.Group;



const MoveModal = ({selectedPhotos}) => {
  const [visible, setVisable] = useState(false)
  const [form] = Form.useForm();
  const dispatch = useDispatch()

  const { photos } = useSelector(state => state.PhotoReducer)
  const {collections} = useSelector(state => state.CollectionsReducer);

  const showModal = () => {
    setVisable(true)
  };

  const handleCancel = () => {
    setVisable(false)
  };

  const showDeleteConfirm = () =>{
    confirm({
      title: 'Are you sure delete these photos?',
      content: '',
      okText: 'Yes',
      okType: 'danger',
      cancelText: 'No',
      onOk() {
        selectedPhotos.forEach(pos => {
          let photo = photos[pos]
          dispatch(collectionActions.remove(photo.id))
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
      values["photos"] = selectedPhotos.map(pos => photos[pos])
      console.log('Received values of Move form: ', values);
      dispatch(collectionActions.move(values))
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
                  treeData={collections}
                  placeholder="Select Collection"
                />
          </Form.Item>
        </Form>
      </Modal>
    </div>
    );
}


export default MoveModal
