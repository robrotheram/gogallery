import React from "react";
import { connect } from 'react-redux';
import {config} from '../store'
import {Link} from "react-router-dom";
import placeholder from "../img/placeholder.png"
import { fuzzySearch } from "../components/Search/utils";
import { LazyImage } from "../components/Lazyloading";

//<LazyImage src={config.imageUrl+ photo.id} alt={photo.name} />
class AlbumsPage extends React.PureComponent {

  render () {
    console.log(this.props.collections)
  return (
    <main>
        <div className="py-5 bg-light" style={{height:"calc(100vh - 60px)"}}>
            <div className="container">
                <div className="row">
                  {this.props.collections.map((value, index) => {
                    return (
                      <div className="col-md-4" key={value.id}>
                            <div className="card mb-4 shadow-sm">
                                <Link to={"album/"+value.id}>
                                {
                                  value.profile_image === ""
                                  ? <img src={placeholder} alt={value.name} width="100%" height="250px" style={{"objectFit": "cover"}}/>
                                  : <LazyImage src={config.imageUrl+value.profile_image} alt={value.name} style={{ width: "100%", height: "250px", "objectFit": "cover"}}/>
                                }
                                </Link>
                                <div className="card-body">
                                    <p className="card-text text-center">{value.name}</p>
                                </div>
                            </div>
                        </div>
                    )
                  })}
                </div>
            </div>
        </div>
    </main>
  );
                }
}
const mapToProps = (state) =>{
  //const collections = state.CollectionsReducer.collections;
  var collections = Object.keys(state.CollectionsReducer.collections).map(function(key) {
    return state.CollectionsReducer.collections[key]
  });
  
  let loc = state.router.location.pathname.split("/")
  let searchTerm = state.search.search[loc[1]] 
  if (searchTerm !== "" && searchTerm !== undefined){
    return {
      collections: fuzzySearch(["name"], collections, searchTerm )
    }
  }



  return {
    collections,
  };
}
export default connect(mapToProps)(AlbumsPage)

