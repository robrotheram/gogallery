import React, { useEffect, useState } from "react";
import Gallery from "../components/gallery";
import { connect } from 'react-redux';
import {config} from '../store'
import './album.css'
import { fuzzySearch } from "../components/Search/utils";
import axios from "axios";
import { useHistory } from "react-router-dom";

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
  },[match.params.date,searchTerm,history])

  useEffect(() => {
    console.log(searchTerm);
    if (searchTerm !== "" && searchTerm !== undefined){
      setPhotoList(fuzzySearch(["name", "caption", "album_name"],photoList, searchTerm ))
    }else{
      setPhotoList(downloadedPhotoList)
    }
  },[searchTerm,photoList,downloadedPhotoList])

  return (
    <main>
      
        <div className="album py-5 bg-light" style={{"marginTop":"0px", height:"calc(100vh - 60px)"}}>
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
