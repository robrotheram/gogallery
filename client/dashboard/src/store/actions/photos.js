
import axios from 'axios';
import {getOptions, notify} from './index';
import {config} from '../index';
import { logout } from './user';

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
        }).catch((error)=>{
            console.log(error.response)
            if (error.response.status === 401 ){
                dispatch(logout());
            };
            notify("warning", "Error from server: "+error)
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
            if (err.response.status === 401 ){
                dispatch(logout());
            };
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

