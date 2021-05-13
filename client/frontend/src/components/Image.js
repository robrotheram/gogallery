import React, { useEffect, useState } from "react";
import{FulfillingSquareSpinner} from "react-epic-spinners"
const Image = ({ src, alt, style }) => {
    const [placeholderLoaded, setPlaceholderLoaded] = useState(false);
    const [imgURL, setURL]= useState("")
    const [LargeURL, setLargeURL]= useState("")
    useEffect(() =>{
        console.log("Image Change")
        setPlaceholderLoaded(false)
        setURL(src)
        setLargeURL(src+"?size=original")
    },[src])
    style = {
        ...style,
        opacity:"100%",
        visibility:"visible",
        transition:"visibility 0.3s linear,opacity 0.3s linear",
    }
    return (
      <div style={style}>
        { placeholderLoaded? null : (
            <div style={{
                position: "absolute",
                top: "calc( 50% - 25px)",
                left: "calc( 50% - 25px)",
            }}>
                <FulfillingSquareSpinner />
            </div>
        )}
         <img
            style={placeholderLoaded ? style : {...style, visibility:"hidden", opacity:0 }}
            src={imgURL}
            alt={alt}
            onLoad={() => setPlaceholderLoaded(true)}
            />
        <img
          style={{display:"none"}}
          src={LargeURL}
          alt=""
          onLoad={() => setURL(LargeURL)}
        />
      </div>
    );
  };

export default Image