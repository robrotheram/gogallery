import React from 'react';
import { LazyLoadImage, trackWindowScroll }
    from 'react-lazy-load-image-component';
import { Row, Col } from 'antd';
import "./loading/loading.css"
import Loading from './loading/loading.svg';
import { config } from '../store';
const Gallery = ({ images, imageSize, addElementRef, selectPhoto, getStyle, scrollPosition }) => {
    
    const spanWidth = (size) => {
        console.log("CHANGE_SIZE", size)
        switch (size) {
            case "xsmall":
                return 2;
            case "small":
                return 4;
            case "medium":
                return 6;
            case "large":
                return 12;
            case "xlarge":
                return 24;
            default: return 2;
        }
    }
    return (
        <Row gutter={[16, 16]}>
            {images.map((image, index) =>
                <Col key={image.id} span={spanWidth(imageSize)}>
                    <div
                        //ref={addElementRef}
                        className={`item`}
                    >
                        <figure className="galleryImg" style={getStyle(image.id)} onClick={(e) => selectPhoto(e, image)}>
                            <LazyLoadImage
                                alt={image.name}
                                effect="blur"
                                scrollPosition={scrollPosition}
                                src={config.imageUrl + image.id + "?size=" + imageSize}
                                wrapperProps={
                                    {
                                        style:{
                                            backgroundSize:"30% 30%",
                                            backgroundRepeat:"no-repeat",
                                            backgroundPosition:"center",
                                            backgroundImage:`url(${Loading})`,
                                            color: "transparent",
                                            display: "inline-block",
                                            height: "100%",
                                            width: "100%",
                                            border: "1px solid white",
                                            borderRadius: "0px"
                                        }
                                    }
                                }
                                width="100%"
                                height="100%"
                                style={{objectFit:"cover"}}
                            />
                        </figure>
                    </div>
                </Col>
            )}
        </Row>
    );
}
// Wrap Gallery with trackWindowScroll HOC so it receives
// a scrollPosition prop to pass down to the images
export default trackWindowScroll(Gallery);

