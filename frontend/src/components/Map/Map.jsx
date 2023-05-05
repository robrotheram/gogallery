import React, { useEffect, useState } from 'react';
import { useRef } from 'react';
import { MapContainer, Marker, Popup, TileLayer, useMap, useMapEvents } from 'react-leaflet'

const  LocationMarker = ({onLocation, center}) => {
  const [position, setPosition] = useState(center)
  const map = useMapEvents({
    click(e) {
      console.log("Map click evebt", e.latlng)
      setPosition(e.latlng)
      onLocation(e.latlng)
    }
  })

  return position === null || (position.lat === 0 && position.lng === 0) ? null : (
    <Marker position={position}>
      <Popup>You are here</Popup>
    </Marker>
  )
}

const ComponentResize = () => {
  const map = useMap()

  setTimeout(() => {
    console.log("MAP CALLED")
      map.invalidateSize()
  }, 0)

  return null
}

const MapView = (props) => {
  let center = [props.lat, props.lng]
  if (props.lng === 0 && props.lat === 0) {
    center = [51.505, -0.09]
  }
  return (
    <MapContainer style={{height:"100%", width:"100%"}} center={center} zoom={13} scrollWheelZoom={true}>
      <ComponentResize/>
      <TileLayer
        attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
        url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
      />
     <LocationMarker onLocation={props.onLocation} center={{lat:props.lat, lng:props.lng}} />
    </MapContainer>
  )
};

export default MapView;
