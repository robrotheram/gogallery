import React, { useState } from 'react';
import { ContainerOutlined, DeleteOutlined } from '@ant-design/icons';
import { Modal, Button, TreeSelect,Form } from 'antd';

import { connect } from 'react-redux';
import { collectionActions } from '../store/actions';

const { confirm } = Modal;
const ButtonGroup = Button.Group;



const MoveModal = (props) => {
  const normFile = e => {
    console.log('Upload event:', e);
    if (Array.isArray(e)) {
      return e;
    }
    return e && e.fileList;
  };
  
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
  


{/* class MoveModal extends React.Component {
  constructor(){
    super();
    this.state = {visible: false}
  }
  handleCancel = () => {
    this.setState({visible: false})
  };

  handleCreate = () => {
    const { form } = this.formRef.props;
    form.validateFields((err, values) => {
      if (err) {
        return;
      }
      
      values["photos"] = this.props.selectedPhotos
      form.resetFields();
      console.log('Received values of form: ', values);
      this.props.dispatch(collectionActions.move(values))
      this.setState({visible: false})
    });
  };

  saveFormRef = formRef => {
    this.formRef = formRef;
  };

  showModal = () => {
    this.setState({
      visible: true,
    });
  };

  render() {
    let _this = this
    function showDeleteConfirm() {
      confirm({
        title: 'Are you sure delete these photos?',
        content: '',
        okText: 'Yes',
        okType: 'danger',
        cancelText: 'No',
        onOk() {
          _this.props.selectedPhotos.forEach(photo => {
            _this.props.dispatch(collectionActions.remove(photo.id))
          })
          
          console.log('OK');
        },
        onCancel() {
          console.log('Cancel');
        },
      });
    }
    return (
      <div>
            <ButtonGroup style={{ float: "left" }}>
                <Button onClick={this.showModal}><ContainerOutlined />move</Button>
                <Button onClick={showDeleteConfirm} ><DeleteOutlined />delete</Button>
              </ButtonGroup>

        <CollectionCreateForm
          wrappedComponentRef={this.saveFormRef}
          visible={this.state.visible}
          onCancel={this.handleCancel}
          onCreate={this.handleCreate}
          collections ={this.props.collections}
        />
        </div>
    );
  }
} */}
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
