import React from "react";
import { connect } from 'react-redux';
import { Link } from "react-router-dom";
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faChevronLeft, faChevronRight } from '@fortawesome/free-solid-svg-icons'

import camera from '../img/icons/camera.svg'
import lens from '../img/icons/lens.svg'
import focal from '../img/icons/focal-length.svg'
import apature from '../img/icons/apature.svg'
import timer from '../img/icons/timer.svg'
import iso from '../img/icons/iso.svg'
import album from '../img/icons/albums.svg'
import { config } from "../store";

import './photo.css'

class Photo extends React.Component {

  render() {
    let { collections, photos } = this.props
    const id = this.props.match.params.id
    const photo = photos.filter(c => c.id === id)[0] || { exif: {} };
    
    let pre_index = ""
    let post_index = ""

    let album_id = ""
    if (collections !== undefined && photo.album !== undefined && collections.length > 0) {
      let album = (collections.filter(c => c.name === photo.album)[0])
      album_id = album.id

      const photoList = photos.filter(c => c.album === album.name) || [];
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
    console.log(album_id)

    return (
      <main>
        <div>
          <div id="gallery_single" className="img-container">
            <img src={config.baseUrl + "/img/" + photo.id} alt={photo.name} />
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
        <div className="container" style={{ "background-color": "white", "margin-top": "20px" }}>
          <div className="row">
            <div className="col-12">
              <h2 className="robotFont">{photo.name}  <span className="badge badge-pill badge-light date-pill">{photo.format_time}</span> </h2>

              <table className="table" style={{ "text-align": "center", "line-height": "50px" }}>
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
                      <img src={album} width="50px" alt="album icon"/>
                    </th>
                    <td><Link to={"/album/" + album_id}>{photo.album}</Link></td>
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
