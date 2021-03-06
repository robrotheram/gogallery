
import axios from 'axios';
import { history, config} from '../index';
import {getOptions, notify} from './index';
export const userActions = {
    login,
    reauth,
    logout,
    update
};

function login(username, password){
    console.log('login: ', username);
    return dispatch => {
        let payload = {
            username: username,
            password: password
        }
        console.log('dispathc: ', username);
        axios.post(config.baseUrl+"/login", payload).then((response)=>{
            console.log(response.data);
            if (response.data.token) {
                localStorage.setItem('token', response.data.token);
                localStorage.setItem('email', response.data.email);
                localStorage.setItem('username', response.data.username);
                dispatch(setUserDetails(response.data));
                console.log("push history ");
                history.push('/');
            }else{
                dispatch(logout())
            }
        }).catch((err)=>{
            dispatch(logoutFailedUser())
        })
    };
}
function update(user){
    return dispatch => {
        axios.post(config.baseUrl+"/auth/update", user, getOptions()).then((response)=>{
            localStorage.setItem('email', response.data.email);
            localStorage.setItem('username', response.data.username);
            dispatch(setUserDetails(response.data));
            notify("success", "User details edited successfully")
            
        }).catch((err)=>{
            console.log("Error in response");
            console.log(err);
        })
    };
}


function reauth(){
    return dispatch => {
        axios.get(config.baseUrl+"/authorised",getOptions()).then((response)=>{
            if (response.data.token) {
                localStorage.setItem('token', response.data.token);
                localStorage.setItem('email', response.data.email);
                localStorage.setItem('username', response.data.username);
                
            }
        }).catch((err)=>{
            console.log("Error in response");
            console.log(err);
            dispatch(logout());
        })
    };
}

function logout(){
    return dispatch => {
        localStorage.removeItem('email');
        localStorage.removeItem('token');
        localStorage.removeItem('username');
        dispatch(logoutUser());
        history.push('/');
    }
}

export function setUserDetails(user){
    return{
        type: "LOGIN_SUCCESS",
        email: user.email,
        username: user.username,
        token: user.token
    }
}

export function logoutFailedUser(){
    return{
        type: "LOGOUT_FAILED",
    }
}
export function logoutUser(){
    return{
        type: "LOGOUT_SUCCESS",
        auth: false,
        email: '',
        token: ''
    }
}