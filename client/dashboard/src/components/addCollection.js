import React from 'react';
import { Modal, Input, TreeSelect, Form } from 'antd';

import { connect } from 'react-redux';
import {galleryActions, collectionActions} from '../store/actions'


const AddCollection = (props) => {
  const [form] = Form.useForm();

  const handleCancel = () => {
    props.dispatch(galleryActions.hideAdd())
  };

  const handleCreate = () => {    
    form.validateFields().then(values => {
        console.log('Received values of form: ', values);
        form.resetFields();
        props.dispatch(collectionActions.create(values))
        props.dispatch(galleryActions.hideAdd())
    })
    .catch(err => {
        console.log("error:",err)
    })
  };
  
  return (
    <Modal
      visible={props.addCollectionModalVisable}
      title="Create a new collection"
      okText="Create"
      onCancel={handleCancel}
      onOk={handleCreate}
    >
      <Form form={form} layout="vertical">
        <Form.Item label="Choose collection" hasFeedback name="id" rules={[{ required: true, message: 'Please select the collection to upload photos to!' }]}>
              <TreeSelect
                  treeData={props.collections}
                  placeholder="Select Collection"
                  //onChange={enableUpload}
                />
        </Form.Item>
        <Form.Item label="Collection Name" name="name" rules={[{ required: true, message: 'Please input the name of collection!' }]}>
          <Input />
        </Form.Item>
      </Form>
    </Modal>
  );
}
  
// class AddCollection extends React.Component {
//   handleCancel = () => {
//     this.props.dispatch(galleryActions.hideAdd())
//   };

//   handleCreate = () => {
//     const { form } = this.formRef.props;
//     form.validateFields((err, values) => {
//       if (err) {
//         return;
//       }

//       console.log('Received values of form: ', values);
//       form.resetFields();

//       this.props.dispatch(collectionActions.create(values))
//       this.props.dispatch(galleryActions.hideAdd())
//     });
//   };

//   saveFormRef = formRef => {
//     this.formRef = formRef;
//   };

//   render() {
//     return (
//         <CollectionCreateForm
//           wrappedComponentRef={this.saveFormRef}
//           visible={this.props.addCollectionModalVisable}
//           onCancel={this.handleCancel}
//           onCreate={this.handleCreate}
//           collections ={this.props.collections}
//         />
//     );
//   }
// }

const mapToProps = (state) =>{
  const addCollectionModalVisable = state.GalleryReducer.addCollectionModalVisable;
  const collections = state.CollectionsReducer.collections
  return {addCollectionModalVisable, collections};
}
export default connect(mapToProps)(AddCollection)
