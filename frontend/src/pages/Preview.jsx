import React, { useEffect, useState } from 'react';
import './Main.css';

import { Layout } from 'antd';
import { useDispatch } from 'react-redux';
import Header from '../components/header'
const { Content } = Layout;

const Preview = () => {
  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header />
      <Layout>
        <Content style={{ padding: '20px', height: 'calc( 100vh - 60px)', overflow:"auto"}}>
         <iframe src='http://localhost:8800/preview-build' width={"100%"} height={"100%"}/>
        </Content>
      </Layout>
    </Layout>
  );

}
export default Preview;