import React, { useState } from 'react';
import { DeleteOutlined, DownloadOutlined, SyncOutlined, UploadOutlined, BuildOutlined } from '@ant-design/icons';
import { message, Row, Col, Divider, Upload, Button, Modal } from 'antd';
import { useDispatch } from 'react-redux';
import { taskActions } from '../../store/actions';
import { config } from "../../store";
import { TaskViewer } from './TaskViewer';


const Maintenance = () => {
  const dispatch = useDispatch();
  const [fileList, setFileList] = useState([])

  const props = {
    name: 'file',
    action: config.baseUrl + "/tasks/upload",
    headers: { Authorization: localStorage.getItem('token') },
    className: "maintenance",

    onChange(info) {
      if (info.file.status !== 'uploading') {
        console.log(info.file, info.fileList);
      }
      if (info.file.status === 'done') {
        message.success(`${info.file.name} file uploaded successfully`);
        info.fileList = []
      } else if (info.file.status === 'error') {
        message.error(`${info.file.name} file upload failed.`);
      }
      let fileList = [...info.fileList];
      fileList = fileList.slice(-2);
      fileList = fileList.map(file => {
        if (file.response) {
          file.url = file.response.url;
        }
        return file;
      });
      setFileList(fileList);
    },
  };

  return (
    <>
      <Row gutter={[16, 16]}>
        <Col span={12}>
          <Button type="default" icon={<SyncOutlined />} size="large" style={{ "width": "100%" }} onClick={() => { dispatch(taskActions.rescan()) }}> Resacan image folder </Button>
        </Col>
        <Col span={12}>
          <Button type="default" icon={<DeleteOutlined />} size="large" style={{ "width": "100%" }} onClick={() => { dispatch(taskActions.purge()) }} > Delete Site </Button><br />
        </Col>
        <Col span={12}>
          <Button type="default" icon={<DownloadOutlined />} size="large" style={{ "width": "100%" }} onClick={() => dispatch(taskActions.backup())}> Backup Database </Button><br />
        </Col>
        <Col span={12}>
          <Upload {...props} fileList={fileList} >
            <Button type="default" size="large" style={{ "width": "100%" }}>
              <UploadOutlined /> Restore Database
            </Button>
          </Upload>
        </Col>
        <Col span={12}>
          <Button type="default" icon={<BuildOutlined />} size="large" style={{ "width": "100%" }} onClick={() => dispatch(taskActions.templateBuild())}> Build Site </Button>
        </Col>
        <Col span={12}>
          <Button type="default" icon={<BuildOutlined />} size="large" style={{ "width": "100%" }} onClick={() => dispatch(taskActions.templateDeploy())}> Deploy Site </Button>
        </Col>
      </Row>
      <Divider>Tasks</Divider>
      <TaskViewer />

    </>
  );
}


export default Maintenance;