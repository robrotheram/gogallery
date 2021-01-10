


import React, {useEffect} from 'react';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faUserCircle } from '@fortawesome/free-solid-svg-icons'
import './header.css'
import logo from '../img/logo.png'
import albums from '../img/icons/albums.svg'
import { useSelector } from 'react-redux';
import { connect } from 'react-redux';
import {Link} from "react-router-dom";
import SearchBar from './Search';

function Header(props) {
const state = useSelector(state => state.ConfigReducer.config)

const showSearch = () => {
    let loc = props.location.pathname.split("/")
    if (loc[1] === undefined) {
        return false
    }
    switch (loc[1]){
        case "": return true
        case "albums": return true
        case "album": return true
        default: return false
    }
}

useEffect(() => {
    console.log("HEADER", props.location.pathname.split("/"))
  }, [props.location.pathname]);


console.log("Fired", state)
  return (
  <header className="fixed-top" style={{"zIndex":"999"}}>
        <nav className="navbar navbar-expand-lg navbar-light bg-light shadow">
            <Link to="/" className="navbar-brand mx-auto">
                <img src={logo} width="30px" alt="Gallery Logo" style={{"marginRight": "10px"}}/>
                <strong>{state.name}</strong>
            </Link>
            <input type="checkbox" id="navbar-toggle-cbox"/>
            <label htmlFor="navbar-toggle-cbox" className="navbar-toggler collapsed" type="button" data-toggle="collapse" data-target="#navbarSupportedContent" aria-controls="navbarSupportedContent" aria-expanded="false" aria-label="Toggle navigation">
                <span className="navbar-toggler-icon"></span>
            </label>

            <div className="collapse navbar-collapse" id="navbarSupportedContent">
                <ul className="navbar-nav mr-auto">
                    <li className="nav-item">
                        <Link to="/albums" className="nav-link text-center" >
                            <img src={albums} width="32px" style={{"marginRight":"2px"}} alt="Album icon"/>Albums
                        </Link>
                    </li>
                    {showSearch() ?
                    <li className="nav-item">
                        <SearchBar/>
                    </li>
                    : null}
                </ul>

                    <li className="nav-item active">
                    <Link className="nav-link text-center" to="/about" style={{"color": "#261F1F"}}>
                            <FontAwesomeIcon icon={faUserCircle} size="2x" style={{"color": "#5f5f5f"}}/>
                            <span> Profile</span></Link>
                    </li>


            </div>
        </nav>
    </header>
    )
}

const mapToProps = (state) =>{
    return state.router
}
export default connect(mapToProps)(Header);
