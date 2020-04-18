import React from 'react';
import {
    Form,
    Input,
    Divider,
    Button,
    Tree,
    Row, 
    Col,
    Select
  } from 'antd';
  import { connect } from 'react-redux';
  import { collectionActions } from '../../store/actions';
  import {formatTree, IDFromTree} from '../../store'
  import {notify} from '../../store/actions';
  const { DirectoryTree } = Tree;
  const { Option } = Select;

  class AlbumSettings extends React.Component {
    state = {
      confirmDirty: false,
      autoCompleteResult: [],
      auth: {username:""},
      albumName: "",
      albumPic: "",
      albumID: ""
    };
 
    handleSubmit = e => {
      e.preventDefault();
      if( this.state.albumID === ""){
        notify("warning", "Please Select album")
        return;
      }
      this.props.dispatch(collectionActions.update({
        id: this.state.albumID,
        name: this.state.albumName,
        profile_image: this.state.albumPic
      }))
    };

    onTreeSelect = (selectedKeys, info) => {
      //console.log('selected',selectedKeys, info);
      let alb = IDFromTree(this.props.collections, selectedKeys["0"])
      this.setState({
        "albumName": alb.name,
        "albumPic": alb.profile_image,
        "albumID": alb.id
      })
      console.log('selected', alb);
    };
  
    handleConfirmBlur = e => {
      const { value } = e.target;
      this.setState({ confirmDirty: this.state.confirmDirty || !!value });
    };
  
    compareToFirstPassword = (rule, value, callback) => {
      const { form } = this.props;
      form.validateFields(['password'], { force: true });
      if (value && value !== form.getFieldValue('password')) {
        callback('Two passwords that you enter is inconsistent!');
      } else {
        callback();
      }
    };

    onChange = (value) => {
      this.setState({"albumPic": value })
    }
  
    updateAlbumName = (evt) => {
      this.setState({
        "albumName": evt.target.value
      })
    }
    validateToNextPassword = (rule, value, callback) => {
      const { form } = this.props;
      if (value && this.state.confirmDirty) {
        form.validateFields(['confirm'], { force: true });
      }
      if (value && value !== form.getFieldValue('confirm')) {
        callback('Two passwords that you enter is inconsistent!');
      } else {
        callback();
      }
    };
  
    handleWebsiteChange = value => {
      let autoCompleteResult;
      if (!value) {
        autoCompleteResult = [];
      } else {
        autoCompleteResult = ['.com', '.org', '.net'].map(domain => `${value}${domain}`);
      }
      this.setState({ autoCompleteResult });
    };
  
    render() {
      const formItemLayout = {
        labelCol: {
          xs: { span: 24 },
          sm: { span: 5 },
        },
        wrapperCol: {
          xs: { span: 24 },
          sm: { span: 19 },
        },
      };
      const tailFormItemLayout = {
        wrapperCol: {
          xs: {
            span: 24,
            offset: 0,
          },
          sm: {
            span: 8,
            offset: 8,
          },
        },
      };
      
      formatTree(this.props.collections)
      const collections = Object.values(this.props.collections)
      return (
        <Row>
          <Col span={8}>
            <DirectoryTree
              className="draggable-tree"
              defaultExpandedKeys={this.state.expandedKeys}
              blockNode
              onSelect={this.onTreeSelect}
              treeData={collections}
            />
          </Col>
          <Col span={16}>
            <Form {...formItemLayout} onSubmit={this.handleSubmit}>
              < Divider/>
              <Form.Item label="Album Name">
                <Input value={this.state.albumName}  name="albumName" onChange={this.updateAlbumName}/>
              </Form.Item>
              <Form.Item label="Album Image id">
                <Select
                  showSearch
                  placeholder="Select a photo"
                  optionFilterProp="children"
                  onChange={this.onChange}
                  value={this.state.albumPic} 
                  filterOption={(input, option) =>
                    option.props.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
                  }
                >
                {this.props.photos.map((el, index) => (<Option key={el.id}>{el.name}</Option> ))}
                </Select>
              </Form.Item>

              

              <Divider/>
              <Form.Item {...tailFormItemLayout}>
                <Button type="primary" htmlType="submit" style={{width:"100%"}}>
                  Update
                </Button>
              </Form.Item>
            </Form>
          </Col>
        </Row>
      );
    }
  }
const mapToProps = (state) =>{
  console.log("REG",state.UserReducer);
  const auth = state.UserReducer;
  const photos = state.PhotoReducer.photos;
  const collections = state.CollectionsReducer.collections
  return {
    auth,
    collections,
    photos
  };
}

export default connect(mapToProps)(Form.create({ name: 'register', mapPropsToFields(props) {
  return {
    username: Form.createFormField({...props.username, value: props.auth.username}),
    email: Form.createFormField({...props.email, value: props.auth.email}),
  };
}})(AlbumSettings));
