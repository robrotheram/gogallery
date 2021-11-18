import React, { useEffect, useState } from 'react';
import './Main.css';

import {
  DeleteOutlined,
  FolderOpenOutlined,
  PictureOutlined,
  ProfileOutlined,
  SettingOutlined,
  ToolOutlined,
  UserOutlined,
} from '@ant-design/icons';

import { Layout, Statistic, Card, Row, Col } from 'antd';
import { useDispatch, useSelector } from 'react-redux';
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
    tab: <span><UserOutlined style={{ "marginRight": "5px" }} /> User Settings</span>,
  },
  {
    key: 'profile',
    tab: <span><ProfileOutlined style={{ "marginRight": "5px" }} /> Profile Settings</span>,
  },
  {
    key: 'settings',
    tab: <span><SettingOutlined style={{ "marginRight": "5px" }} />Site Settings</span>,
  },
  {
    key: 'album',
    tab: <span><FolderOpenOutlined style={{ "marginRight": "5px" }} />Album Settings</span>,
  },
  {
    key: 'maintenance',
    tab: <span><ToolOutlined style={{ "marginRight": "5px" }} /> Maintenance</span>,
  },
];

const contentListNoTitle = {
  user: <RegistrationForm />,
  profile: <ProfileForm />,
  settings: <SettingsForm />,
  album: <AlbumSettings />,
  maintenance: <Maintenance />
};

const Settings = () => {
  const stats = useSelector(state => state.SettingsReducer.stats);
  const [tab, setTab] = useState({ key: 'tab1', noTitleKey: 'settings' });
  const dispatch = useDispatch()

  useEffect(() => {
    dispatch(settingsActions.all());
  }, [dispatch])

  const onTabChange = (key, type) => {
    console.log(key, type);
    setTab({ [type]: key });
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header />
      <Layout>
        <Content style={{ padding: '50px' }}>
          <Row gutter={16}>
            <Col span={8}>
              <Card>
                <Statistic
                  value={stats.Photos}
                  precision={2}
                  valueStyle={{ textAlign: "center" }}
                  prefix={<PictureOutlined style={{ "marginRight": "5px" }} />}
                />
              </Card>
            </Col>
            <Col span={8}>
              <Card>
                <Statistic
                  value={stats.Albums}
                  precision={2}
                  valueStyle={{ textAlign: "center" }}
                  prefix={<FolderOpenOutlined style={{ "": "5px" }} />}
                />
              </Card>
            </Col>
            <Col span={8}>
              <Card>
                <Statistic
                  value={stats.Rubish}
                  precision={2}
                  valueStyle={{ textAlign: "center" }}
                  prefix={<DeleteOutlined style={{ "marginRight": "5px" }} />}
                />
              </Card>
            </Col>
          </Row>
          <Card
            style={{ width: '100%', marginTop: "20px" }}
            tabList={tabListNoTitle}
            activeTabKey={tab.noTitleKey}
            bodyStyle={{ backgroundColor: "#000" }}
            onTabChange={key => {
              onTabChange(key, 'noTitleKey');
            }}
          >
            {contentListNoTitle[tab.noTitleKey]}
          </Card>
        </Content>
      </Layout>
    </Layout>
  );

}
export default Settings;