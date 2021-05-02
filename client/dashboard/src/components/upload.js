import React from 'react';
import { InboxOutlined } from '@ant-design/icons';
import { Form } from 'antd';
import { Modal, TreeSelect } from 'antd';
import {config} from "../store";
import axios from "axios"
import {notify} from '../store/actions';
import { connect } from 'react-redux';
import {galleryActions, collectionActions, getOptions} from '../store/actions'
import { useState } from 'react';

import { Upload, message } from 'antd';
const { Dragger } = Upload;



const UploadCollection = (props) => {
  const {collections } = props;
  const [upload, setEnableUpload] = useState(false)
  const [collection, setCollection] = useState("")
  const [files, setFiles] = useState([])
  const [form] = Form.useForm();

  const normFile = e => {
    console.log('Upload event:', e);
    if (Array.isArray(e)) {
      return e;
    }
    return e && e.fileList;
  };

  const enableUpload = (collection) =>{
    setEnableUpload(true);
    setCollection(collection)
  }

  const customRequest = (options ) => {
    const data= new FormData()
    data.append('file', options.file)
    let config = getOptions();
    config.headers["content-type"] = 'multipart/form-data; boundary=----WebKitFormBoundaryqTqJIxvkWFYqvP5s';
    
    axios.post(options.action, data, config).then((res) => {
      options.onSuccess(res.data, options.file)
      message.success(`${options.file.name} file uploaded successfully.`);
      setFiles([...files, options.file.name])
    }).catch(() => {
      message.error(`${options.file.name} file upload failed.`);
    })
  }
  const onCancel = () => {
    props.dispatch(galleryActions.hideUpload())
  }
  const onCreate = () => {
    form.validateFields().then(values => {
      console.log('Received values of form: ', values);
      form.resetFields();
      props.dispatch(collectionActions.upload({
        album: values.select,
        photos: files
      }))
      props.dispatch(galleryActions.hideUpload())

    }).catch(() => {
      notify("warning", "Invaild data, could not upload")
      return;
    })
  }

  const onChange = (info) => {
    const { status } = info.file;
    if (status !== 'uploading') {
      console.log(info.file, info.fileList);
    }
    if (status === 'done') {
      message.success(`${info.file.name} file uploaded successfully.`);
    } else if (status === 'error') {
      message.error(`${info.file.name} file upload failed.`);
    }
  }
  return (
      <Modal
        visible={props.uploadModalVisable}
        title="Upload New Photos"
        okText="Upload"
        onCancel={onCancel}
        onOk={onCreate}
      >
        <Form form={form} layout="vertical">
          <Form.Item 
            label="Choose collection" 
            hasFeedback
            name="select"
            rules={[{ required: true, message: 'Please select the collection to upload photos to!' }]}
            >
              <TreeSelect
                  treeData={collections}
                  placeholder="Select Collection"
                  onChange={enableUpload}
              />
          </Form.Item>
              <Dragger
                name="files" 
                customRequest={customRequest} 
                action={config.baseUrl+"/collection/uploadFile"}  
                multiple={true}  
                listType={'picture'}
                disabled={!enableUpload}
              >
                <p className="ant-upload-drag-icon">
                  <InboxOutlined />
                </p>
                <p className="ant-upload-text">Click or drag file to this area to upload</p>
                <p className="ant-upload-hint">
                  Support for a single or bulk upload. Strictly prohibit from uploading company data or other
                  band files
                </p>
              </Dragger>
         
        </Form>
      </Modal>
    );
}

// const ssUploadCollection = (props) => {
//   const handleCancel = () => {
//     props.dispatch(galleryActions.hideUpload())
//   };

//   const handleCreate = () => {
    
//     form.validateFields((err, values) => {
//       if (err) {
//         return;
//       }
//       console.log('Received values of form: ', values);
//       form.resetFields();

//       this.props.dispatch(collectionActions.upload({
//         album: values.select,
//         photos: values.photos.map(a => a.name)
//       }))
//       this.props.dispatch(galleryActions.hideUpload())
//     });
//   };


//     return (
//         <CollectionCreateForm
//           visible={this.props.uploadModalVisable}
//           onCancel={this.handleCancel}
//           onCreate={this.handleCreate}
//           collections ={this.props.collections}
//         />
//     );
//   }


const mapToProps = (state) =>{
  const uploadModalVisable = state.GalleryReducer.uploadModalVisable;
  const collections = state.CollectionsReducer.collections
  return {uploadModalVisable, collections};
}
export default connect(mapToProps)(UploadCollection)

