
import axios from 'axios';
import {getOptions, notify} from './index';
import {config} from '../index';

export const photoActions = {
    getAll,
    edit
};

function getAll(){
    return dispatch => {
        dispatch(photoUpdating());
        axios.get(config.baseUrl+"/photos", getOptions()).then((response)=>{
            console.log(response.data);
            dispatch(setPhotoDetails(response.data));
        }).catch((err)=>{
            if (err.includes("401") ){
                window.location.href = '/dashboard/login';
            };
            notify("warning", "Error from server: "+err)
        })
    }

}

function edit(photo){
    return dispatch => {
        dispatch(photoUpdating());
        axios.post(config.baseUrl+"/photo/"+photo.id, photo, getOptions()).then((response)=>{
            notify("success", "Photo edited successfully")
            dispatch(getAll(response.data));
        }).catch((err)=>{
            notify("warning", "Error from server: "+err)
        })
    }
}

function photoUpdating(){
    return{
        type: "PHOTO_FETCHING"
    }
}

function setPhotoDetails(photos){
    return{
        type: "PHOTO_RECEIVED",
        photos: photos,
    }
}

