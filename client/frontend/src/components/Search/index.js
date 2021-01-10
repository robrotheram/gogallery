import React, {useEffect, useState} from 'react';
import './search.css'

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faSearch } from '@fortawesome/free-solid-svg-icons'
import { connect } from 'react-redux';
import {galleryActions} from '../../store/actions'

const SearchBar = (props) => {

    const [searchTerm, setSeachTerm] = useState("");

    const handleChange = (e) => {
        let loc = props.location.location.pathname.split("/")
        let obj = {}
        obj[loc[1]] = e.target.value
        props.dispatch(galleryActions.setSearch(obj));
    }

    useEffect(() => {
        let loc = props.location.location.pathname.split("/")
        setSeachTerm(props.search[loc[1]])
      }, [props.location, props.search]);
    

    return (
    <div className="searchbar">
        <input className="search_input" value={searchTerm || ""} type="text" name="" placeholder="Search..." onChange={handleChange} />
        <button className="search_icon"><FontAwesomeIcon icon={faSearch}/></button>
    </div>)
}

const mapStateToProps = (state) =>{
    return {
        search: state.search.search,
        location: state.router

    }
}
export default connect(mapStateToProps)(SearchBar)