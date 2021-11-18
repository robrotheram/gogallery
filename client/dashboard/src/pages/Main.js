import React, { useEffect, useState } from 'react';
import './Main.css';

import {
  CalendarOutlined,
  CameraOutlined,
  ClockCircleOutlined,
  PictureOutlined,
  PlusOutlined,
  UploadOutlined,
} from '@ant-design/icons';

import { Layout, Menu, Radio, Tree } from 'antd';
import moment from 'moment';
import { useDispatch, useSelector } from 'react-redux';
import { collectionActions, photoActions } from '../store/actions';

/*****************************/
//import actions from "../store/actions";
import SideBar from '../components/sidebar'
import MoveModal from '../components/modal'
import Header from '../components/header'
import AddCollection from '../components/addCollection'
import UploadPhotos from '../components/upload'
import { galleryActions } from '../store/actions/gallery';
import { IDFromTree } from '../store';
import Gallery from '../components/Gallery';

const { Content, Sider, Footer } = Layout;
const { SubMenu } = Menu;
const { DirectoryTree } = Tree;


const Main = () => {

  const dispatch = useDispatch()

  const { photos } = useSelector(state => state.PhotoReducer)
  const { dates, collections } = useSelector(state => state.CollectionsReducer)
  const { imageSize } = useSelector(state => state.GalleryReducer)

  const [selectedElements, setSelectedElements] = useState([])
  const [collapsed, setCollapsed] = useState(true)
  const [selectedPhoto, setSelectedPhoto] = useState({})
  const [filter, setFilter] = useState("")
  const [uploaded_filter, setUploadedFilter] = useState("")

  useEffect(() => {
    dispatch(photoActions.getAll());
    dispatch(collectionActions.getAll());
  }, [dispatch])

  const selectPhoto = (e, photo) => {
    e.stopPropagation();
    setSelectedPhoto(photos.findIndex(x => x.id === photo.id))
    setSelectedElements([photos.findIndex(x => x.id === photo.id)])

    console.log("Clear Selected", selectedPhoto)
  }

  const clearSelection = () => {
    setSelectedPhoto({})
    console.log("Clear Selected", selectedPhoto)
  }

  const onCollapse = collapsed => {
    setCollapsed(collapsed)
  };


  const getStyle = (index) => {
    let elements = selectedElements.map(s => photos[s]);
    if (elements.find(e => e.id === index) !== undefined) {
      // Selected state
      return {
        border: '3px solid #2185d0',
        margin: 0,
        textAlign: "center"
      };
    }
    return {
      textAlign: "center"
    };
  };

  const handleSizeChange = (e) => {
    dispatch(galleryActions.changeImageSize(e.target.value))
  }

  const onTreeSelect = (selectedKeys, info) => {
    console.log('selected', selectedKeys, info); //.node.props.id);
    filterPhotos({
      key: selectedKeys[0]
    })
  };

  const filterPhotos = (item, datesList) => {
    switch (item.key) {
      case "all":
        setFilter("")
        setUploadedFilter("")
        break;
      case "add":
        setFilter("")
        setUploadedFilter("")
        break;
      case "upload":
        setFilter("")
        setUploadedFilter("")
        break;
      case "uploaded":
        setFilter(datesList[0])
        setUploadedFilter("")
        break;
      default:
        if (item.key !== undefined) {
          setFilter(item.key)
          setUploadedFilter("")
        } else {
          let name = IDFromTree(collections, item["0"])
          setFilter(name.id)
          setUploadedFilter("")
        }
        break;
    }
  }
  const lowercasedFilter = filter.toLowerCase();

  const filteredData = photos.filter(item => {
    return search(item, uploaded_filter, lowercasedFilter)
  });

  function formatDate(date) {
    let formattedDate = moment(date, "YYYY-MM-DD").format("DD-MM-YYYY");
    return formattedDate;
  }

  function search(item, uploaded_filter, lowercasedFilter) {
    if (uploaded_filter !== "") {
      return moment(item.meta["DateAdded"], "YYYY-MM-DD'T'HH:mm:SSZ").format("YYYY-MM-DD HH:mm") === uploaded_filter
    } else {
      if (item["name"].toLowerCase().includes(lowercasedFilter)) { return true }
      if (item["album"].toLowerCase() === (lowercasedFilter)) { return true }
      if (item.exif["DateTaken"].toLowerCase().includes(lowercasedFilter)) { return true }
    }
    return false
  }

  let selectMessage = filteredData.length + " photos"
  if (selectedElements.length > 0) {
    selectMessage = selectedElements.length + " out of " + filteredData.length + " selected"
  }

  let datesList = dates.sort((a, b) => {
    var dateA = new Date(a), dateB = new Date(b);
    return dateA - dateB;
  }).reverse();

  return (

    <Layout style={{ minHeight: '100vh' }}>
      <Header search={filterPhotos} />
      <Layout>
        <Sider collapsible collapsed={collapsed} onCollapse={onCollapse} width={300} style={{ overflowY: "auto" }}>
          <Menu theme="dark" mode="inline" selectable={true} defaultSelectedKeys={["all"]} onSelect={(item) => filterPhotos(item, datesList)}>
            <Menu.Item key="all"><PictureOutlined /><span>All Content</span></Menu.Item>
            <Menu.Item key="uploaded"><ClockCircleOutlined /><span>Last Uploaded</span></Menu.Item>
            <SubMenu
              key="calendar"
              title={
                <span>
                  <CalendarOutlined />
                  <span>Date Captured</span>
                </span>
              }
            >
              {datesList.map((el, index) => (<Menu.Item key={el}>{formatDate(el)}</Menu.Item>))}
            </SubMenu>
            <SubMenu
              key="collections"
              title={
                <span>
                  <CameraOutlined />
                  <span>Collections</span>
                </span>
              }
            >
              <li className="menu-tree">
                <DirectoryTree
                  draggable
                  blockNode
                  onSelect={onTreeSelect}
                  treeData={collections}
                />
              </li>
            </SubMenu>
          </Menu>

          <AddCollection />
          <UploadPhotos />
          <Menu theme="dark" mode="inline" selectable={false}>
            <Menu.Item onClick={() => dispatch(galleryActions.showAdd())} key="add" style={{ backgroundColor: "@popover-background", position: "absolute", bottom: 50 }}><PlusOutlined /> <span>Add Collection</span></Menu.Item>
            <Menu.Item onClick={() => dispatch(galleryActions.showUpload())} key="upload" style={{ backgroundColor: "@popover-background", position: "absolute", bottom: 100 }}><UploadOutlined /> <span>Upload</span></Menu.Item>
          </Menu>
        </Sider>
        <Layout>
          <Content
            style={{
              padding: 28,
              margin: 0,
              height: "calc( 100vh - 106px)",
              overflow: "auto"
            }}
            onClick={clearSelection}
          >
            <Gallery
              images={filteredData}
              imageSize={imageSize}
              selectPhoto={selectPhoto}
              getStyle={getStyle}
            />
          </Content>


          <Footer style={{ backgroundColor: "#141414", height: 42, border: "1px solid black", padding: 4, zIndex: 2, borderBottom: "0px", textAlign: "center" }}>
            {selectedElements.length > 0 && <MoveModal selectedPhotos={selectedElements} />}
            <span style={{ lineHeight: "32px" }}> {selectMessage}</span>
            <Radio.Group onChange={handleSizeChange} style={{ float: "right" }} defaultValue={imageSize}>
              <Radio.Button value="1">tiny</Radio.Button>
              <Radio.Button value="4">Small</Radio.Button>
              <Radio.Button value="6">Medium</Radio.Button>
              <Radio.Button value="8">Large</Radio.Button>
            </Radio.Group>
          </Footer>
        </Layout>
        <SideBar photo={photos[selectedPhoto]} />
      </Layout>
    </Layout>
  );

}

export default Main;