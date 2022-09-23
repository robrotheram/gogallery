import React, {useState} from 'react';
import { Modal,  Input, Button} from 'antd';
import Map from './Map'
import "./Map.css"

const defaultLat = 52.56815737826566
const defaultLng =-1.4654394633258416

const InputGroup = Input.Group;
export const LocationModal = (props) =>{
    const [visable, setVisable] = useState(false);

    const [lng, setLng] = useState(defaultLng);
    const [lat, setLat] = useState(defaultLat);

    React.useEffect(() => {
        if (props.lat !== undefined){
            setLat(props.lat);
        }
      }, [props.lat]);

    React.useEffect(() => {
        if (props.lng !== undefined){
            setLng(props.lng);
        }
    }, [props.lng]);


    const onUpdate = () => {
       props.onUpdate(lat,lng);
       setVisable(false)
    }
    const onCancel = () => {setVisable(false)}

    const onLocationChange = ({lat, lng}) => {
        setLng(lng)
        setLat(lat)
    }

    return (
        <div style={{width:"100%"}}>
        <InputGroup compact style={{width:"100%"}} onClick={()=> setVisable(true)}>
          <Input style={{ width: '50%' }} value={lat.toFixed(6)} onClick={()=> setVisable(true)}/>
          <Input style={{ width: '50%' }} value={lng.toFixed(6)} onClick={()=> setVisable(true)} />
        </InputGroup>
        <Modal
          open={visable}
          title="Update Location"
          okText="Update"
          onCancel={onCancel}
          onOk={onUpdate}
          className="map-model"
        >
          
           <div style={{width:"100%", height:"400px"}}> <Map lat={lat} lng={lng} onLocation={onLocationChange}/></div>
           <InputGroup compact style={{width:"100%", padding:"12px"}}>
                <Input style={{ width: '35%' }} value={lat.toFixed(6)}/>
                <Input style={{ width: '35%' }} value={lng.toFixed(6)}/>
                <Button type="danger" style={{ width: '30%' }} onClick={() => onLocationChange({lat:0, lng:0})}>Clear Location</Button>
            </InputGroup>
        </Modal>
        </div>
    )
} 
export {default as Map} from "./Map"