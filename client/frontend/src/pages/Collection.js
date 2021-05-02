import React, { useEffect, useState } from "react";
import Gallery from "../components/gallery";
import { connect } from 'react-redux';
import {config, searchTree} from '../store'
import placeholder from "../img/placeholder.png"
import './album.css'
import { Link } from "react-router-dom";
import { fuzzySearch } from "../components/Search/utils";
import axios from "axios";
import { useHistory } from "react-router-dom";


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

const CollectionPage = ({match, searchTerm}) => {
  console.log("CollectionPage", match.params.date)
  const history = useHistory();
  const [downloadedPhotoList, setDownloadedPhotoList] = useState([])
  const [photoList, setPhotoList] = useState([])
  const photoclass = "col-12"

  useEffect(() =>{
    let date = match.params.date
    axios.get(config.baseUrl+"/photos/"+date).then(res => {
      if (date === "latest"){
        history.push("/collection/"+res.data)
        return
      }
      setDownloadedPhotoList(res.data)
      if (searchTerm !== "" && searchTerm !== undefined){
          setPhotoList(fuzzySearch(["name", "caption", "album_name"],res.data, searchTerm ))
          return;
      }else{
        setPhotoList(res.data)  
      }
      if (res.data.length === 0){
        history.push("/")
      }
    }).catch(() =>{
      history.push("/")
    })
  },[match.params.date])

  useEffect(() => {
    console.log(searchTerm);
    if (searchTerm !== "" && searchTerm !== undefined){
      setPhotoList(fuzzySearch(["name", "caption", "album_name"],photoList, searchTerm ))
    }else{
      setPhotoList(downloadedPhotoList)
    }
  },[searchTerm])

  return (
    <main>
      
        <div className="album py-5 bg-light" style={{"marginTop":"0px"}}>
           <div>
              <div className="row" style={{"margin":"0px 40px"}}>
              <div className={photoclass}><Gallery images={photoList}/></div>
               </div>
               <hr/>
           </div>
        </div>
    </main>
  );
}

const mapToProps = (state) =>{
  let loc = state.router.location.pathname.split("/")
  let searchTerm = state.search.search[loc[1]] 
  return {
    searchTerm: searchTerm
  }
}
export default connect(mapToProps)(CollectionPage)
