import React, { useRef, useEffect, useState } from 'react';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css'
import './Map.css';

mapboxgl.accessToken =
  'pk.eyJ1IjoibWFwYm94IiwiYSI6ImNpejY4M29iazA2Z2gycXA4N2pmbDZmangifQ.-g_vE53SD2WrJ6tFX7QHmA';

const Map = (props) => {
  const mapContainerRef = useRef(null);

  const [lng, setLng] = useState(-1.5);
  const [lat, setLat] = useState(52.5);
  const [zoom, setZoom] = useState(8);

  useEffect(() => {
    if (props.lng !== undefined){
      setLng(props.lng);
    }
    if (props.lat !== undefined){
      setLat(props.lat);
    }
    if ((parseInt(props.lat)|| 0) === 0 && (parseInt(props.lng)|| 0) === 0){
      console.log("set zoom")
      setZoom(1)
    }
  }, [props.lng, props.lat]);
  

  // Initialize map when component mounts
  useEffect(() => {
    const map = new mapboxgl.Map({
      container: mapContainerRef.current,
      style: 'mapbox://styles/mapbox/streets-v11',
      center: [lng, lat],
      zoom: zoom
    });

    // Add navigation control (the +/- zoom buttons)
    map.addControl(new mapboxgl.NavigationControl(), 'top-right');

    map.on('move', () => {
      setLng(map.getCenter().lng.toFixed(4));
      setLat(map.getCenter().lat.toFixed(4));
      setZoom(map.getZoom().toFixed(2));
    });
    let marker = new mapboxgl.Marker({draggable: true})
    .setLngLat([lng, lat])
    .addTo(map);

    marker.on('dragend', () => {
      var lngLat = marker.getLngLat();
      props.onLocation(lngLat.lat, lngLat.lng)
    });
    map.on('click', function(e){
      marker.setLngLat([e.lngLat.lng, e.lngLat.lat])
      props.onLocation(e.lngLat.lat, e.lngLat.lng)
    })

    // Clean up on unmount
    return () => map.remove();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  return (
      <div className='map-container card' ref={mapContainerRef} />
  );
};

export default Map;
