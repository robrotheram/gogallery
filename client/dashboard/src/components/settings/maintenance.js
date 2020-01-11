import React from 'react';
import {
    Form,
    Input,
    Tooltip,
    Icon,
    Cascader,
    message,
    Row,
    Col,
    Upload,
    Button,
    Modal,
  } from 'antd';
  import { connect } from 'react-redux';
  import { taskActions } from '../../store/actions';
  import {config} from "../../store";
  const { confirm } = Modal;
  class Maintenance extends React.Component {

    state = {
      fileList: [
      ],
    };
    
    showResacanConfirm = () => {
      let _this = this;
      confirm({
        title: 'Are you sure you want to run the rescan task?',
        okText: 'Yes',
        okType: 'danger',
        cancelText: 'No',
        onOk() {
          _this.props.dispatch(taskActions.purge())
        },
        onCancel() {
          console.log('Cancel');
        },
      });
    }
    showClearConfirm = () => {
      let _this = this;
      confirm({
        title: 'Are you sure you want to run the clear task?',
        okText: 'Yes',
        okType: 'danger',
        cancelText: 'No',
        onOk() {
          _this.props.dispatch(taskActions.clear())
        },
        onCancel() {
          console.log('Cancel');
        },
      });
    }
    
    render() {
      const {dispatch} = this.props
      let _this = this;
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
          _this.setState({ fileList });
        },
      };
      return (
        <Row gutter={[16,16]}>
          <Col span={12}>
            <Button type="default" icon="sync" size="large" style={{"width":"100%"}} onClick={this.showResacanConfirm}> Resacan image folder </Button>
          </Col>
          <Col span={12}>
            <Button type="default" icon="delete" size="large" style={{"width":"100%"}} onClick={this.showClearConfirm} > Clear cache </Button><br/>
          </Col>
          <Col span={12}>
            <Button type="default" icon="download" size="large" style={{"width":"100%"}} onClick={() => dispatch(taskActions.backup())}> Backup Database </Button><br/>
          </Col>
          <Col span={12}>
          <Upload {...props} fileList={this.state.fileList}>
            <Button type="default" size="large" style={{"width":"100%"}}>
              <Icon type="upload" /> Restore Database
            </Button>
          </Upload>
            </Col>
        </Row>
      );
    }
  }

const mapToProps = (state) =>{
  console.log("SETTINGS:",state.SettingsReducer);
  const settings = state.SettingsReducer.gallery
  return {
    settings
  };
}
export default connect(mapToProps)(Maintenance);