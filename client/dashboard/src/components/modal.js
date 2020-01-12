import React from 'react';
import { Modal, Button, Icon, Form, Select } from 'antd';

import { connect } from 'react-redux';
import { collectionActions } from '../store/actions';



const { Option } = Select;
const { confirm } = Modal;
const ButtonGroup = Button.Group;



const CollectionCreateForm = Form.create({ name: 'form_in_modal' })(
  // eslint-disable-next-line
  class extends React.Component {
    normFile = e => {
      console.log('Upload event:', e);
      if (Array.isArray(e)) {
        return e;
      }
      return e && e.fileList;
    };
    render() {
      const { visible, onCancel, onCreate, form } = this.props;
      const { getFieldDecorator } = form;
      return (
        <Modal
          visible={visible}
          title="Move photos to collections"
          okText="Move Photos"
          onCancel={onCancel}
          onOk={onCreate}
        >
          <Form layout="vertical">
            <Form.Item label="Choose collection" hasFeedback>
              {getFieldDecorator('album', {
                rules: [{ required: true, message: 'Please select the collection to upload photos to!' }],
              })(
                <Select placeholder="Please select a collection">
                  {this.props.collections.map((el, index) => (<Option key={el.name}>{el.name}</Option> ))}
                </Select>,
              )}
            </Form.Item>
          </Form>
        </Modal>
      );
    }
  },
);

class MoveModal extends React.Component {
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
          _this.props.dispatch(collectionActions.move({
            album: "rubish",
            photos: _this.props.selectedPhotos
          }))
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
                <Button onClick={this.showModal}><Icon type="container" />move</Button>
                <Button onClick={showDeleteConfirm} ><Icon type="delete" />delete</Button>
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
