import React from 'react';
import { LazyLoadImage, trackWindowScroll }
    from 'react-lazy-load-image-component';
import { config } from '../store';
import { Row, Col } from 'antd';
const Gallery = ({ images, imageSize,addElementRef,selectPhoto,getStyle, scrollPosition }) => (
    <Row gutter={[16, 16]}>
        {images.map((image, index) =>
            <Col key={image.id} span={parseInt(imageSize)}>
                <div
                    //ref={addElementRef}
                    className={`item`}
                >
                    <figure className="galleryImg" style={getStyle(image.id)} onClick={(e) => selectPhoto(e, image)}>
                        <LazyLoadImage
                            alt={image.name}
                            // Make sure to pass down the scrollPosition,
                            // this will be used by the component to know
                            // whether it must track the scroll position or not
                            scrollPosition={scrollPosition}
                            src={config.imageUrl + image.id + "?size=tiny&token=" + localStorage.getItem('token')}
                            width="100%"
                            height="100%"
                        />
                    </figure>
                </div>
            </Col>
        )}
    </Row>
);
// Wrap Gallery with trackWindowScroll HOC so it receives
// a scrollPosition prop to pass down to the images
export default trackWindowScroll(Gallery);

