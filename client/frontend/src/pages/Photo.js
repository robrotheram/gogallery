import React from "react";
import { connect } from 'react-redux';
import { Link } from "react-router-dom";
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faChevronLeft, faChevronRight } from '@fortawesome/free-solid-svg-icons'

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

import './photo.css'
import { LazyImage } from "../components/Lazyloading";

class Photo extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      isOpen: false
    };
  }

  render() {
    let { collections, photos } = this.props
    const id = this.props.match.params.id
    const photo = photos.filter(c => c.id === id)[0] || { exif: {} };
    
    let pre_index = ""
    let post_index = ""

    let album_id = ""
    let album = {}
    if (collections !== undefined && photo.album !== undefined) {
      album = searchTree(collections, photo.album )
      album_id = album.id

      const photoList = photos.filter(c => c.album === album.id) || [];
      let index = photoList.findIndex(x => x.id === id);
      if (index -1 >= 0){
        pre_index = photoList[index-1].id
      } 
      if (index +1 <= photoList.length -1){
        post_index = photoList[index+1].id
      } 

      console.log(album_id, photo.album)
      console.log(album_id, photo.album)
    }

    
    const { isOpen } = this.state;
    console.log()
    return (
      <main>
        <div>
        {isOpen && (
          <Lightbox
            mainSrc={config.imageUrl+ photo.id+"?size=original"}
            onCloseRequest={() => this.setState({ isOpen: false })}
          />
        )}
          <div id="gallery_single" className="img-container" onClick={() => this.setState({ isOpen: true })}>
            <LazyImage src={config.imageUrl+ photo.id} alt={photo.name} />
            <div className="downloadLink"><a href={config.imageUrl+ photo.id+"?size=original"} target="_blank" rel="noopener noreferrer" download={photo.name}>Download Orginal</a></div>
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
        <div className="container" style={{ "backgroundColor": "white", "marginTop": "20px" }}>
          <div className="row">
            <div className="col-12">
              <h2 className="robotFont">{photo.name}  <span className="badge badge-pill badge-light date-pill">{photo.format_time}</span> </h2>

              <table className="table" style={{ "textAlign": "center", "lineHeight": "50px" }}>
                <tbody>

                  <tr>
                    <th scope="row"><img src={camera} width="50px" alt="camera icon" /></th>
                    <td>{photo.exif.camera}</td>
                  </tr>

                  <tr>
                    <th scope="row"><img src={lens} width="50px" alt="lens icon" /></th>
                    <td>{photo.exif.LensModel}</td>
                  </tr>

                  <tr>
                    <th scope="row"><img src={focal} width="50px" alt="focal icon"/></th>
                    <td>{photo.exif.focal_length}mm</td>
                  </tr>

                  <tr>
                    <th scope="row"><img src={apature} width="50px" alt="apature icon"/></th>
                    <td>{photo.exif.f_stop}</td>
                  </tr>

                  <tr>
                    <th scope="row"><img src={timer} width="50px" alt="timer icon"/></th>
                    <td>{photo.exif.shutter_speed}</td>
                  </tr>

                  <tr>
                    <th scope="row"><img src={iso} width="50px" alt="iso icon"/></th>
                    <td>{photo.exif.iso}</td>
                  </tr>

                  <tr>
                    <th scope="row">
                      <img src={albumSVG} width="50px" alt="album icon"/>
                    </th>
                    <td><Link to={"/album/" + album_id}>{album.name}</Link></td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

      </main>
    );
  }
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
