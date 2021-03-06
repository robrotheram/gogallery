import React, {useEffect, useState} from "react";
import { connect } from 'react-redux';
import { Link } from "react-router-dom";
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faChevronLeft, faChevronRight, faDownload } from '@fortawesome/free-solid-svg-icons'

import Lightbox from "react-image-lightbox";
import "react-image-lightbox/style.css";

import camera from '../img/icons/camera.svg'
import lens from '../img/icons/lens.svg'
import focal from '../img/icons/focal-length.svg'
import apature from '../img/icons/apature.svg'
import timer from '../img/icons/timer.svg'
import iso from '../img/icons/iso.svg'
import albumSVG from '../img/icons/albums.svg'
import { config, searchTree } from "../store";
import {Map} from '../components/Map'
import './photo.css'
// import { LazyLoadImage } from 'react-lazy-load-image-component';
// import 'react-lazy-load-image-component/src/effects/blur.css';
import Image from "../components/Image"

const Photo = ({ collections, photos, match }) => {
  const [isOpen, open] = useState(false)
  const [photo, setPhoto] = useState({
    exif:{
      GPS:{
        latitude:0,
        longitude:0
      }
    }
  })
  const [album, setAlbum] = useState({})
  const [pre_index, setPreIndex] = useState("")
  const [post_index, setPostIndex] = useState("")
  
  useEffect(()=>{
    const id = match.params.id
    setPhoto(photos.filter(c => c.id === id)[0] || { exif: {GPS:{latitude:0}} });
    if (collections !== undefined && photo.album !== undefined) {
      setAlbum(searchTree(collections, photo.album ));
      
      const photoList = photos.filter(c => c.album === album.id) || [];
      let index = photoList.findIndex(x => x.id === id);
      if (index -1 >= 0){
        setPreIndex(photoList[index-1].id)
      } else {
        setPreIndex("")
      }
      if (index +1 <= photoList.length -1){
        setPostIndex(photoList[index+1].id)
      }  else {
        setPostIndex("")
      }

      console.log(album.id, photo.album)
      console.log(album.id, photo.album)
    }

  },[match, album.id, collections, photos, photo.album])

  
    
   

    
    
    const isLocation = () => {
      if (photo.exif.GPS.latitude === 0) {
          if (album.GPS === undefined){
            return false;
          }
          if (album.GPS.latitude === 0){
            return false
          }
      }
      return true;
    }
    const getLong = () => {
      if (photo.exif.GPS.longitude === 0) {
        if (album.GPS !== undefined){
          return album.GPS.longitude;
        }
      }
      return photo.exif.GPS.longitude
    }
    const getLat = () => {
      if (photo.exif.GPS.latitude === 0) {
        if (album.GPS !== undefined){
          return album.GPS.latitude;
        }
      }
      return photo.exif.GPS.latitude
    }

    return (
      <main>
        <div>
        {isOpen && (
          <Lightbox
            mainSrc={config.imageUrl+ photo.id+"?size=original"}
            onCloseRequest={() => open(false)}
            toolbarButtons={[
              <a 
                style={{"textDecoration": "none", paddingRight: "10px", "color": "#AAAAAA"}} 
                href={config.imageUrl+ photo.id+"?size=original"} 
                target="_blank" 
                rel="noopener noreferrer" 
                download={photo.name} 
                className="ril__toolbarItemChild ril__builtinButton">
                <FontAwesomeIcon icon={faDownload} />
              </a>
            ]}
          />
        )}
        
          <div 
            id="gallery_single" 
            className="img-container" 
            onClick={() => open(true)}
          >

          {/* <LazyLoadImage
            placeholderSrc={config.imageUrl+ photo.id}
            src={config.imageUrl+ photo.id+"?size=original"}
            alt={photo.name}
            width={"100%"} 
            height={"100%"}
            effect="blur"
            //wrapperProps={{style:{"objectFit": "contain"}}}
            /> */}

            {/* <LazyImage src={config.imageUrl+ photo.id+"?size=original"} alt={photo.name} style={{ width: "100%", height: "100%", "objectFit": "contain"}} /> */}
            {photo.id !== undefined ?
            <Image
              style={{ width: "100%", height: "100%", objectFit: "contain" }}
              placeholderImg={config.imageUrl+ photo.id}
              src={config.imageUrl+ photo.id}
              alt={photo.name}
            />
            :null}
          </div>
         
          <nav className="navbar navbar-expand-md navbar-dark bg-dark">
            <div className="">
              <ul className="navbar-nav mr-auto">
              { pre_index !== "" &&
                <li className="nav-item">
                <Link to={"/photo/"+pre_index} className="nav-link">
                  <FontAwesomeIcon icon={faChevronLeft} />
                </Link>
              </li>
              }
              </ul>
            </div>
            <ul className="navbar-nav mr-auto photo-title">
              <h2 className="robotFont">{photo.name}</h2>
            </ul>
            <div className="" style={{marginLeft:"auto"}}>
              <ul className="navbar-nav ml-auto">
              { post_index !== "" &&
                <li className="nav-item">
                <Link to={"/photo/"+post_index} className="nav-link">
                  <FontAwesomeIcon icon={faChevronRight} />
                </Link>
              </li>
              }
              </ul>
            </div>
          </nav>

        </div>
        <div className="container" style={{ "backgroundColor": "white", width:"100%", overflow:"auto", padding:"20px 50px"}}>
          <div className="row">
            <div className={ isLocation() ? "col-7" : "col-12"} >
              <table className="table photo-table" style={{ "textAlign": "center", "lineHeight": "50px" }}>
                <tbody>

                  
                  {photo.exif.camera !== "" ?
                    <tr>
                      <th scope="row"><img src={camera} width="50px" alt="camera icon" /></th>
                      <td>{photo.exif.camera}</td>
                    </tr>
                  : null}
                  {photo.exif.LensModel !== "" ?
                    <tr>
                      <th scope="row"><img src={lens} width="50px" alt="lens icon" /></th>
                      <td>{photo.exif.LensModel}</td>
                    </tr>
                  : null}
                  {photo.exif.focal_length !== "" ?
                    <tr>
                      <th scope="row"><img src={focal} width="50px" alt="focal icon"/></th>
                      <td>{photo.exif.focal_length}mm</td>
                    </tr>
                  : null}
                  {photo.exif.f_stop !== "" ?
                    <tr>
                      <th scope="row"><img src={apature} width="50px" alt="apature icon"/></th>
                      <td>{photo.exif.f_stop}</td>
                    </tr>
                  : null}
                  {photo.exif.shutter_speed !== "" ?
                    <tr>
                      <th scope="row"><img src={timer} width="50px" alt="timer icon"/></th>
                      <td>{photo.exif.shutter_speed}</td>
                    </tr>
                  : null}
                  {photo.exif.iso !== "" ?
                    <tr>
                      <th scope="row"><img src={iso} width="50px" alt="iso icon"/></th>
                      <td>{photo.exif.iso}</td>
                    </tr>
                  : null}
                  {album.name !== "" ?
                    <tr>
                      <th scope="row">
                        <img src={albumSVG} width="50px" alt="album icon"/>
                      </th>
                      <td><Link to={"/album/" + album.id}>{album.name}</Link></td>
                    </tr>
                  : null}
                  {photo.format_time !== "" ?
                    <tr>
                      <th scope="row">Date Taken</th>
                      <td><span className="badge badge-pill badge-light date-pill">{photo.format_time}</span></td>
                    </tr>
                  : null}
                </tbody>
              </table>
            </div>
            {isLocation()  ?
              <div className="col-5">
                <Map lng={getLong()} lat={getLat()}/>
              </div>
            : null}
          </div>
        </div>

      </main>
    );
  }


const mapToProps = (state) => {
  const photos = state.PhotosReducer.photos;
  const collections = state.CollectionsReducer.collections;
  return {
    photos,
    collections
  };
}
export default connect(mapToProps)(Photo)
