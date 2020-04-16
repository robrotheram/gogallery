
import axios from 'axios';
import {getOptions, notify} from './index';
import {config} from '../index';

import {photoActions} from './photos'
export const collectionActions = {
    getAll,
    create,
    move,
    remove,
    upload
};

function getAll(){
    return dispatch => {
        dispatch(collectionUpdating());
        axios.get(config.baseUrl+"/collections", getOptions()).then((response)=>{
            console.log(response.data);
            dispatch(setPhotoDetails(response.data));
        }).catch((err)=>{
            notify("warning", "Error from server: "+err)
        })
    }
};

function create (collection) {
    return dispatch => {
        axios.post(config.baseUrl + '/collection',  collection, getOptions()).then(result => {
            dispatch(getAll());
            notify("success", "Collections created successfully")
        }).catch((err)=>{
            notify("warning", "Error from server: "+err)
        })
    }
}

function upload(collection) {
    return dispatch => {
        axios.post(config.baseUrl + '/collection/upload',  collection, getOptions()).then(result => {
            dispatch(photoActions.getAll());
            notify("success", "Collections uploaded successfully")
        }).catch((err)=>{
            notify("warning", "Error from server: "+err)
        })
    }
}

function move (collection){
    return dispatch => {
        axios.post(config.baseUrl + '/collection/move',  collection, getOptions()).then(result => {
            dispatch(photoActions.getAll());
            notify("success", "Photo deleted successfully")
        }).catch((err)=>{
            notify("warning", "Error from server: "+err)
        })
    }
}

function remove (photoID){
    return dispatch => {
        axios.delete(config.baseUrl + '/photo/'+photoID, getOptions()).then(result => {
            dispatch(photoActions.getAll());
            notify("success", "Collections moved successfully")
        }).catch((err)=>{
            notify("warning", "Error from server: "+err)
        })
    }
}


function collectionUpdating(){
    return{
        type: "COLLECTION_FETCHING"
    }
}

function setPhotoDetails(data){
    return{
        type: "COLLECTIONS_RECEIVED",
        collections: data.albums,
        dates: [...data.dates],
        uploadDates: [...data.uploadDates]
    }
}

