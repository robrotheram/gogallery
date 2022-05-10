import React, { useEffect, useState } from 'react';
import Map, {Marker} from 'react-map-gl';
import maplibregl from 'maplibre-gl';
import 'maplibre-gl/dist/maplibre-gl.css';


const MapView = (props) => {

  const [lng, setLng] = useState(-1.5);
  const [lat, setLat] = useState(52.5);
  const [zoom, setZoom] = useState(8);

  const [marker_lng, setMarkerLng] = useState(-1.5);
  const [marker_lat, setMarkerLat] = useState(52.5);

  useEffect(() => {
    if (props.lng !== undefined){
      setLng(props.lng);
      setMarkerLng(props.lng);
    }
    if (props.lat !== undefined){
      setLat(props.lat);
      setMarkerLat(props.lat)
    }
    if ((parseInt(props.lat)|| 0) === 0 && (parseInt(props.lng)|| 0) === 0){
      console.log("set zoom")
      setZoom(1)
    }
  }, [props.lng, props.lat]);

  const onMarkerDragEnd = (event) => {
    console.log(event);
    props.onLocation(event.lngLat.lat, event.lngLat.lng)
  };
  return <Map
    initialViewState={{
      width: "100%",
      height: "100%",
      latitude: lat,
      longitude: lng,
      zoom: zoom
    }}
    mapLib={maplibregl}
    mapStyle='https://api.maptiler.com/maps/topo/style.json?key=TcB1jNiCmLlNPGRfj3mM'
    onMouseDown={(event)=>{
      setMarkerLat(event.lngLat.lat)
      setMarkerLng(event.lngLat.lng)
      console.log(event);
      props.onLocation(event.lngLat.lat, event.lngLat.lng)
    }}
  >
  <Marker
          longitude={marker_lng}
          latitude={marker_lat}
          anchor="center"
          draggable
          onDragEnd={onMarkerDragEnd}
        />
</Map>
;
};

export default MapView;
