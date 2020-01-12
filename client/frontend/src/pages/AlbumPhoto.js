import React from "react";
import Gallery from "../components/gallery";
import { connect } from 'react-redux';
import {config} from '../store'

import './album.css'

class AlbumPhotoPage extends React.PureComponent {

  render(){
  let {collections, photos} = this.props
  const id = this.props.match.params.id
  const collection = collections.filter(c => c.id === id)[0] || {};
  let backgroundImage = config.imageUrl+collection.profile_image
  const photoList = photos.filter(c => c.album === collection.name) || [];
  return (
    <main>
      <section className="jumbotron text-center hero-bg" style={{"zIndex":"-1", "backgroundImage": `url(${backgroundImage})`}} />
        <div className="album py-5 bg-light" style={{"marginTop":"500px"}}>
            <div className="hero-text">
                <h1 className="jumbotron-heading">{collection.name}</h1>
            </div>
           <div>
              <Gallery images={photoList}/>
           </div>
        </div>
    </main>
  );
  }
}
const mapToProps = (state) =>{
  const photos = state.PhotosReducer.photos;
  const collections = state.CollectionsReducer.collections;
  return {
    photos,
    collections
  };
}
export default connect(mapToProps)(AlbumPhotoPage)
