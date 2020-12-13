import React from 'react';
import './Main.css';
import { Layout, Statistic, Card, Icon, Row, Col}  from 'antd';
import { connect } from 'react-redux';
import Header from '../components/header'
import RegistrationForm from '../components/settings/RegistrationForm'
import ProfileForm from '../components/settings/ProfileForm'
import SettingsForm from '../components/settings/SettingsForm'
import Maintenance from '../components/settings/maintenance'
import AlbumSettings from '../components/settings/AlbumSettings'
import { settingsActions } from '../store/actions/settings';

const { Content } = Layout;
const tabListNoTitle = [
  {
    key: 'user',
    tab: <span><Icon style={{"margin-right":"5px"}}  type="user" /> User Settings</span>,
  },
  {
    key: 'profile',
    tab: <span><Icon style={{"margin-right":"5px"}}  type="profile" /> Profile Settings</span>,
  },
  {
    key: 'settings',
    tab: <span><Icon style={{"margin-right":"5px"}}  type="setting" />Site Settings</span>,
  },
  {
    key: 'album',
    tab: <span><Icon style={{"margin-right":"5px"}}  type="folder-open" />Album Settings</span>,
  },
  {
    key: 'maintenance',
    tab: <span><Icon style={{"margin-right":"5px"}}  type="tool" /> Maintenance</span>,
  },
];

const contentListNoTitle = {
  user: <RegistrationForm/>,
  profile: <ProfileForm/>,
  settings: <SettingsForm/>,
  album: <AlbumSettings/>,
  maintenance: <Maintenance/>
};

class Settings extends React.PureComponent {
  constructor() {
    super();
    this.state = {
      key: 'tab1',
      noTitleKey: 'settings',
    };
  }

  componentDidMount() {
    this.props.dispatch(settingsActions.all());
  }
  
  onTabChange = (key, type) => {
    console.log(key, type);
    this.setState({ [type]: key });
  };

  render() {
    console.log("STATS", this.props.stats)
    return (
      <Layout style={{ minHeight: '100vh' }}>
        <Header/>
        <Layout>
          <Content style={{ padding: '50px' }}>
            <Row gutter={16}>
              <Col span={8}>
                <Card>
                  <Statistic
                    value={this.props.stats.Photos}
                    precision={2}
                    valueStyle={{ textAlign: "center" }}
                    prefix={<Icon style={{"marginRight":"5px"}} type="picture" />}
                  />
                </Card>
              </Col>
              <Col span={8}>
                <Card>
                  <Statistic
                    value={this.props.stats.Albums}
                    precision={2}
                    valueStyle={{ textAlign: "center" }}
                    prefix={<Icon style={{"":"5px"}}  type="folder-open" />}
                    />
                </Card>
              </Col>
              <Col span={8}>
                <Card>
                  <Statistic
                    value={this.props.stats.Rubish}
                    precision={2}
                    valueStyle={{ textAlign: "center" }}
                    prefix={<Icon style={{"marginRight":"5px"}}  type="delete" />}
                  />
                </Card>
              </Col>
              {/* <Col span={4}>
                <Card>
                  <Statistic
                    value={this.props.stats.ProcessQue}
                    precision={2}
                    valueStyle={{ textAlign: "center" }}
                    prefix={<Icon style={{"margin-right":"5px"}}  type="code" />}
                  />
                </Card>
              </Col>
              <Col span={4}>
                <Card>
                  <Statistic
                    value={this.props.stats.ViewCount}
                    precision={2}
                    valueStyle={{ textAlign: "center" }}
                    prefix={<Icon style={{"margin-right":"5px"}}  type="eye" />}
                  />
                </Card>
              </Col>
              <Col span={4}>
                <Card>
                  <Statistic
                    value={this.props.stats.Albums}
                    precision={2}
                    valueStyle={{ textAlign: "center" }}
                    prefix={<Icon style={{"margin-right":"5px"}} type="delete" />}
                  />
                </Card>
              </Col>
              <Col span={4}>
                <Card>
                  <Statistic
                    value={this.props.stats.Albums}
                    precision={2}
                    valueStyle={{ textAlign: "center" }}
                    prefix={<Icon style={{"margin-right":"5px"}} type="delete" />}
                  />
                </Card>
              </Col> */}
            </Row>
            <Card
              style={{ width: '100%', marginTop:"20px"}}
              tabList={tabListNoTitle}
              activeTabKey={this.state.noTitleKey}
              bodyStyle={{backgroundColor:"#000"}}
              onTabChange={key => {
                this.onTabChange(key, 'noTitleKey');
              }}
            >
            {contentListNoTitle[this.state.noTitleKey]}
            </Card>
          </Content>
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
  const stats = state.SettingsReducer.stats
  return {
    stats,
    photos,
    dates,
    collections,
    uploadDates
  };
}

export default connect(mapToProps)(Settings);