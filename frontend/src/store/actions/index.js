import { notification } from 'antd';

export * from './user';
export * from './photos';
export * from './collections';
export * from './gallery';
export * from './settings';
export * from './tasks';

export function getOptions(){
  if(localStorage.getItem('token')){
      return{headers: {Authorization:localStorage.getItem('token')}}
  }
}

export function notify(type, description){
  let message = ""
  switch(type){
    case "warning": message = "Oh dear something went wong!"; break
    default: message = "Completed successfully" ; break
  }

  notification[type]({
    message: message,
    description: description,
    duration: 5,
  });
}