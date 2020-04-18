import React from 'react';
// as an array
import { Layout, Form,  Input,  Select, TreeSelect } from 'antd';
import { Collapse } from 'antd';
import moment from 'moment';

import { connect } from 'react-redux';
import { config, formatTree } from '../store';
import { photoActions } from '../store/actions';


const { Panel } = Collapse;
const { Sider } = Layout;
const { Option } = Select;
const WAIT_INTERVAL = 2000
const ENTER_KEY = 13
const defaultState = {
  meta:{},
  exif:{}
}

class SideBar extends React.PureComponent {

  constructor(props) {
    super(props);
    // if (props.data !== undefined){
    //     this.state = { data : props.data}
    // } else {
      this.state = {data: defaultState}
  }
  timer = null
  componentWillReceiveProps(nextProps) {
    // You don't have to do this check first, but it can help prevent an unneeded render
    if(nextProps.data === undefined){
        this.setState({data: defaultState});   
      return
    }
    if (nextProps.data !== this.state.data) {
      if(nextProps.data.meta === undefined){
        this.setState({data: defaultState});   
      }else{
        this.setState({data: nextProps.data});
      }
    }
  }

 handleChange = (evt) => {
    clearTimeout(this.timer)
    const value = evt.target.type === "checkbox" ? evt.target.checked : evt.target.value;
    var data = {...this.state.data}
    data[evt.target.name] = value
    this.setState({data});
    this.timer = setTimeout( () => this.triggerChange(data), WAIT_INTERVAL)
  }

  triggerChange = (data) => {
    console.log("CHANGFE TRIGGERED")
    this.props.dispatch(photoActions.edit(data))
  }

  
   handleKeyDown = e => {
    if (e.keyCode === ENTER_KEY) {
      clearTimeout(this.timer)
      this.triggerChange(this.state.data)
    }
  }
 
  handleVisablityChange = (value) => {
    var data = {...this.state.data}
    data.meta["visibility"] = value
    this.setState({data});
    this.triggerChange(data)
  }
   handleCollectionChange = value => {
     console.log(value)
    var data = {...this.state.data}
    data["album"] = value
    this.setState({data});
    this.triggerChange(data)
   }

  render() {
    const formItemLayout = {
        labelCol: {
          xs: { span: 24 },
          sm: { span: 8 },
        },
        wrapperCol: {
          xs: { span: 24 },
          sm: { span: 16 },
        },
      };
      let width = 300;
      if (this.state.data.name === undefined) {
        width = 0;
      }


      function formatDate(date){
        let formattedDate = moment(date, "YYYY-MM-DD'T'HH:mm:SSZ").format("DD-MM-YYYY HH:mm A");
        return formattedDate;
      }
  
    formatTree(this.props.collections)
    const collections = Object.values(this.props.collections)
    return (
          <Sider width={width} style={{ overflow: "auto", height: "calc(100vh - 64px)" }}>
            <img src={config.imageUrl+this.state.data.id+"?size=tiny&token="+localStorage.getItem('token')} width="100%" alt="thumbnail" />
            <Form {...formItemLayout} style={{ padding: "0px 5px" }}>
              <Collapse bordered={false} defaultActiveKey={['1', '2', '3']}>
                <Panel header="Properties" key="1">
                  <Form.Item label="id">
                    {this.state.data.id} 
                  </Form.Item>
                  <Form.Item label="Title">
                    <Input value={this.state.data.name}  name="name" onChange={this.handleChange}  onKeyDown={this.handleKeyDown}/>
                  </Form.Item>
                  <Form.Item label="Caption">
                    <Input value={this.state.data.caption} name="caption" onChange={this.handleChange}  onKeyDown={this.handleKeyDown}/>
                  </Form.Item>
                  <Form.Item label="Collection">
                  <TreeSelect
                    value={this.state.data.album}
                    treeData={collections}
                    placeholder="Select Collection"
                    onChange={this.handleCollectionChange}
                  />          
                  </Form.Item>
                </Panel>
                <Panel header="Visability" key="2">
                  <Form.Item label="Access" hasFeedback>
                    <Select placeholder="Please select a country" value={this.state.data.meta.visibility} name="visibility" onChange={this.handleVisablityChange}>
                      <Option value="PUBLIC">PUBLIC</Option>
                      <Option value="HIDDEN">HIDDEN</Option>
                      <Option value="PRIVATE">PRIVATE</Option>
                    </Select>
                  </Form.Item>
                  <Form.Item label="link">
                   <a href={"/photo/"+this.state.data.id}>{this.state.data.id} </a>
                  </Form.Item>
                </Panel>
                <Panel header="History" key="3">
                  <Form.Item label="Captured">
                  {formatDate(this.state.data.exif.DateTaken)}
                  </Form.Item>
                  <Form.Item label="Uploaded">
                  {formatDate(this.state.data.meta.DateAdded)}
                  </Form.Item>
                  <Form.Item label="Modified">
                  {formatDate(this.state.data.meta.DateModified)}
                  </Form.Item>
                </Panel>
              </Collapse>
            </Form>
          </Sider>
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
export default connect(mapToProps)(SideBar);
