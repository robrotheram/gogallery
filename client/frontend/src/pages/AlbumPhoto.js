import React from "react";
import Gallery from "../components/gallery";
import { connect } from 'react-redux';
import {config, searchTree} from '../store'
import placeholder from "../img/placeholder.png"
import './album.css'
import { Link } from "react-router-dom";
import { fuzzySearch } from "../components/Search/utils";


function AlbumList(props){
  let albums = props.albums
  let classSize = "col-md-12"
  if (props.inline){
    classSize = "col-md-3"
  }
  return (
    albums.map((k,i) => (
      <div className={classSize} key={i}>
        <div className="card mb-4 shadow-sm">
            <Link to={albums[i].id}>
            {
              albums[i].profile_image === ""
              ? <img src={placeholder} alt={albums[i].name} width="100%" height="250px" style={{"objectFit": "cover"}}/>
              : <img src={config.imageUrl+albums[i].profile_image} alt={albums[i].name} width="100%" height="250px" style={{"objectFit": "cover"}}/>
            }
            </Link>
            <div className="card-body">
                <p className="card-text text-center">{albums[i].name}</p>
            </div>
        </div>
      </div>
    ))
  )
}

class AlbumPhotoPage extends React.PureComponent {
  render(){
  let {collections, photos} = this.props
  const id = this.props.match.params.id
  const collection = searchTree(collections, id)
  console.log(collection, id)
  let backgroundImage = placeholder
  if (collection.profile_image !== null){
    if (collection.profile_image !== "" ) {
      backgroundImage = config.imageUrl+collection.profile_image
    }
  }
  const photoList = photos.filter(c => c.album === collection.id) || [];
  let albumclass = "col-lg-2"
  let photoclass = "col-lg-10"
  let albums = []
  if (collection.children !== undefined){
    albums = Object.values(collection.children);
  }
  if (albums.length === 0){
    albumclass = "col-0"
    photoclass = "col-12"
  }
  if (photoList.length === 0){
    albumclass = "col-lg-12"
    photoclass = "col-0"
  }


  console.log(albums)

  return (
    <main>
      <section className="jumbotron text-center hero-bg" style={{"zIndex":"-1", "backgroundImage": `url(${backgroundImage})`}} />
        <div className="album py-5 bg-light" style={{"marginTop":"500px"}}>
            <div className="hero-text">
                <h1 className="jumbotron-heading">{collection.name}</h1>
            </div>
           <div>
              <div className="row" style={{"margin":"0px 40px"}}>
              <div className={albumclass}>
                {photoList.length === 0 ? <div className="row"><AlbumList albums={albums} inline={true}/></div> : <AlbumList albums={albums} inline={false}/> }
              </div>
              <div className={photoclass}><Gallery images={photoList}/></div>



            
               </div>
               <hr/>
               


           </div>
        </div>
    </main>
  );
  }
}
const mapToProps = (state) =>{
  const photos = state.PhotosReducer.photos;

  const collections = state.CollectionsReducer.collections;
  console.log(state.CollectionsReducer.collections)


  let loc = state.router.location.pathname.split("/")
  let searchTerm = state.search.search[loc[1]] 
  if (searchTerm !== "" && searchTerm !== undefined){
    return {
      photos: fuzzySearch(["name", "caption", "album_name"],photos, searchTerm ),
      collections
    }
  }

  return {
    photos,
    collections
  };
}
export default connect(mapToProps)(AlbumPhotoPage)
