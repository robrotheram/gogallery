import React, { useCallback, useEffect, useState } from 'react';
import { AutoComplete, Form } from 'antd';
// as an array
import { Layout, Input, Select, TreeSelect, Typography } from 'antd';
import { Collapse } from 'antd';

import { useDispatch, useSelector } from 'react-redux';
import { config } from '../store';
import { getOptions, photoActions } from '../store/actions';
import { LocationModal } from './Map'
import axios from 'axios';

const { Paragraph } = Typography;
const { Panel } = Collapse;
const { Sider } = Layout;
const { Option } = Select;

const SideBar = ({ photo }) => {

  const { collections } = useSelector(state => state.CollectionsReducer);
  const dispatch = useDispatch()
  const [form] = Form.useForm();
  const [data, setData] = useState({})
  const [captionOptions, setCaptionOptions] = useState([])
  
  const editPhoto = () => {
    let formData = form.getFieldValue()
    let newPhoto = { ...data, ...formData }
    if (JSON.stringify(newPhoto) !== JSON.stringify(data)) {
      dispatch(photoActions.edit(newPhoto))
    }

  }

  const getCaptionList = useCallback(() => {
    axios.get(`${config.baseUrl}/photo/${photo.id}/caption`, getOptions()).then(res => {
      console.log("CAPTIONS", res)
      if (res.data.status === "ok") {
         let options = res.data.predictions.map( prediction => { return {value: prediction.caption}})
         setCaptionOptions(options)
      }
    }).catch(err => console.log(err))
  }, [photo]); 

  useEffect(() => {
    if (photo) {
      console.log("Photo Update")
      setData(photo)
      //getCaptionList()
      form.setFieldsValue(photo)
    } else {
      setData({})
    }
  }, [photo, form, getCaptionList])

  const updateGPS = (lat, lng) => {
    data.exif.GPS = {
      latitude: lat,
      longitude: lng
    }
    console.log("UPDATE GPS", data)
    dispatch(photoActions.edit(data))
  }

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

  const formatDate = (datestr) => {
    console.log("DATE:",datestr)
    const date = new Date(Date.parse(datestr));
    return date.toLocaleString();
  }

  return (
    <Sider width={data.name ? 500 : 0} style={{ overflow: "auto", height: "calc(100vh - 64px)" }}>
      <img src={config.imageUrl + data.id + "?size=small"} width="100%" alt="thumbnail" />
      <Form
        form={form}
        layout="horizontal"
        {...formItemLayout}>

        <Collapse bordered={false} defaultActiveKey={['1']}>
          <Panel header="Properties" key="1">
            <Form.Item label="id" name="id">
              <Paragraph ellipsis style={{ marginBottom: "0px" }}>
                {data.id}
              </Paragraph>
            </Form.Item>
            <Form.Item label="Photo Name" name="name">
              <Input onBlur={editPhoto} />
            </Form.Item>
            <Form.Item label="Caption" name="caption">
              <Input onBlur={editPhoto} />
            </Form.Item>
            <Form.Item label="Path" name="path">
              <Input onBlur={editPhoto} />
            </Form.Item>
            <Form.Item label="Location">
              <LocationModal lat={data.exif ? data.exif.gps.latitude : 0} lng={data.exif ? data.exif.gps.longitude : 0} onUpdate={updateGPS} />
            </Form.Item>

            <Form.Item label="Collection" name="album">
              <TreeSelect
                treeData={collections}
                placeholder="Select Collection"
                onChange={editPhoto}
              />
            </Form.Item>
          </Panel>
          <Panel header="Visability" key="2">
            <Form.Item label="Access" name={["meta", "visibility"]} hasFeedback onChange={editPhoto}>
              <Select placeholder="Please select a country">
                <Option value="PUBLIC">PUBLIC</Option>
                <Option value="HIDDEN">HIDDEN</Option>
                <Option value="PRIVATE">PRIVATE</Option>
              </Select>
            </Form.Item>
            <Form.Item label="link">
              <a href={"/photo/" + data.id}>{data.name} </a>
            </Form.Item>
          </Panel>
          <Panel header="History" key="3">
            <Form.Item label="Captured" style={{marginBottom:"10px"}}>
              {data.exif ? formatDate(data.exif.date_taken) : ""}
            </Form.Item>
            <Form.Item label="Uploaded" style={{marginBottom:"10px"}}>
              {data.meta ? formatDate(data.meta.date_added) : ""}
            </Form.Item>
            <Form.Item label="Modified" style={{marginBottom:"10px"}}>
              {data.meta ? formatDate(data.meta.date_modified) : ""}
            </Form.Item>
          </Panel>
        </Collapse>
      </Form>
    </Sider>
  );
}
export default SideBar;
