import React from "react";
import { LazyLoadImage } from 'react-lazy-load-image-component';
import 'react-lazy-load-image-component/src/effects/blur.css';

const LazyImage = ({ src, alt }) => {
  return (
    <LazyLoadImage
      alt={alt}
      height={"100%"}
      effect="blur"
      src={src} // use normal <img> attributes as props
      width={"100%"} />
  );
};
export default LazyImage;
