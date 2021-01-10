import React from "react";
import Gallery from "../components/gallery";
import { connect } from 'react-redux';
import { fuzzySearch } from "../components/Search/utils";


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

  let loc = state.router.location.pathname.split("/")
  let searchTerm = state.search.search[loc[1]] 
  if (searchTerm !== "" && searchTerm !== undefined){
    return {
      photos: fuzzySearch(["name", "caption","album_name"],photos, searchTerm )
    }
  }

  return {
    photos,
  };
}
export default connect(mapToProps)(IndexPage)



