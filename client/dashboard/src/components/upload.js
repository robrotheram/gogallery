import React from 'react';
import { InboxOutlined } from '@ant-design/icons';
import { Form } from 'antd';
import { Modal, TreeSelect } from 'antd';
import {config} from "../store";
import axios from "axios"
import {notify} from '../store/actions';
import { useDispatch, useSelector } from 'react-redux';
import {galleryActions, collectionActions, getOptions} from '../store/actions'
import { useState } from 'react';

import { Upload, message } from 'antd';
const { Dragger } = Upload;

const UploadCollection = () => {
  const  {uploadModalVisable} = useSelector(state => state.GalleryReducer);
  const {collections} = useSelector(state => state.CollectionsReducer)

  const [upload, setEnableUpload] = useState(false)
  const [files, setFiles] = useState([])
  
  const [form] = Form.useForm();
  const dispatch = useDispatch()

  const enableUpload = () =>{
    setEnableUpload(true);
  }

  const customRequest = (options ) => {
    const data= new FormData()
    console.log("FILE UPLOAD", options);
    data.append('file', options.file)
    let config = getOptions();
    config.headers["content-type"] = 'multipart/form-data; boundary=----WebKitFormBoundaryqTqJIxvkWFYqvP5s';
    
    axios.post(options.action, data, config).then((res) => {
      options.onSuccess(res.data, options.file)
      message.success(`${options.file.name} file uploaded successfully.`);
      setFiles(files => [...files, options.file.name]);
    }).catch(() => {
      message.error(`${options.file.name} file upload failed.`);
    })
  }
  const onCancel = () => {
    dispatch(galleryActions.hideUpload())
  }

  const onCreate = () => {
    form.validateFields().then(values => {
      console.log('Received values of form: ', values, files);
      form.resetFields();
      dispatch(collectionActions.upload({
        album: values.select,
        photos: files
      }))
      dispatch(galleryActions.hideUpload())

    }).catch(() => {
      notify("warning", "Invaild data, could not upload")
      return;
    })
  }

  return (
      <Modal
        visible={uploadModalVisable}
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
                disabled={!upload}
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
export default UploadCollection

