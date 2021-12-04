import React, { useCallback, useEffect, useState } from 'react';
import { AutoComplete, Form } from 'antd';
// as an array
import { Layout, Input, Select, TreeSelect, Typography } from 'antd';
import { Collapse } from 'antd';
import moment from 'moment';

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
      getCaptionList()
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
    setData(data)
    editPhoto();
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

  const formatDate = (date) => {
    let formattedDate = moment(date, "YYYY-MM-DD'T'HH:mm:SSZ").format("DD-MM-YYYY HH:mm A");
    return formattedDate;
  }

  

  return (
    <Sider width={data.name ? 500 : 0} style={{ overflow: "auto", height: "calc(100vh - 64px)" }}>
      <img src={config.imageUrl + data.id + "?size=tiny&token=" + localStorage.getItem('token')} width="100%" alt="thumbnail" />
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
              <AutoComplete options={captionOptions}>
                <Input onBlur={editPhoto} />
              </AutoComplete>
            </Form.Item>
            <Form.Item label="Path" name="path">
              <Input onBlur={editPhoto} />
            </Form.Item>
            <Form.Item label="Location">
              <LocationModal lat={data.exif ? data.exif.GPS.latitude : 0} lng={data.exif ? data.exif.GPS.longitude : 0} onUpdate={updateGPS} />
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
            <Form.Item label="Captured">
              {data.exif ? formatDate(data.exif.DateTaken) : ""}
            </Form.Item>
            <Form.Item label="Uploaded">
              {data.meta ? formatDate(data.meta.DateAdded) : ""}
            </Form.Item>
            <Form.Item label="Modified">
              {data.meta ? formatDate(data.meta.DateModified) : ""}
            </Form.Item>
          </Panel>
        </Collapse>
      </Form>
    </Sider>
  );
}
export default SideBar;
