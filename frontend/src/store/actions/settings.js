
import axios from 'axios';
import {config} from '../index';
import {getOptions, notify} from './index';
export const settingsActions = {
    stats,
    all,
    setProfile,
    setGallery,
    setDeploy
};

function all(){
    return dispatch => {
        dispatch(stats())
        dispatch(gallery())
        dispatch(profile())
        dispatch(deploy())
    }
}
function stats(){
    return dispatch => {
        axios.get(config.baseUrl+"/settings/stats",getOptions()).then((response)=>{
           dispatch(statsUpdated(response.data))
        }).catch((err)=>{
            notify("warning", "Error from server: "+err)
        })
    };
}
function profile(){
    return dispatch => {
        axios.get(config.baseUrl+"/settings/profile",getOptions()).then((response)=>{
           dispatch(profileUpdated(response.data))
        }).catch((err)=>{
            notify("warning", "Error from server: "+err)
        })
    };
}
function gallery(){
    return dispatch => {
        axios.get(config.baseUrl+"/settings/gallery",getOptions()).then((response)=>{
           dispatch(galleryUpdated(response.data))
        }).catch((err)=>{
            notify("warning", "Error from server: "+err)
        })
    };
}
function deploy(){
    return dispatch => {
        axios.get(config.baseUrl+"/settings/deploy",getOptions()).then((response)=>{
           dispatch(deployUpdated(response.data))
        }).catch((err)=>{
            notify("warning", "Error from server: "+err)
        })
    };
}

function setProfile(profile){
    return dispatch => {
        axios.post(config.baseUrl+"/settings/profile", profile ,getOptions()).then((response)=>{
           dispatch(profileUpdated(response.data))
           notify("success", "Profile edited successfully")
        }).catch((err)=>{
            notify("warning", "Error from server: "+err)
        })
    };
}
function setGallery(gallery){
    return dispatch => {
        axios.post(config.baseUrl+"/settings/gallery", gallery, getOptions()).then((response)=>{
           dispatch(galleryUpdated(response.data))
           notify("success", "Gallery edited successfully")
        }).catch((err)=>{
            notify("warning", "Error from server: "+err)
        })
    };
}

function setDeploy(deploy){
    return dispatch => {
        axios.post(config.baseUrl+"/settings/deploy", deploy, getOptions()).then((response)=>{
           dispatch(deployUpdated(response.data))
           notify("success", "Deploy edited successfully")
        }).catch((err)=>{
            notify("warning", "Error from server: "+err)
        })
    };
}


function statsUpdated(stats){
    return{
        type: "STATS_UPDATED",
        stats: stats
    }
}
function profileUpdated(profile){
    return{
        type: "PROFILE_UPDATED",
        profile: profile
    }
}
function galleryUpdated(gallery){
    return{
        type: "GALLERY_UPDATED",
        gallery: gallery
    }
}
function deployUpdated(deploy){
    return{
        type: "DEPLOY_UPDATED",
        deploy: deploy
    }
}

