
import axios from 'axios';
import { history, config} from '../index';

export const galleryActions = {
    showAdd,
    hideAdd,
    showUpload,
    hideUpload,
    changeImageSize
};

function showAdd(){
    return{type: "SHOW_ADD_MODAL"}
}

function showUpload(){
    return{type: "SHOW_UPLOAD_MODAL"}
}

function hideAdd(){
    return{type: "HIDE_ADD_MODAL"}
}

function hideUpload(){
    return{type: "HIDE_UPLOAD_MODAL"}
}

function changeImageSize(size){
    const type = "CHANGE_IMAGE_SIZE"
    return{type: type, size: size}
        
   
}
