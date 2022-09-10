import React, { useState } from 'react';
import { DeleteOutlined, DownloadOutlined, SyncOutlined, UploadOutlined, BuildOutlined } from '@ant-design/icons';
import { Progress, message, Row, Col, Divider, Upload, Button, Modal } from 'antd';
import { useDispatch } from 'react-redux';
import { taskActions } from '../../store/actions';
import {config} from "../../store";
import { TaskViewer } from './TaskViewer';
const { confirm } = Modal;

const Maintenance = () => {
  const dispatch = useDispatch();
  const [fileList, setFileList] = useState([])
  
  const showResacanConfirm = () => {
    confirm({
      title: 'Are you sure you want to run the rescan task?',
      okText: 'Yes',
      okType: 'danger',
      cancelText: 'No',
      onOk() {
        dispatch(taskActions.rescan())
      },
      onCancel() {
        console.log('Cancel');
      },
    });
  }

  const showPurgeConfirm = () => {
    confirm({
      title: 'Are you sure you want to Delete the Database?',
      okText: 'Yes',
      okType: 'danger',
      cancelText: 'No',
      onOk() {
        dispatch(taskActions.purge())
      },
      onCancel() {
        console.log('Cancel');
      },
    });
  }
  const showClearConfirm = () => {
    confirm({
      title: 'Are you sure you want to run the clear task?',
      okText: 'Yes',
      okType: 'danger',
      cancelText: 'No',
      onOk() {
        dispatch(taskActions.clear())
      },
      onCancel() {
        console.log('Cancel');
      },
    });
  }
  
  const showClearTemplateCache = () => {
    confirm({
      title: 'Are you sure you want to run the clear template cache task?',
      okText: 'Yes',
      okType: 'danger',
      cancelText: 'No',
      onOk() {
        dispatch(taskActions.templateCacheClear())
      },
      onCancel() {
        console.log('Cancel');
      },
    });
  }

  const showBuildSite = () => {
    confirm({
      title: 'Are you sure you want to run the build site?',
      okText: 'Yes',
      okType: 'success',
      cancelText: 'No',
      onOk() {
        dispatch(taskActions.templateBuild())
      },
      onCancel() {
        console.log('Cancel');
      },
    });
  }

  const showDeploySite  = () => {
    confirm({
      title: 'Are you sure you want to run the deploy site?',
      okText: 'Yes',
      okType: 'success',
      cancelText: 'No',
      onOk() {
        dispatch(taskActions.templateDeploy())
      },
      onCancel() {
        console.log('Cancel');
      },
    });
  }
  
  const props = {
    name: 'file',
    action: config.baseUrl+"/tasks/upload",
    headers: {Authorization:localStorage.getItem('token')},
    className:"maintenance",
    
    onChange(info) {
      if (info.file.status !== 'uploading') {
        console.log(info.file, info.fileList);
      }
      if (info.file.status === 'done') {
        message.success(`${info.file.name} file uploaded successfully`);
        info.fileList=[]
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
      <Row gutter={[16,16]}>
        <Col span={12}>
          <Button type="default" icon={<SyncOutlined />} size="large" style={{"width":"100%"}} onClick={showResacanConfirm}> Resacan image folder </Button>
        </Col>
        <Col span={12}>
          <Button type="default" icon={<DeleteOutlined />} size="large" style={{"width":"100%"}} onClick={showPurgeConfirm} > Purge cache </Button><br/>
        </Col>
        <Col span={12}>
          <Button type="default" icon={<DownloadOutlined />} size="large" style={{"width":"100%"}} onClick={() => dispatch(taskActions.backup())}> Backup Database </Button><br/>
        </Col>
        <Col span={12}>
          <Upload {...props} fileList={fileList} >
            <Button type="default" size="large" style={{"width":"100%"}}>
              <UploadOutlined /> Restore Database
            </Button>
          </Upload>
        </Col>
        <Col span={12}>
          <Button type="default" icon={<BuildOutlined />} size="large" style={{"width":"100%"}}  onClick={()=>showBuildSite()}> Build Site </Button>
        </Col>
        <Col span={12}>
          <Button type="default" icon={<BuildOutlined />} size="large" style={{"width":"100%"}}  onClick={()=>showDeploySite()}> Deploy Site </Button>
        </Col>
      </Row>
      <Divider>Tasks</Divider>
      <TaskViewer/>

    </>
  );
}


export default Maintenance;