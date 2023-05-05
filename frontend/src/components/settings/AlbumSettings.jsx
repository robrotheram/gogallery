import React, {useState, useEffect} from 'react';
import { Form } from "antd"
import { Input, Divider, Button, Tree, Row, Col, Select } from 'antd';
import { useDispatch, useSelector } from 'react-redux';
import { collectionActions } from '../../store/actions';
import {notify} from '../../store/actions';
import {LocationModal} from '../Map'

const { Option } = Select;
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

const AlbumSettings = () => {

  const photos = useSelector(state =>state.PhotoReducer.photos);
  const collections = useSelector(state =>state.CollectionsReducer.collections);
  const dispatch = useDispatch();
  
  const [albumName, setAlbumName] = useState("")
  const [albumPic, setAlbumPic] = useState("")
  const [albumID, setAlbumID] = useState("")
  
  const [GPS, setGPS] = useState({
    latitude:0,
    longitude:0
  })
  
  const [form] = Form.useForm();
  useEffect(() => {
    form.setFieldsValue({
      albumName: albumName,
      albumPic: albumPic,
      albumID: albumID,
    });
  }, [form, albumName, albumPic,albumID]);





  const handleSubmit = () => {
    if(albumID === ""){
      notify("warning", "Please Select album")
      return;
    }
    dispatch(collectionActions.update({
      id: albumID,
      name: albumName,
      profile_image: albumPic,
      GPS: GPS
    }))
  };

  const onTreeSelect = (selectedKeys, info) => {
    let alb = findInTree(collections, selectedKeys[0])
    console.log("TREE_SELECT", alb, selectedKeys)
    if(alb === undefined){
      return
    }
    setAlbumID(alb.id)
    setAlbumPic(alb.profile_image)
    setAlbumName(alb.name)
    setGPS(alb.gps)
   
  };

  const onChange = (value) => {
    setAlbumPic(value)
  }

  const updateAlbumName = (evt) => {
    setAlbumName(evt.target.value)
  }

  const updateGPS = (lat, lng) => {
    setGPS(
      {
        latitude:lat,
        longitude:lng
      }
    )
  }
  
  
  const findInTree = (tree, id) => {
    let el
    const proceesNode = (node) => {
      if (node.id === id) {
        el = node
        return
      }
      return node.children.map(n => proceesNode(n))
    }
    tree = Object.values(tree)
    tree.map(node => proceesNode(node))
    return el
  }


  console.log("COLLECTIONS:", collections)
   return (
      <Row>
        <Col span={8} style={{"overflowY": "auto","maxHeight": "500px"}} >
          <Tree
            className="draggable-tree"
            defaultExpandedKeys={[]}
            blockNode
            onSelect={onTreeSelect}
            treeData={collections}
          />
        </Col>
        <Col span={16} style={{"paddingLeft":"30px"}}>
          <Form form={form} {...formItemLayout} onFinish={handleSubmit}>
            < Divider/>
            <Form.Item label="Album Name" name="albumName">
              <Input onChange={updateAlbumName}/>
            </Form.Item>
            <Form.Item label="Album Image id" name="albumPic">
              <Select
                showSearch
                placeholder="Select a photo"
                optionFilterProp="children"
                onChange={onChange}
                filterOption={(input, option) =>
                  option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
                }
              >
              {photos.map((el, index) => (<Option key={el.id}>{el.name}</Option> ))}
              </Select>
            </Form.Item>
            <Form.Item label="Location">
                  <LocationModal lat={GPS.latitude} lng={GPS.longitude} onUpdate={updateGPS}/>
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

export default AlbumSettings;
