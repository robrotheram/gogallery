import React from 'react';
import Selection from 'react-ds';
import './Main.css';
import { Layout, Menu, Icon, Radio, Row, Col, Tree}  from 'antd';
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
import { config, formatTree, IDFromTree } from '../store';
import { LazyImage } from '../components/Lazyloading';

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
    console.log('selected', info); //.node.props.id);
  };

  filterPhotos = (item) => {
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
        this.setState({ filter: "", uploaded_filter: this.props.uploadDates.sort().reverse()[0] })
        break;
      default:
        if (item.key !== undefined){
          this.setState({ filter: item.key, uploaded_filter: "" })
        }else{
          let name = IDFromTree(this.props.collections, item["0"])
          this.setState({ filter: name, uploaded_filter: "" })
          
        }
        break;
    }
  }

  render() {
    const selectedElements = this.state.selectedElements.map(s => this.props.photos[s]);
    const lowercasedFilter = this.state.filter.toLowerCase();

    formatTree(this.props.collections)
    const collections = Object.values(this.props.collections)


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

    return (
      <Layout style={{ minHeight: '100vh' }}>
        <Header search={this.filterPhotos}/>
        <Layout>
          <Sider collapsible collapsed={this.state.collapsed} onCollapse={this.onCollapse} style={{ overflowY: "auto" }}>
            <Menu theme="dark" mode="inline" selectable={true} defaultSelectedKeys={["all"]} onSelect={this.filterPhotos}>
              <Menu.Item key="all">
                <Icon type="picture" />
                <span>All Content</span>
              </Menu.Item>
              <Menu.Item key="uploaded">
                <Icon type="clock-circle" />
                <span>Last Uploaded</span>
              </Menu.Item>
              <SubMenu
                key="calendar"
                title={
                  <span>
                    <Icon type="calendar" />
                    <span>Date Captured</span>
                  </span>
                }
              >
                {this.props.dates.map((el, index) => (<Menu.Item key={el}>{formatDate(el)}</Menu.Item>))}
              </SubMenu>
              <SubMenu
                key="collections"
                title={
                  <span>
                    <Icon type="camera" />
                    <span>Collections</span>
                  </span>
                }
              >
                <DirectoryTree
                className="draggable-tree"
                defaultExpandedKeys={this.state.expandedKeys}
                draggable
                blockNode
                onSelect={this.onTreeSelect()}
                treeData={collections}
              />
              </SubMenu>
            </Menu>
            <AddCollection />
            <UploadPhotos /> 
            <Menu theme="dark" mode="inline" selectable={false}>
              <Menu.Item onClick={() => this.props.dispatch(galleryActions.showAdd())} key="add" style={{ backgroundColor: "@popover-background", position: "absolute", bottom: 50 }}><Icon type="plus" /> <span>Add Collection</span></Menu.Item>
              <Menu.Item onClick={() => this.props.dispatch(galleryActions.showUpload())} key="upload" style={{ backgroundColor: "@popover-background", position: "absolute", bottom: 100 }}><Icon type="upload" /> <span>Upload</span></Menu.Item>
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
                <Row gutter={[16, 16]}>
                  {filteredData.map((el, index) => (
                    <Col key={el.id} span={parseInt(this.props.imageSize)} style={{ padding: 2 }}>
                      <div
                        ref={this.addElementRef}
                        className={`item`}
                      >
                        <figure className="galleryImg" style={this.getStyle(el.id)} onClick={(e) => this.selectPhoto(e, el)}>
                          <LazyImage src={config.imageUrl + el.id+"?size=tiny&token="+localStorage.getItem('token')} width="100%" height="100%" alt="thumbnail" />
                        </figure>
                      </div>
                    </Col>
                  ))}
                  {this.renderSelection()}
                </Row>
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