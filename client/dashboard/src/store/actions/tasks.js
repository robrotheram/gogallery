
import axios from 'axios';
import { history, config} from '../index';
import {getOptions, notify} from './index';

export const taskActions = {
    purge,
    clear,
    backup
};

function download(content, fileName, contentType) {
    var a = document.createElement("a");
    var file = new Blob([JSON.stringify(content,null,2)], {type: contentType});
    a.href = URL.createObjectURL(file);
    a.download = fileName;
    a.click();
}

function purge(){
    return dispatch => {
        axios.get(config.baseUrl+"/tasks/purge",getOptions()).then((response)=>{
            notify("success", "Rescan task started")
        }).catch((err)=>{
            notify("warning", "Error from server: "+err)
        })
    };
}
function clear(){
    return dispatch => {
        axios.get(config.baseUrl+"/tasks/clear",getOptions()).then((response)=>{
            notify("success", "Clear task started")
        }).catch((err)=>{
            notify("warning", "Error from server: "+err)
        })
    };
}

function backup(){
    return dispatch => {
        axios.get(config.baseUrl+"/tasks/backup",getOptions()).then((response)=>{
            download(response.data, 'galleryBackup.txt', 'text/plain');
        }).catch((err)=>{
            notify("warning", "Error from server: "+err)
        })
    };
}