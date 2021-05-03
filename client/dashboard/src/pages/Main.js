import React from 'react';
import Selection from 'react-ds';
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
import { connect } from 'react-redux';
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
import Gallery  from '../components/Gallery';

const { Content, Sider, Footer } = Layout;
const { SubMenu } = Menu;
const { DirectoryTree } = Tree;


class Main extends React.PureComponent {

  constructor() {
    super();

    this.state = {
      ref: null,
      elRefs: [],
      selectedElements: [], // track the elements that are selected
      collapsed: true,
      selectedPhoto: {},
      filter: "",
      uploaded_filter: ""
    };
  }

  _handleKeyDown = (event) => {
    switch (event.keyCode) {
      case 27:
        this.setState({
          selectedPhoto: {},
          selectedElements: []
        })
        break;
      default:
        break;
    }
  }

  componentDidMount() {
   this.props.dispatch(photoActions.getAll());
   this.props.dispatch(collectionActions.getAll());
    // this.props.getAllPhotos();
    // this.props.getCollections();

    document.addEventListener("keydown", this._handleKeyDown);
  }
  componentWillUnmount() {
    document.removeEventListener("keydown", this._handleKeyDown);
  }

  selectPhoto = (e, photo) => {
    e.stopPropagation();
    this.setState({
      selectedPhoto: this.props.photos.findIndex(x => x.id === photo.id),
      selectedElements: [this.props.photos.findIndex(x => x.id === photo.id)]
    })
  }

  clearSelection = () => {
    console.log("Clear Selected")
    this.setState({ selectedPhoto: {} });
  }

  onCollapse = collapsed => {
    console.log(collapsed);
    this.setState({ collapsed });
  };

  handleSelection = (indexes) => {
    this.setState({
      selectedElements: indexes,
    });
  };

  getStyle = (index) => {
    let selectedElements = this.state.selectedElements.map(s => this.props.photos[s]);
    if (selectedElements.find(e => e.id === index) !== undefined) {
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

  handleSizeChange = (e) => {
    this.props.dispatch(galleryActions.changeImageSize(e.target.value))
  }

  addElementRef = (ref) => {
    const elRefs = this.state.elRefs;
    elRefs.push(ref);
    this.setState({
      elRefs,
    });
  };

  renderSelection() {
    if (!this.state.ref || !this.state.elRefs) {
      return null;
    }
    return (
      <Selection
        target={this.state.ref}
        elements={this.state.elRefs}
        onSelectionChange={this.handleSelection}
        style={this.props.style}
      />
    );
  }

  search = (event) => {
    this.setState({ filter: event.target.value })
  }

  onTreeSelect = (selectedKeys, info) => {
    console.log('selected', selectedKeys, info); //.node.props.id);
    this.filterPhotos({
      key: selectedKeys[0]
    })
  };

  filterPhotos = (item, datesList) => {
    switch (item.key) {
      case "all":
        this.setState({ filter: "", uploaded_filter: "" })
        break;
      case "add":
        this.setState({ filter: "", uploaded_filter: "" })
        break;
      case "upload":
        this.setState({ filter: "", uploaded_filter: "" })
        break;
      case "uploaded":
        this.setState({ filter: datesList[0], uploaded_filter: "" })
        break;
      default:
        if (item.key !== undefined){
          this.setState({ filter: item.key, uploaded_filter: "" })
        }else{
          let name = IDFromTree(this.props.collections, item["0"])
          this.setState({ filter: name.id, uploaded_filter: "" })
          
        }
        break;
    }
  }

  render() {
    const selectedElements = this.state.selectedElements.map(s => this.props.photos[s]);
    const lowercasedFilter = this.state.filter.toLowerCase();

    const filteredData = this.props.photos.filter(item => {
      return search(item, this.state.uploaded_filter, lowercasedFilter)
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
    if (this.state.selectedElements.length > 0) {
      selectMessage = this.state.selectedElements.length + " out of " + filteredData.length + " selected"
    }

    let datesList = this.props.dates.sort((a, b) => {
        var dateA = new Date(a), dateB = new Date(b);
        return dateA - dateB;
    }).reverse();

    return (
     
      <Layout style={{ minHeight: '100vh' }}>
        <Header search={this.filterPhotos}/>
        <Layout>
          <Sider collapsible collapsed={this.state.collapsed} onCollapse={this.onCollapse} width={300} style={{ overflowY: "auto" }}>

            <Menu theme="dark" mode="inline" selectable={true} defaultSelectedKeys={["all"]} onSelect={(item) => this.filterPhotos(item, datesList)}>
              <Menu.Item key="all">
                <PictureOutlined />
                <span>All Content</span>
              </Menu.Item>
              <Menu.Item key="uploaded">
                <ClockCircleOutlined />
                <span>Last Uploaded</span>
              </Menu.Item>
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
                <div className="menu-tree">
                  <DirectoryTree
                    defaultExpandedKeys={this.state.expandedKeys}
                    draggable
                    blockNode
                    onSelect={this.onTreeSelect}
                    treeData={this.props.collections}
                  />
                </div>
              </SubMenu>
            </Menu>
            <AddCollection />
            <UploadPhotos /> 
            <Menu theme="dark" mode="inline" selectable={false}>
              <Menu.Item onClick={() => this.props.dispatch(galleryActions.showAdd())} key="add" style={{ backgroundColor: "@popover-background", position: "absolute", bottom: 50 }}><PlusOutlined /> <span>Add Collection</span></Menu.Item>
              <Menu.Item onClick={() => this.props.dispatch(galleryActions.showUpload())} key="upload" style={{ backgroundColor: "@popover-background", position: "absolute", bottom: 100 }}><UploadOutlined /> <span>Upload</span></Menu.Item>
            </Menu>
          </Sider>
          <Layout>
              <div ref={(ref) => { this.setState({ ref }); }} className='item-container' onClick={this.clearSelection}>
              <Content
                style={{
                  padding: 28,
                  margin: 0,
                  height: "calc( 100vh - 106px)",
                  overflow: "auto"
                }}
              >
              
                <Gallery 
                  images={filteredData} 
                  imageSize={this.props.imageSize} 
                  addElementRef={this.addElementRef} 
                  selectPhoto={this.selectPhoto}
                  getStyle={this.getStyle}
                />
                {this.renderSelection()}
              </Content>
              </div>
           
            <Footer style={{ backgroundColor: "#141414", height: 42, border: "1px solid black", padding: 4, zIndex: 2, borderBottom: "0px", textAlign: "center" }}>
              {this.state.selectedElements.length > 0 && <MoveModal selectedPhotos={selectedElements} />}
              <span style={{ lineHeight: "32px" }}> {selectMessage}</span>
              <Radio.Group onChange={this.handleSizeChange} style={{ float: "right" }} defaultValue={this.props.imageSize}>
                <Radio.Button value="1">tiny</Radio.Button>
                <Radio.Button value="4">Small</Radio.Button>
                <Radio.Button value="6">Medium</Radio.Button>
                <Radio.Button value="8">Large</Radio.Button>
              </Radio.Group>
            </Footer>
          </Layout>
          <SideBar data={this.props.photos[this.state.selectedPhoto]} /> 
        </Layout>
      </Layout>
    );
  }
}



const mapToProps = (state) =>{
  const photos = state.PhotoReducer.photos;
  const dates = state.CollectionsReducer.dates
  const uploadDates = state.CollectionsReducer.uploadDates
  const collections = state.CollectionsReducer.collections
  const imageSize = state.GalleryReducer.imageSize
  return {
    photos,
    dates,
    collections,
    uploadDates,
    imageSize

  };
}

export default connect(mapToProps)(Main);