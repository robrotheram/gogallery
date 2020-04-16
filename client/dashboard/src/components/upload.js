import React from 'react';
import { Modal, Icon, Form, TreeSelect } from 'antd';
import {config, formatTree} from "../store";
import { Upload } from 'antd';
import axios from "axios"

import { connect } from 'react-redux';
import {galleryActions, collectionActions, getOptions} from '../store/actions'

const CollectionCreateForm = Form.create({ name: 'form_in_modal' })(
  // eslint-disable-next-line
  class extends React.Component {
    constructor(){
      super()
      this.state = {
        enableUpload: false,
        collection: ""
      }
    }
    normFile = e => {
      console.log('Upload event:', e);
      if (Array.isArray(e)) {
        return e;
      }
      return e && e.fileList;
    };

    enableUpload = (collection) =>{
      this.setState({enableUpload: true, collection: collection})
    }

    customRequest = (options ) => {
			const data= new FormData()
      data.append('file', options.file)
      let config = getOptions();
      config.headers["content-type"] = 'multipart/form-data; boundary=----WebKitFormBoundaryqTqJIxvkWFYqvP5s';
			
			axios.post(options.action, data, config).then((res) => {
				options.onSuccess(res.data, options.file)
			}).catch((err) => {
				console.log(err)
			})
			
    }
    
    
    render() {
      const { visible, onCancel, onCreate, form } = this.props;
      const { getFieldDecorator } = form;

      formatTree(this.props.collections)
      const collections = Object.values(this.props.collections)
      return (
        <Modal
          visible={visible}
          title="Upload New Photos"
          okText="Upload"
          onCancel={onCancel}
          onOk={onCreate}
        >
          <Form layout="vertical">
            <Form.Item label="Choose collection" hasFeedback>
              {getFieldDecorator('select', {
                rules: [{ required: true, message: 'Please select the collection to upload photos to!' }],
              })(
                <TreeSelect
                    treeData={collections}
                    placeholder="Select Collection"
                    onChange={this.enableUpload}
                  />
              )}
          </Form.Item>
            <Form.Item>
          {getFieldDecorator('photos', {
            valuePropName: 'fileList',
            getValueFromEvent: this.normFile,
          })(
            <Upload.Dragger name="files" customRequest={this.customRequest} action={config.baseUrl+"/collection/uploadFile"}  multiple={true}  listType={'picture'} disabled={!this.state.enableUpload}>
              <p className="ant-upload-drag-icon">
                <Icon type="inbox" />
              </p>
          <p className="ant-upload-text">Click or drag file to this area to upload </p>
              <p className="ant-upload-hint">Support for a single or bulk upload.</p>
            </Upload.Dragger>,
          )}
        </Form.Item>

          </Form>
        </Modal>
      );
    }
  },
);

class UploadCollection extends React.Component {
  handleCancel = () => {
    this.props.dispatch(galleryActions.hideUpload())
  };

  handleCreate = () => {
    const { form } = this.formRef.props;
    form.validateFields((err, values) => {
      if (err) {
        return;
      }
      console.log('Received values of form: ', values);
      form.resetFields();

      this.props.dispatch(collectionActions.upload({
        album: values.select,
        photos: values.photos.map(a => a.name)
      }))
      this.props.dispatch(galleryActions.hideUpload())
    });
  };

  saveFormRef = formRef => {
    this.formRef = formRef;
  };

  render() {
    return (
        <CollectionCreateForm
          wrappedComponentRef={this.saveFormRef}
          visible={this.props.uploadModalVisable}
          onCancel={this.handleCancel}
          onCreate={this.handleCreate}
          collections ={this.props.collections}
        />
    );
  }
}


const mapToProps = (state) =>{
  const uploadModalVisable = state.GalleryReducer.uploadModalVisable;
  const collections = state.CollectionsReducer.collections
  return {uploadModalVisable, collections};
}
export default connect(mapToProps)(UploadCollection)

