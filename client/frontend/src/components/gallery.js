


import React from 'react';
import {LazyImage} from './Lazyloading'
import './gallery.css'
import {
    Link,
  } from "react-router-dom";

import {config} from '../store'
function Gallery({ images}) {
    let items = []
    images = images || []
    console.log(images)
    images.map((x,i) =>  items.push(
        <figure key={x.id} className="masonry-brick masonry-brick--h">
            <Link to={"/photo/"+x.id} >
                <LazyImage src={config.imageUrl+x.id} className="masonry-LazyImage" alt={x.name} />
            </Link>
        </figure>
    ));
    return (
        <div className="masonry masonry--h">
        {items}
      </div>
    )
}
export default Gallery;
