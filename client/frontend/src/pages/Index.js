import React from "react";
import Gallery from "../components/gallery";
import { connect } from 'react-redux';


class IndexPage  extends React.PureComponent {
  render() {
  return (
    <div className="App">
      <div style={{"width":"100%", marginTop:"100px" }}>
      <Gallery images={this.props.photos}/>
      </div>
    </div>
  );
  }
}
const mapToProps = (state) =>{
  const photos = state.PhotosReducer.photos;
  return {
    photos,
  };
}
export default connect(mapToProps)(IndexPage)



